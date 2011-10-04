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
  codebooks []codebook
}

type codebookEntry struct {
  unused bool
  length int
}

type codebook struct {
  dimensions int
  entries    []codebookEntry
  multiplicands []uint32

  minimum_value float32
  delta_value   float32
  sequence_p    bool
}

func (book *codebook) decode(br *BitReader) {
  fmt.Printf("Reading codebook\n")
  if br.ReadBits(24) != 0x564342 {
    panic("Codebook sync pattern not found")
  }

  book.dimensions = int(br.ReadBits(16))
  num_entries := int(br.ReadBits(24))
  book.entries = make([]codebookEntry, num_entries)
  ordered := br.ReadBits(1) == 1


  // Decode codeword lengths
  if ordered {
    current_entry := 0
    for current_entry < num_entries {
      current_length := int(br.ReadBits(5)) + 1
      number := int(br.ReadBits(ilog(uint32(num_entries - current_entry))))
      for i := 0; i < number; i++ {
        book.entries[current_entry + i].length = current_length
      }
      current_length++
      current_entry += number
      if current_entry >= num_entries {
        panic("Error decoding codebooks")
      }
    }
  } else {
    sparse := br.ReadBits(1) == 1
    if sparse {
      for i := range book.entries {
        flag := br.ReadBits(1) == 1
        if flag {
          book.entries[i].length = int(br.ReadBits(5)) + 1
        } else {
          book.entries[i].unused = true
        }
      }
    } else {
      for i := range book.entries {
        book.entries[i].length = int(br.ReadBits(5)) + 1
      }
    }
  }

  // read the vector lookup table
  codebook_lookup_type := int(br.ReadBits(4))
  switch codebook_lookup_type {
    case 0:
      // no vector lookup

    case 1:
      fallthrough
    case 2:
      book.minimum_value := math.Float32frombits(br.ReadBits(32))
      book.delta_value := math.Float32frombits(br.ReadBits(32))
      codebook_value_bits := int(br.ReadBits(4) + 1)
      book.sequence_p := br.ReadBits(1) == 1
      var codebook_lookup_values int
      if codebook_lookup_type == 1 {
        codebook_lookup_values = Lookup1Values(book.entries, book.dimensions)
      } else {
        codebook_lookup_values = len(book.entries) * book.dimensions
      }
      book.multiplicands = make([]uint32, codebook_lookup_values)
      for i := range book.multiplicands {
        book.multiplicands = br.ReadBits(codebook_value_bits)
      }

    default:
      panic("Unknown vectork lookup method")
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


  // Decode codebooks
  num_codebooks,err := v.buffer.ReadByte()
  check(err)
  v.setup_header.codebooks = make([]codebook, int(num_codebooks))
  br := MakeBitReader(v.buffer)
  for i := range v.setup_header.codebooks {
    v.setup_header.codebooks[i].decode(br)
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


