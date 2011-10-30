package vorbis

type Residue interface {
  read(br *BitReader)
}

type residueBase struct {
  begin               int
  end                 int
  partition_size      int
  num_classifications int
  classbook           int

  cascades []uint32
  books    [][]int
}

type residue1 struct {
  residueBase
}

type residue2 struct {
  residueBase
}

type residue3 struct {
  residueBase
}

func readResidue(br *BitReader) Residue {
  var residue Residue
  residue_type := int(br.ReadBits(16))
  if residue_type < 0 || residue_type > 2 {
    panic("Unknown residue type.")
  }
  print(residue_type, "\n")
  switch residue_type {
    case 1:
      residue = &residue1{}
    case 2:
      residue = &residue2{}
    case 3:
      residue = &residue3{}
  }
  residue.read(br)

  return residue
}

func (r *residueBase) read(br *BitReader) {
  r.begin = int(br.ReadBits(24))
  r.end = int(br.ReadBits(24))
  r.partition_size = int(br.ReadBits(24) + 1)
  r.num_classifications = int(br.ReadBits(6) + 1)
  r.classbook = int(br.ReadBits(8))

  // TODO: There are some checks that can go here
  cascades := make([]uint32, r.num_classifications)
  for i := range cascades {
    high_bits := 0
    low_bits := int(br.ReadBits(3))
    bit_flag := br.ReadBits(1) != 0
    if bit_flag {
      high_bits = int(br.ReadBits(5))
    }
    cascades[i] = uint32(high_bits * 8 + low_bits)
  }

  r.books = make([][]int, r.num_classifications)
  for i := 0; i < r.num_classifications; i++ {
    r.books[i] = make([]int, 8)
    for j := 0; j < 8; j++ {
      if (cascades[i] & (uint32(1) << uint32(j))) != 0 {
        r.books[i][j] = int(br.ReadBits(8))
      }
    }
  }
}
