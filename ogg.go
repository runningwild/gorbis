package ogg

import (
  "errors"
  "io"
  "encoding/binary"
  "hash/crc32"
  "fmt"
  "bytes"
)

type HeaderFixed struct {
  Capture_pattern [4]byte // needs to be "OggS"

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

type Packet struct {
  Granule_position     uint64
  Page_sequence_number uint32
  Data                 []byte
}

type Codec interface {
  Input() chan<- Packet
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
  for magic, format := range formats {
    if len(page.Data) >= len(magic) && string(page.Data[0:len(magic)]) == magic {
      return format()
    }
  }
  fmt.Printf("Unknown format: %s\n", string(page.Data))
  return nil
}

func DecodePage(in io.Reader) (Page, error) {
  var page Page
  err := binary.Read(in, binary.LittleEndian, &page.HeaderFixed)
  if err != nil {
    return page, err
  }
  page.Segment_table = make([]uint8, int(page.Page_segments))
  _, err = io.ReadFull(in, page.Segment_table)
  if err != nil {
    return page, err
  }

  remaining_data := 0
  for _, v := range page.Segment_table {
    remaining_data += int(v)
  }
  page.Data = make([]byte, remaining_data)
  _, err = io.ReadFull(in, page.Data)
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

type codecBuffer struct {
  codec  Codec
  buffer *bytes.Buffer
}

func Decode(in io.Reader) error {
  fmt.Printf("")
  streams := make(map[uint32]*codecBuffer)
  var page Page
  var err error
  for ; err == nil; page, err = DecodePage(in) {
    serial := page.Bitstream_serial_number
    if page.Header_type&0x2 != 0 {
      // First packet in a bitstream, shouldn't already have a codec for it
      // check for one first, then make one
      if _, ok := streams[serial]; ok {
        // TODO: issue a warning, there was already a codec here
        continue
      }
      streams[serial] = &codecBuffer{GetCodec(page), bytes.NewBuffer(nil)}
    }
    cb, ok := streams[serial]
    if !ok {
      fmt.Printf("!ok\n")
      continue
    }
    for _, seg_len := range page.Segment_table {
      cb.buffer.Write(page.Data[0:seg_len])
      page.Data = page.Data[seg_len:]
      if seg_len != 255 {
        cb.codec.Input() <- Packet{
          page.Granule_position,
          page.Page_sequence_number,
          cb.buffer.Bytes(),
        }
        cb.buffer = bytes.NewBuffer(nil)
      }
    }
    if page.Header_type&0x4 != 0 {
      close(cb.codec.Input())
      streams[serial] = nil, false
    }
  }
  if err == nil {
    return errors.New("Quit processing without reaching EOF")
  }
  if err != io.EOF {
    return err
  }
  if len(streams) > 0 {
    return errors.New(fmt.Sprintf("%d streams did not complete.", len(streams)))
  }
  return nil
}
