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
  num_Codebooks++
  check(err)
  v.setup_header.Codebooks = make([]Codebook, int(num_Codebooks))
  br := MakeBitReader(v.buffer)
  for i := range v.setup_header.Codebooks {
    v.setup_header.Codebooks[i].decode(br)
  }

  // Read Time Domain Transfers
  // These are placeholder values in the vorbis 1
  // bitstream, but they must be read anyway
  time_transfers_count := int(br.ReadBits(6) + 1)
  for i := 0; i < time_transfers_count; i++ {
    if br.ReadBits(16) != 0 {
      panic("Time Domain Transfer Value != 0")
    }
  }

  // Read Floors
  floor_count := int(br.ReadBits(6) + 1)
  v.floor_configs = make([]float64, floor_count)
  fmt.Printf("Parsing floors %d\n", floor_count)
  for _ = range v.floor_configs {
    floor_type := int(br.ReadBits(16))
    switch floor_type {
      case 0:
        var f Floor0
        fmt.Printf("0\n")
        f.HeaderDecode(br, len(v.setup_header.Codebooks))
      case 1:
        var f Floor1
        fmt.Printf("1\n")
        f.HeaderDecode(br)
      default:
        panic("Unknown floor type.")
    }
  }

  // Read Resiudes
  residue_count := int(br.ReadBits(6) + 1)
  residue_types := make([]int, residue_count)
  for i := range residue_types {
    residue_types[i] = int(br.ReadBits(16))
    if residue_types[i] > 2 {
      panic("Unknown residue type.")
    }
    br.ReadBits(24)
    br.ReadBits(24)
    br.ReadBits(24) // + 1
    residue_classifications := br.ReadBits(6) + 1
    br.ReadBits(8)
    // TODO: There are some checks that can go here
    residue_cascades := make([]uint32, residue_classifications)
    for i := range residue_cascades {
      high_bits := 0
      low_bits := int(br.ReadBits(3))
      bit_flag := br.ReadBits(1) != 0
      if bit_flag {
        high_bits = int(br.ReadBits(5))
      }
      //residue cascade
      residue_cascades[i] = uint32(high_bits * 8 + low_bits)
    }
    for i := 0; i < int(residue_classifications); i++ {
      for j := 0; j < 8; j++ {
        if (residue_cascades[i] & (uint32(1) << uint32(j))) != 0 {
          br.ReadBits(8)
        }
      }
    }
  }

  // Read Mappings
  mapping_count := int(br.ReadBits(6) + 1)
  for i := 0; i < mapping_count; i++ {
    mapping_type := int(br.ReadBits(16))
    if mapping_type != 0 {
      panic("Found a non-zero mapping type.")
    }
    flag := br.ReadBits(1) != 0
    submaps := 1
    if flag {
      submaps = int(br.ReadBits(4) + 1)
    }
    if br.ReadBits(1) != 0 {
      coupling_steps := int(br.ReadBits(8) + 1)
      for j := 0; j < coupling_steps; j++ {
        bits := ilog(uint32(v.id_header.Channels) - 1)
        br.ReadBits(bits)
        br.ReadBits(bits)
      }
    }
    if br.ReadBits(2) != 0 {
      panic("Non-zero reserved bits found when reading mappings.")
    }
    if submaps > 1 {
      for j := 0; j < int(v.id_header.Channels); j++ {
        br.ReadBits(4)
      }
    }
    for j := 0; j < submaps; j++ {
      br.ReadBits(8)
      br.ReadBits(8)
      br.ReadBits(8)
    }
  }

  // Read Modes
  mode_count := int(br.ReadBits(6) + 1)
  for i := 0; i < mode_count; i++ {
    br.ReadBits(1)
    window_type := int(br.ReadBits(16))
    if window_type != 0 {
      panic("Found non-zero window type while reading modes.")
    }
    transform_type := int(br.ReadBits(16))
    if transform_type != 0 {
      panic("Found non-zero transform type while reading modes.")
    }
    br.ReadBits(8)
  }

  // Frame
  if br.ReadBits(1) == 0 {
    panic("Framing error in setup header.")
  }
}

