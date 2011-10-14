package ogg

import (
  "io"
  "os"
  "encoding/binary"
  "hash/crc32"
  "fmt"
)

type HeaderFixed struct {
  Capture_pattern [4]byte  // needs to be "OggS"

  Version                 uint8
  Header_type             uint8
  Granule_position        uint64
  Bitstream_serial_number uint32
  Page_sequence_number    uint32
  Crc_checksum            uint32
  Page_segments           uint8
}

type Header struct {
  HeaderFixed
  Segment_table []uint8
}

type Page struct {
  Header
  Data []byte
}


type Codec interface {
  Add(Page)
  Finish()
}

var ogg_table *crc32.Table

type Format func() Codec
var formats map[string]Format
func init() {
  ogg_table = crc32.MakeTable(0x04c11db7)
  formats = make(map[string]Format)
}

func RegisterFormat(magic string, format Format) {
  formats[magic] = format
}

func GetCodec(page Page) Codec {
  for magic,format := range formats {
    if len(page.Data) >= len(magic) && string(page.Data[0 : len(magic)]) == magic {
      return format()
    }
  }
  fmt.Printf("Unknown format: %s\n", string(page.Data))
  return nil
}

func DecodePage(in io.Reader) (Page, os.Error) {
  var page Page
  err := binary.Read(in, binary.LittleEndian, &page.HeaderFixed)
  if err != nil {
    return page, err
  }
  page.Segment_table = make([]uint8, int(page.Page_segments))
  _,err = io.ReadFull(in, page.Segment_table)
  if err != nil {
    return page, err
  }

  remaining_data := 0
  for _,v := range page.Segment_table {
    remaining_data += int(v)
  }
  page.Data = make([]byte, remaining_data)
  _,err = io.ReadFull(in, page.Data)
  if err != nil {
    return page, err
  }
  // The checksum is made by zeroing the checksum value and CRC-ing the entire page
  checksum := page.Crc_checksum
  page.Crc_checksum = 0
  crc := crc32.New(ogg_table)
  binary.Write(crc, binary.LittleEndian, &page.HeaderFixed)
  crc.Write(page.Segment_table)
  crc.Write(page.Data)

  if crc.Sum32() != checksum {
    // TODO: Figure out why this CRC isn't working
//    return page, os.NewError(fmt.Sprintf("CRC failed: expected %x, got %x.", checksum, crc.Sum32()))
  }
  return page, nil
}

func Decode(in io.Reader) os.Error {
  fmt.Printf("")
  streams := make(map[uint32]Codec)
  var page Page
  var err os.Error
  for ; err == nil; page,err = DecodePage(in) {
//    fmt.Printf("Header: %d %v\n", page.Header.Header_type, page.Header.Segment_table)
    serial := page.Bitstream_serial_number
    if page.Header_type & 0x2 != 0 {
      // First packet in a bitstream, shouldn't already have a codec for it
      // check for one first, then make one
      if _,ok := streams[serial]; ok {
        // TODO: issue a warning, there was already a codec here
        continue
      }
      streams[serial] = GetCodec(page)
      fmt.Printf("Set codec(%x): %v\n", serial, streams[serial])
    }
    codec := streams[serial]
    if codec == nil { 
      fmt.Printf("nil\n")
      continue
    }
    codec.Add(page)
    if page.Header_type & 0x4 != 0 {
      codec.Finish()
      streams[serial] = nil,false
    }
  }
  if err == nil {
    return os.NewError("Quit processing without reaching EOF")
  }
  if err != os.EOF {
    return err
  }
  if len(streams) > 0 {
    return os.NewError(fmt.Sprintf("%d streams did not complete.", len(streams)))
  }
  return nil
}









