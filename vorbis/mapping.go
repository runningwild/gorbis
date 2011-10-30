package vorbis

type Mapping interface {  
}

type Mapping0 struct {
  couplings []coupling

  muxs    []int

  submaps []submap
}
type coupling struct {
  magnitude int
  angle     int
}
type submap struct {
  floor   int
  residue int
}

func readMapping(br *BitReader, num_channels,num_floors,num_residues int) Mapping {
  mapping_type := int(br.ReadBits(16))
  if mapping_type != 0 {
    panic("Found a non-zero mapping type.")
  }
  var mapping Mapping0

  flag := br.ReadBits(1) != 0
  submaps := 1
  if flag {
    submaps = int(br.ReadBits(4) + 1)
  }
  if br.ReadBits(1) != 0 {
    coupling_steps := int(br.ReadBits(8) + 1)
    mapping.couplings = make([]coupling, coupling_steps)
    for i := range mapping.couplings {
      bits := ilog(uint32(num_channels) - 1)
      mapping.couplings[i].magnitude = int(br.ReadBits(bits))
      mapping.couplings[i].angle = int(br.ReadBits(bits))
      if mapping.couplings[i].magnitude == mapping.couplings[i].angle {
        panic("Mapping angle channel and mapping magnitude channel cannot be equal.")
      }
      if mapping.couplings[i].magnitude >= num_channels {
        panic("Mapping magnitude channel is out of range.")
      }
      if mapping.couplings[i].angle >= num_channels {
        panic("Mapping angle channel is out of range.")
      }
    }
  }
  if br.ReadBits(2) != 0 {
    panic("Non-zero reserved bits found when reading mappings.")
  }
  if submaps > 1 {
    mapping.muxs = make([]int, num_channels)
    for i := range mapping.muxs {
      mapping.muxs[i] = int(br.ReadBits(4))
      if mapping.muxs[i] >= submaps {
        panic("Mapping mux is out of range.")
      }
    }
  }

  mapping.submaps = make([]submap, submaps)
  for i := range mapping.submaps {
    br.ReadBits(8)  // explicitly discarded
    mapping.submaps[i].floor = int(br.ReadBits(8))
    if mapping.submaps[i].floor >= num_floors {
      panic("Mapping submap floor is out of range.")
    }
    mapping.submaps[i].residue = int(br.ReadBits(8))
    if mapping.submaps[i].residue >= num_residues {
      panic("Mapping submap residue is out of range.")
    }
  }

  return &mapping
}