type Floor interface {}
type Floor0 struct {
  order            int
  rate             int
  bark_map_size    int
  amplitude_bits   int
  amplitude_offset int
  books []int
}
func (f *Floor0) HeaderDecode(br *BitReader, max_books int) {
  f.order = int(br.ReadBits(8))
  f.rate = int(br.ReadBits(16))
  f.bark_map_size = int(br.ReadBits(16))
  f.amplitude_bits = int(br.ReadBits(6))
  f.amplitude_offset = int(br.ReadBits(8))
  num_books := int(br.ReadBits(4) + 1)
  f.books = make([]int, num_books)
  for i := range f.books {
    f.books[i] = int(br.ReadBits(8))
    if f.books[i] < 0 || f.books[i] >= max_books {
      panic("Invalid codebook specified in Floor0 decode.")
    }
  }
}

type Floor1 struct {
  
}
func (f *Floor1) HeaderDecode(br *BitReader) {
  num_partitions := int(br.ReadBits(5))
  max_class := -1
  partition_class_list := make([]int, num_partitions)
  for i := range partition_class_list {
    partition_class_list[i] = int(br.ReadBits(4))
    if partition_class_list[i] > max_class {
      max_class = partition_class_list[i]
    }
  }
  class_dims := make([]int, max_class + 1)
  class_subclasses := make([]int, max_class + 1)
  class_masterbooks := make([]int, max_class + 1)
  subclass_books := make([][]int, max_class + 1)
  for i := 0; i <= max_class; i++ {
    class_dims[i] = int(br.ReadBits(3) + 1)
    class_subclasses[i] = int(br.ReadBits(2))
    if class_subclasses[i] > 0 {
      class_masterbooks[i] = int(br.ReadBits(8))
    }
    subclass_books[i] = make([]int, int(1 << uint(class_subclasses[i])))
    for j := 0; j < int(1 << uint(class_subclasses[i])); j++ {
      // 12
      subclass_books[i][j] = int(br.ReadBits(8) - 1)
    }
  }
  multiplier := int(br.ReadBits(2)) + 1
  _ = multiplier
  rangebits := int(br.ReadBits(4))
  xvals := make([]int, 2)
  xvals[0] = 0
  xvals[1] = int(1 << uint(rangebits))
  for i := 0; i < num_partitions; i++ {
    current_class_number := partition_class_list[i]
    for j := 0; j < class_dims[current_class_number]; j++ {
      xvals = append(xvals, int(br.ReadBits(rangebits)))
    }
  }
}

func (v *vorbisDecoder) readAudioPacket(page ogg.Page) {
  br := MakeBitReader(v.buffer)
  if br.ReadBits(1) != 0 {
    fmt.Printf("Warning: Not an audio packet")
  }
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
  id_header      idHeader
  comment_header commentHeader
  setup_header   setupHeader
  floor_configs  []float64
  buffer *bytes.Buffer
}
func (v *vorbisDecoder) Add(page ogg.Page) {
  // Might neet to paste a few packets together before we start reading
  v.buffer.Write(page.Data)
  if len(page.Segment_table) > 0 && page.Segment_table[len(page.Segment_table) - 1] == 255 {
    return
  }

  switch v.mode {
    case readId:
      fmt.Printf("Read id\n")
      v.readIdHeader(page)
      fmt.Printf("After id: %d\n", v.buffer.Len())
      v.mode++
      fallthrough

    case readComment:
      if v.buffer.Len() == 0 {
        // This could happen if the id and comment headers aren't in the
        // same packet.  The spec really doesn't specify how it should be.
        // TODO: For this pair of headers this might be specified to never
        //       happen, so remove this if statement if that's the case.
        return
      }
      fmt.Printf("Read comment\n")
      v.readCommentHeader(page)
      fmt.Printf("After comment: %d\n", v.buffer.Len())
      v.mode++
      fallthrough

    case readSetup:
      if v.buffer.Len() == 0 {
        // This could happen if the comment and setup headers aren't in the
        // same packet.  The spec really doesn't specify how it should be.
        return
      }
      v.readSetupHeader(page)
      v.mode++

    case readData:
      v.readAudioPacket(page)
      v.buffer.Truncate(0)
  }
}
func (v *vorbisDecoder) Finish() {
  fmt.Printf("Finished stream\n")
}


