package vorbis

import (
  "fmt"
  "ogg"
  "encoding/binary"
  "bytes"
  "os"
  "math"
)

var magic_string string = "\x01vorbis"

type idHeaderFixed struct {
  Version uint32
  Channels uint8
  Sample_rate uint32
  Bitrate_maximum uint32
  Bitrate_nominal uint32
  Bitrate_minimum uint32
}

type idHeader struct {
  idHeaderFixed
  Blocksize_0 int
  Blocksize_1 int
}

func (v *vorbisDecoder) readIdHeader(page ogg.Page) {
  b,err := v.buffer.ReadByte()
  check(err)
  if b != 1 {
    panic(fmt.Sprintf("Header type == %d, expected type == 1.", b))
  }

  if string(v.buffer.Next(6)) != "vorbis" {
    panic("vorbis string not found in id header")
  }

  err = binary.Read(v.buffer, binary.LittleEndian, &v.id_header.idHeaderFixed)
  check(err)
  var block_sizes uint8
  err = binary.Read(v.buffer, binary.LittleEndian, &block_sizes)
  check(err)
  v.id_header.Blocksize_0 = int(1 << (block_sizes & 0x0f))
  v.id_header.Blocksize_1 = int(1 << ((block_sizes & 0xf0) >> 4))

  var framing uint8
  err = binary.Read(v.buffer, binary.LittleEndian, &framing)
  check(err)
  if framing != 1 {
    panic("Id header not properly framed")
  }

  if v.id_header.Version != 0 {
    panic(fmt.Sprintf("Unexpected version number in id header: %d", v.id_header.Version))
  }
  if v.id_header.Channels == 0 {
    panic("Channels set to zero in id header")
  }
  if v.id_header.Sample_rate == 0 {
    panic("Sample rate set to zero in id header")
  }
  if v.id_header.Blocksize_0 != 64 &&
     v.id_header.Blocksize_0 != 128 &&
     v.id_header.Blocksize_0 != 256 &&
     v.id_header.Blocksize_0 != 512 &&
     v.id_header.Blocksize_0 != 1024 &&
     v.id_header.Blocksize_0 != 2048 &&
     v.id_header.Blocksize_0 != 4096 &&
     v.id_header.Blocksize_0 != 8192 {
     panic(fmt.Sprintf("Invalid block 0 size: %d", v.id_header.Blocksize_0))
  }
  if v.id_header.Blocksize_1 != 64 &&
     v.id_header.Blocksize_1 != 128 &&
     v.id_header.Blocksize_1 != 256 &&
     v.id_header.Blocksize_1 != 512 &&
     v.id_header.Blocksize_1 != 1024 &&
     v.id_header.Blocksize_1 != 2048 &&
     v.id_header.Blocksize_1 != 4096 &&
     v.id_header.Blocksize_1 != 8192 {
     panic(fmt.Sprintf("Invalid block 1 size: %d", v.id_header.Blocksize_1))
  }
  if v.id_header.Blocksize_0 > v.id_header.Blocksize_1 {
    panic(fmt.Sprintf("Block 0 size > block 1 size: %d > %d", v.id_header.Blocksize_0, v.id_header.Blocksize_1))
  }

  if v.buffer.Len() > 0 {
    // Shouldn't be anything leftover, log a warning?
  }
}

type commentHeader struct {
  Vendor_string string
  User_comments []string
  Framing       bool
}

func check(err os.Error) {
  if err != nil {
    panic(err.String())
  }
}

func (v *vorbisDecoder) readCommentHeader(page ogg.Page) {
  b,err := v.buffer.ReadByte()
  check(err)
  if b != 3 {
    panic(fmt.Sprintf("Header type == %d, expected type == 3.", b))
  }

  if string(v.buffer.Next(6)) != "vorbis" {
    panic("vorbis string not found in comment header")
  }

  var length uint32
  err = binary.Read(v.buffer, binary.LittleEndian, &length)
  check(err)
  v.comment_header.Vendor_string = string(v.buffer.Next(int(length)))

  err = binary.Read(v.buffer, binary.LittleEndian, &length)
  check(err)
  v.comment_header.User_comments = make([]string, length)
  for i := range v.comment_header.User_comments {
    err := binary.Read(v.buffer, binary.LittleEndian, &length)
    check(err)
    v.comment_header.User_comments[i] = string(v.buffer.Next(int(length)))
  }

  framing,err := v.buffer.ReadByte()
  check(err)
  v.comment_header.Framing = (framing & 0x1) != 0
  if !v.comment_header.Framing {
    panic("Framing bit not set in comment header")
  }

  fmt.Printf("comment: %v\n", v.comment_header)
}

type setupHeader struct {
  Codebooks []Codebook
}

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

//func (book *Codebook) VectorLookup1() {
//  last := 0
//  index_divisor := 1
//  for lookup_offset := 0 
//  for i := 0; i < book.Dimensions; i++ {
//    multiplicand_offset := 
//  }
//}

func (book *Codebook) VectorLookup2() {
  
}

func (book *Codebook) decode(br *BitReader) {
  fmt.Printf("Reading Codebook\n")
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
      panic("Unknown vectork lookup method")
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

func ilog(n uint32) int {
  e := 31
  bit := uint32(1) << 31
  for e > 0 {
    if (n & bit) != 0 { return e }
    bit = bit >> 1
    e--
  }
  return 0
}

func (v *vorbisDecoder) readSetupHeader(page ogg.Page) {
  b,err := v.buffer.ReadByte()
  check(err)
  if b != 5 {
    panic(fmt.Sprintf("Header type == %d, expected type == 5.", b))
  }

  if string(v.buffer.Next(6)) != "vorbis" {
    panic("vorbis string not found in setup header")
  }


  // Decode Codebooks
  num_Codebooks,err := v.buffer.ReadByte()
  check(err)
  v.setup_header.Codebooks = make([]Codebook, int(num_Codebooks))
  br := MakeBitReader(v.buffer)
  for i := range v.setup_header.Codebooks {
    v.setup_header.Codebooks[i].decode(br)
  }

}

func (v *vorbisDecoder) readAudioPacket(page ogg.Page) {
  b,err := v.buffer.ReadByte()
  check(err)
  fmt.Printf("Packet type: %d\n", b)
}


func init() {
  ogg.RegisterFormat(magic_string, makeVorbisDecoder)
}

func makeVorbisDecoder() ogg.Codec {
  var v vorbisDecoder
  v.buffer = bytes.NewBuffer(make([]byte, 0, 256))
  return &v
}

type mode int
const (
  readId mode = iota
  readComment
  readSetup
  readData
)

type vorbisDecoder struct {
  mode mode
  id_header idHeader
  comment_header commentHeader
  setup_header setupHeader
  buffer *bytes.Buffer
}
func (v *vorbisDecoder) Add(page ogg.Page) {
  v.buffer.Write(page.Data)
  if len(page.Segment_table) > 0 && page.Segment_table[len(page.Segment_table) - 1] == 255 {
    return
  }
  switch v.mode {
    case readId:
      fmt.Printf("Read id\n")
      v.readIdHeader(page)
      v.mode++

    case readComment:
      fmt.Printf("Read comment\n")
      v.readCommentHeader(page)
      v.mode++

    case readSetup:
      fmt.Printf("Read setup\n")
      v.readSetupHeader(page)
      v.mode++

    case readData:
      v.readAudioPacket(page)
  }
}
func (v *vorbisDecoder) Finish() {
  fmt.Printf("Finished stream\n")
}


