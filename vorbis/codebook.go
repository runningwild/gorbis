package vorbis

import "math"

type CodebookEntry struct {
  Unused   bool
  Length   int
  Codeword uint32
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

func (book *Codebook) Decode(br *BitReader) int {
  // TODO: This obviously needs to be seriously optimized
  var word uint32
  for length := 0; length < 32; length++ {
    for i := range book.Entries {
      if book.Entries[i].Length == length && book.Entries[i].Codeword == word {
        return i
      }
    }
    word = word << 1
    word |= br.ReadBits(1)
  }
  panic("Codebook failed to decode properly.")
}

func (book *Codebook) allocateTable() {
  // Build the table out of a single array
  vector := make([]float64, len(book.Entries) * book.Dimensions)
  book.Value_vectors = make([][]float64, book.Dimensions)
  for i := range book.Value_vectors {
    book.Value_vectors[i] = vector[i * len(book.Entries) : (i + 1) * len(book.Entries)]
  }
}

func (book *Codebook) BuildVQType1() {
  book.allocateTable()
  for entry := range book.Value_vectors {
    last := 0.0
    index_divisor := 1
    for dim := range book.Value_vectors[entry] {
      offset := (entry / index_divisor) % len(book.Multiplicands)
      // TODO: The java implementation takes the absolute value of the Multiplicand here, find out if that is necessary or meaningful
      book.Value_vectors[entry][dim] = float64(book.Multiplicands[offset]) * book.Delta_value + book.Minimum_value + last
      if book.Sequence_p {
        last = book.Value_vectors[entry][dim]
      }
      index_divisor *= len(book.Multiplicands)
    }
  }
}
func (book *Codebook) BuildVQType2() {
  book.allocateTable()
  for entry := range book.Value_vectors {
    last := 0.0
    offset := entry * book.Dimensions
    for dim := range book.Value_vectors[entry] {
      // TODO: Same thing with absolute value in the java implementation
      book.Value_vectors[entry][dim] = float64(book.Multiplicands[offset]) * book.Delta_value + book.Minimum_value + last
      if book.Sequence_p {
        last = book.Value_vectors[entry][dim]
      }
      offset++
    }
  }
}

func (book *Codebook) AssignCodewords() {
  max_len := 0
  for i := range book.Entries {
    if book.Entries[i].Unused { continue }
    if book.Entries[i].Length > max_len {
      max_len = book.Entries[i].Length
    }
  }
  min := make([]uint32, max_len + 1)
  for i := range book.Entries {
    if book.Entries[i].Unused { continue }
    length := book.Entries[i].Length
    book.Entries[i].Codeword = min[length]
    min[length]++
    for j := length + 1; j < len(min); j++ {
      next := min[j-1] << 1
      if next > min[j] {
        min[j] = next
      }
    }
    for j := length - 1; j >= 0; j-- {
      prev := min[j+1] >> 1
      if prev > min[j] {
        min[j] = prev
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
        book.Entries[current_entry + i].Length = current_length
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
      for i := range book.Entries {
        flag := br.ReadBits(1) == 1
        if flag {
          book.Entries[i].Length = int(br.ReadBits(5)) + 1
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
