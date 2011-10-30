package vorbis

import (
  "bytes"
  "fmt"
)

type setupHeader struct {
  Codebooks []Codebook

  Floor_configs   []Floor
  Residue_configs []Residue
  Mapping_configs []Mapping
  Mode_configs    []Mode
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
  header.Mapping_configs = make([]Mapping, mapping_count)
  for i := range header.Mapping_configs {
    header.Mapping_configs[i] = readMapping(br, num_channels, floor_count, residue_count)
  }

  // Read Modes
  mode_count := int(br.ReadBits(6) + 1)
  header.Mode_configs = make([]Mode, mode_count)
  for i := range header.Mode_configs {
    header.Mode_configs[i] = readMode(br, len(header.Mapping_configs))
  }

  // Frame
  if br.ReadBits(1) == 0 {
    panic("Framing error in setup header.")
  }
}

