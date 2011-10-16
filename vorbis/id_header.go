package vorbis

import (
  "encoding/binary"
  "bytes"
  "fmt"
)

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

func (header *idHeader) read(buffer *bytes.Buffer) {
  b,_ := buffer.ReadByte()
  if b != 1 {
    panic(fmt.Sprintf("Header type == %d, expected type == 1.", b))
  }

  if string(buffer.Next(6)) != "vorbis" {
    panic("vorbis string not found in id header")
  }

  binary.Read(buffer, binary.LittleEndian, &header.idHeaderFixed)
  var block_sizes uint8
  binary.Read(buffer, binary.LittleEndian, &block_sizes)
  header.Blocksize_0 = int(1 << (block_sizes & 0x0f))
  header.Blocksize_1 = int(1 << ((block_sizes & 0xf0) >> 4))

  var framing uint8
  binary.Read(buffer, binary.LittleEndian, &framing)
  if framing != 1 {
    panic("Id header not properly framed")
  }

  if header.Version != 0 {
    panic(fmt.Sprintf("Unexpected version number in id header: %d", header.Version))
  }
  if header.Channels == 0 {
    panic("Channels set to zero in id header")
  }
  if header.Sample_rate == 0 {
    panic("Sample rate set to zero in id header")
  }
  if header.Blocksize_0 != 64 &&
     header.Blocksize_0 != 128 &&
     header.Blocksize_0 != 256 &&
     header.Blocksize_0 != 512 &&
     header.Blocksize_0 != 1024 &&
     header.Blocksize_0 != 2048 &&
     header.Blocksize_0 != 4096 &&
     header.Blocksize_0 != 8192 {
     panic(fmt.Sprintf("Invalid block 0 size: %d", header.Blocksize_0))
  }
  if header.Blocksize_1 != 64 &&
     header.Blocksize_1 != 128 &&
     header.Blocksize_1 != 256 &&
     header.Blocksize_1 != 512 &&
     header.Blocksize_1 != 1024 &&
     header.Blocksize_1 != 2048 &&
     header.Blocksize_1 != 4096 &&
     header.Blocksize_1 != 8192 {
     panic(fmt.Sprintf("Invalid block 1 size: %d", header.Blocksize_1))
  }
  if header.Blocksize_0 > header.Blocksize_1 {
    panic(fmt.Sprintf("Block 0 size > block 1 size: %d > %d", header.Blocksize_0, header.Blocksize_1))
  }

  if buffer.Len() > 0 {
    // TODO: Shouldn't be anything leftover, log a warning?
  }
}
