package vorbis

import (
  "encoding/binary"
  "bytes"
  "fmt"
)

type commentHeader struct {
  Vendor_string string
  User_comments []string
  Framing       bool
}

func (header *commentHeader) read(buffer *bytes.Buffer) {
  b, _ := buffer.ReadByte()
  if b != 3 {
    panic(fmt.Sprintf("Header type == %d, expected type == 3.", b))
  }

  if string(buffer.Next(6)) != "vorbis" {
    panic("vorbis string not found in comment header")
  }

  var length uint32
  binary.Read(buffer, binary.LittleEndian, &length)
  header.Vendor_string = string(buffer.Next(int(length)))

  binary.Read(buffer, binary.LittleEndian, &length)
  header.User_comments = make([]string, length)
  for i := range header.User_comments {
    binary.Read(buffer, binary.LittleEndian, &length)
    header.User_comments[i] = string(buffer.Next(int(length)))
  }

  framing, _ := buffer.ReadByte()
  header.Framing = (framing & 0x1) != 0
  if !header.Framing {
    panic("Framing bit not set in comment header")
  }

  fmt.Printf("comment: %v\n", header)
}
