package vorbis

import (
  "bytes"
  "fmt"
)

type setupHeader struct {
  Codebooks []Codebook

  Floor_configs   []Floor
  Residue_configs []Residue
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
  codebook_count,_ := buffer.ReadByte()
  codebook_count++
  header.Codebooks = make([]Codebook, int(codebook_count))
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

  floor_count := int(br.ReadBits(6) + 1)
  header.Floor_configs = make([]Floor, floor_count)
  for i := range header.Floor_configs {
    header.Floor_configs[i] = readFloor(br, len(header.Codebooks))
  }

  // Read Resiudes
  residue_count := int(br.ReadBits(6) + 1)
  header.Residue_configs = make([]Residue, residue_count)
  for i := range header.Residue_configs {
    header.Residue_configs[i] = readResidue(br)
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

