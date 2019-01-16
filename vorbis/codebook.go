package vorbis

import "math"

type CodebookEntry struct {
  Unused   bool
  Length   int
  Codeword uint32
  Num      int
}

type Codebook struct {
  Dimensions    int
  Entries       []CodebookEntry
  Multiplicands []uint32

  Minimum_value float64
  Delta_value   float64
  Sequence_p    bool

  // Value_vectors[entry][dimension]
  Value_vectors [][]float64
}

func toBin(n uint32, l int) string {
  ret := ""
  for n > 0 {
    if n%2 == 0 {
      ret = "0" + ret
    } else {
      ret = "1" + ret
    }
    n = n >> 1
  }
  for len(ret) < l {
    ret = "0" + ret
  }
  return ret
}

func (book *Codebook) DecodeScalar(br *BitReader) int {
  // TODO: This obviously needs to be seriously optimized
  var word uint32
  for length := 0; length < 32; length++ {
    for i := range book.Entries {
      if book.Entries[i].Unused {
        continue
      }
      if book.Entries[i].Length == length && book.Entries[i].Codeword == word {
        return book.Entries[i].Num
      }
    }
    word = word << 1
    bit := br.ReadBits(1)
    word |= bit
  }
  panic("Codebook failed to decode properly.")
}

func (book *Codebook) DecodeVector(br *BitReader) []float64 {
  index := book.DecodeScalar(br)
  return book.Value_vectors[index]
}

func (book *Codebook) allocateTable() {
  // Build the table out of a single array
  vector := make([]float64, len(book.Entries)*book.Dimensions)
  book.Value_vectors = make([][]float64, book.Dimensions)
  for i := range book.Value_vectors {
    book.Value_vectors[i] = vector[i*len(book.Entries) : (i+1)*len(book.Entries)]
  }
}

func (book *Codebook) BuildVQType1() {
  book.allocateTable()
  for dim := range book.Value_vectors[0] {
    last := 0.0
    index_divisor := 1
    for entry := range book.Value_vectors {
      offset := (dim / index_divisor) % len(book.Multiplicands)
      // TODO: The java implementation takes the absolute value of the Multiplicand here, find out if that is necessary or meaningful
      book.Value_vectors[entry][dim] = float64(book.Multiplicands[offset])*book.Delta_value + book.Minimum_value + last
      if book.Sequence_p {
        last = book.Value_vectors[entry][dim]
      }
      index_divisor *= len(book.Multiplicands)
    }
  }
}
func (book *Codebook) BuildVQType2() {
  book.allocateTable()
  last := 0.0
  for entry := range book.Value_vectors {
    offset := entry * book.Dimensions
    for dim := range book.Value_vectors[entry] {
      // TODO: Same thing with absolute value in the java implementation
      book.Value_vectors[entry][dim] = float64(book.Multiplicands[offset])*book.Delta_value + book.Minimum_value + last
      if book.Sequence_p {
        last = book.Value_vectors[entry][dim]
      }
      offset++
    }
  }
}

func (book *Codebook) AssignCodewords() {
  marker := make([]uint32, 33)
  for i := range book.Entries {
    entry := &book.Entries[i]
    if entry.Unused {
      continue
    }
    if entry.Length == 0 {
      continue
    }
    word := marker[entry.Length]
    if entry.Length < 32 && (word>>uint(entry.Length)) != 0 {
      panic("Codebook contains an overspecified huffman tree.")
    }

    entry.Codeword = word
    for j := entry.Length; j > 0; j-- {
      if marker[j]&1 != 0 {
        if j == 1 {
          marker[1]++
        } else {
          marker[j] = marker[j-1] << 1
        }
        break
      }
      marker[j]++
    }

    for j := entry.Length + 1; j <= 32; j++ {
      if marker[j]>>1 == word {
        word = marker[j]
        marker[j] = marker[j-1] << 1
      } else {
        break
      }
    }
  }
}

func (book *Codebook) decode(br *BitReader) {
  if br.ReadBits(24) != 0x564342 {
    panic("Codebook sync pattern not found")
  }

  book.Dimensions = int(br.ReadBits(16))
  num_entries := int(br.ReadBits(24))
  book.Entries = make([]CodebookEntry, num_entries)
  ordered := br.ReadBits(1) == 1

  // Decode codeword lengths
  if ordered {
    current_entry := 0
    for current_entry < num_entries {
      current_length := int(br.ReadBits(5)) + 1
      number := int(br.ReadBits(ilog(uint32(num_entries - current_entry))))
      for i := 0; i < number; i++ {
        book.Entries[current_entry+i].Length = current_length
        book.Entries[current_entry+i].Num = current_entry + i
      }
      current_length++
      current_entry += number
      if current_entry >= num_entries {
        panic("Error decoding Codebooks")
      }
    }
  } else {
    sparse := br.ReadBits(1) == 1
    if sparse {
      current_entry := 0
      for i := range book.Entries {
        flag := br.ReadBits(1) == 1
        if flag {
          book.Entries[i].Length = int(br.ReadBits(5)) + 1
          book.Entries[i].Num = current_entry
          current_entry++
        } else {
          book.Entries[i].Unused = true
        }
      }
    } else {
      for i := range book.Entries {
        book.Entries[i].Length = int(br.ReadBits(5)) + 1
      }
    }
  }

  // read the vector lookup table
  Codebook_lookup_type := int(br.ReadBits(4))
  switch Codebook_lookup_type {
  case 0:
    // no vector lookup

  case 1:
    fallthrough
  case 2:
    book.Minimum_value = float64(math.Float32frombits(br.ReadBits(32)))
    book.Delta_value = float64(math.Float32frombits(br.ReadBits(32)))
    Codebook_value_bits := int(br.ReadBits(4) + 1)
    book.Sequence_p = br.ReadBits(1) == 1
    var Codebook_lookup_values int
    if Codebook_lookup_type == 1 {
      Codebook_lookup_values = Lookup1Values(len(book.Entries), book.Dimensions)
    } else {
      Codebook_lookup_values = len(book.Entries) * book.Dimensions
    }
    book.Multiplicands = make([]uint32, Codebook_lookup_values)
    for i := range book.Multiplicands {
      book.Multiplicands[i] = br.ReadBits(Codebook_value_bits)
    }

  default:
    panic("Unknown vector lookup method")
  }

  // Assign huffman values
  book.AssignCodewords()

  switch Codebook_lookup_type {
  case 1:
    book.BuildVQType1()
  case 2:
    book.BuildVQType2()
  }
}
