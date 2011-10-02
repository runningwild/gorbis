package ogg

import (
  "io"
  "os"
  "encoding/binary"
  "hash/crc32"
  "fmt"
)

type bufferedBitReader struct {
  in     io.Reader

  // pointer to either buff0 or buff1 representing unread data
  buff *[]byte

  // current position in the stream
  byte_pos, bit_pos int

  // last byte in the stream, if it is known
  base,end_byte int

  // underlying arrays that buff is based on
  buff0 []byte
  buff1 []byte

  // Number of bytes that this buffered reader should always have ready
  size int
}



func (bbr *bufferedBitReader) routine() {
  n,err := bbr.in.Read(bbr.buff0)
  if err != nil {
    bbr.end_byte = n
  } else {
    n,err = bbr.in.Read(bbr.buff1)

  }
}
func makeBufferedBitReader(in io.Reader, size int) *bufferedBitReader {
  var bbr bufferedBitReader
  bbr.in = in
  bbr.size = size
  bbr.buff0 = make([]byte, bbr.size)
  bbr.buff1 = make([]byte, bbr.size)

  go bbr.routine()
  return &bbr
}

type bitReader struct {
  data []byte
  byte_pos,bit_pos int
}
func MakeBitReader(data []byte) *bitReader {
  var br bitReader
  br.data = data
  return &br
}

// 0 <= n < 8
func (br *bitReader) readAtMost(n int) (read int, bits uint32) {
  bits = uint32(br.data[br.byte_pos])
  bits = bits >> uint(br.bit_pos)
  bits = bits & ((1 << uint(n)) - 1)
  read = 8 - br.bit_pos
  if read > n {
    read = n
  }
  br.bit_pos += read
  if br.bit_pos == 8 {
    br.bit_pos = 0
    br.byte_pos++
  }
  return
}

// 0 <= n < 32
func (br *bitReader) ReadBits(n int) (bits uint32) {
  pos := 0
  for n > 0 {
    read,next := br.readAtMost(n)
    bits = bits | (next << uint(pos))
    pos += read
    n -= read
  }
  return
}

type oggPageHeaderFixed struct {
  Capture_pattern [4]byte  // needs to be "OggS"

  Version                 uint8
  Header_type             uint8
  Granule_position        uint64
  Bitstream_serial_number uint32
  Page_sequence_number    uint32
  Crc_checksum            uint32
  Page_segments           uint8
}

type oggPageHeader struct {
  oggPageHeaderFixed
  Segment_table []uint8
}

type oggPage struct {
  header  oggPageHeader
  data []byte
}


type Codec interface {
  Add(oggPage)
  Finish()
}

type vorbis struct {
}
func (v *vorbis) Add(page oggPage) {
  fmt.Printf("Added page\n")
}
func (v *vorbis) Finish() {
  fmt.Printf("Finished stream\n")
}
func VorbisFormat(page oggPage) Codec {
  if len(page.data) < 7 {
    return nil
  }
  if string(page.data[0:7]) != "\x01vorbis" {
    return nil
  }
  return &vorbis{}
}

var ogg_table *crc32.Table

type Format func() Codec
var formats map[string]Format
func init() {
  ogg_table = crc32.MakeTable(0x04c11db7)
  formats = make(map[string]Format)
  RegisterFormat("\x01vorbis", func() Codec { return &vorbis{} })
}

func RegisterFormat(magic string, format Format) {
  formats[magic] = format
}

func GetCodec(page oggPage) Codec {
  for magic,format := range formats {
    if len(page.data) >= len(magic) && string(page.data[0 : len(magic)]) == magic {
      return format()
    }
  }
  return nil
}

func DecodePage(in io.Reader) (oggPage, os.Error) {
  var page oggPage
  err := binary.Read(in, binary.LittleEndian, &page.header.oggPageHeaderFixed)
  if err != nil {
    return page, err
  }
  page.header.Segment_table = make([]uint8, int(page.header.Page_segments))
  _,err = io.ReadFull(in, page.header.Segment_table)
  if err != nil {
    return page, err
  }

  remaining_data := 0
  for _,v := range page.header.Segment_table {
    remaining_data += int(v)
  }
  page.data = make([]byte, remaining_data)
  _,err = io.ReadFull(in, page.data)
  if err != nil {
    return page, err
  }
  // The checksum is made by zeroing the checksum value and CRC-ing the entire page
  checksum := page.header.Crc_checksum
  page.header.Crc_checksum = 0
  crc := crc32.New(ogg_table)
  binary.Write(crc, binary.LittleEndian, &page.header.oggPageHeaderFixed)
  crc.Write(page.header.Segment_table)
  crc.Write(page.data)

  if crc.Sum32() != checksum {
    // TODO: Figure out why this CRC isn't working
//    return page, os.NewError(fmt.Sprintf("CRC failed: expected %x, got %x.", checksum, crc.Sum32()))
  }
  return page, nil
}

func Decode(in io.Reader) os.Error {
  fmt.Printf("")
  streams := make(map[uint32]Codec)
  var page oggPage
  var err os.Error
  for ; err == nil; page,err = DecodePage(in) {
    serial := page.header.Bitstream_serial_number
    //fmt.Printf("Header: %x\n", page.header.Header_type)
    if page.header.Header_type & 0x2 != 0 {
      // First packet in a bitstream, shouldn't already have a codec for it
      // check for one first, then make one
      if _,ok := streams[serial]; ok {
        // TODO: issue a warning, there was already a codec here
        continue
      }
      streams[serial] = GetCodec(page)
      fmt.Printf("Set codec(%x): %v", serial, streams[serial])
    }
    codec := streams[serial]
    if codec == nil { 
      fmt.Printf("nil\n")
      continue
    }
    codec.Add(page)
    if page.header.Header_type & 0x4 != 0 {
      codec.Finish()
      streams[serial] = nil
    }
  }
  return err
}









