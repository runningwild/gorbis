package vorbis

import (
  "bytes"
  "fmt"
)

type setupHeader struct {
  Codebooks []Codebook

  floor_configs  []float64
}

func (header setupHeader) read(buffer *bytes.Buffer, num_channels int) {
  b,_ := buffer.ReadByte()
  if b != 5 {
    panic(fmt.Sprintf("Header type == %d, expected type == 5.", b))
  }

  if string(buffer.Next(6)) != "vorbis" {
    panic("vorbis string not found in setup header")
  }


  // Decode Codebooks
  num_Codebooks,_ := buffer.ReadByte()
  num_Codebooks++
  header.Codebooks = make([]Codebook, int(num_Codebooks))
  br := MakeBitReader(buffer)
  for i := range header.Codebooks {
    header.Codebooks[i].decode(br)
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
  header.floor_configs = make([]float64, floor_count)
  fmt.Printf("Parsing floors %d\n", floor_count)
  for _ = range header.floor_configs {
    floor_type := int(br.ReadBits(16))
    switch floor_type {
      case 0:
        var f Floor0
        fmt.Printf("0\n")
        f.HeaderDecode(br, len(header.Codebooks))
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
        bits := ilog(uint32(num_channels) - 1)
        br.ReadBits(bits)
        br.ReadBits(bits)
      }
    }
    if br.ReadBits(2) != 0 {
      panic("Non-zero reserved bits found when reading mappings.")
    }
    if submaps > 1 {
      for j := 0; j < int(num_channels); j++ {
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

