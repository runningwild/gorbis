package vorbis

type Floor interface {
  Decode(*BitReader, []Codebook) []float64
}

type Floor0 struct {
  order            int
  rate             int
  bark_map_size    int
  amplitude_bits   int
  amplitude_offset int

  books []int
}

func (f *Floor0) Decode(br *BitReader, codebooks []Codebook) []float64 {
  amplitude := int(br.ReadBits(f.amplitude_bits))
  if amplitude > 0 {
    var coefficient []float64
    print(coefficient)
    book_num := ilog(uint32(len(f.books)))
    if book_num >= len(f.books) {
      panic("Floor codebook index out of range.")
    }
    book := codebooks[f.books[book_num]]
    last := 0.0
    for len(coefficient) < f.order {
      temp := book.DecodeVector(br)
      for _,v := range temp {
        coefficient = append(coefficient, v + last)
      }
      last = coefficient[len(coefficient) - 1]
    }
    return coefficient[0 : f.order]
  }
  return nil
}

func readFloor(br *BitReader, num_codebooks int) Floor {
  floor_type := int(br.ReadBits(16))
  switch floor_type {
    case 0:
      return decodeFloor0(br, num_codebooks)
    case 1:
      return decodeFloor1(br)
    default:
      panic("Unknown floor type.")
  }
  return nil
}

func decodeFloor0(br *BitReader, max_books int) Floor {
  var f Floor0
  f.order = int(br.ReadBits(8))
  f.rate = int(br.ReadBits(16))
  f.bark_map_size = int(br.ReadBits(16))
  f.amplitude_bits = int(br.ReadBits(6))
  f.amplitude_offset = int(br.ReadBits(8))
  num_books := int(br.ReadBits(4) + 1)
  f.books = make([]int, num_books)
  for i := range f.books {
    f.books[i] = int(br.ReadBits(8))
    if f.books[i] < 0 || f.books[i] >= max_books {
      panic("Invalid codebook specified in Floor0 decode.")
    }
  }
  return &f
}

type Floor1 struct {
  X []int

  multiplier int

  partition_classes []int

  classes []floorClass
}

type floorClass struct {
  dimensions int
  subclass   int
  masterbook int

  subclass_books []int
}

func (f *Floor1) Decode(br *BitReader, codebooks []Codebook) []float64 {
  return nil
}

func decodeFloor1(br *BitReader) Floor {
  var f Floor1

  num_partitions := int(br.ReadBits(5))
  max_class := -1
  f.partition_classes = make([]int, num_partitions)
  for i := range f.partition_classes {
    f.partition_classes[i] = int(br.ReadBits(4))
    if f.partition_classes[i] > max_class {
      max_class = f.partition_classes[i]
    }
  }

  f.classes = make([]floorClass, max_class + 1)
  for i := range f.classes {
    class := &f.classes[i]
    class.dimensions = int(br.ReadBits(3) + 1)
    class.subclass = int(br.ReadBits(2))
    if class.subclass > 0 {
      class.masterbook = int(br.ReadBits(8))
    }
    class.subclass_books = make([]int, int(1 << uint(class.subclass)))
    for j := 0; j < int(1 << uint(class.subclass)); j++ {
      // 12
      class.subclass_books[j] = int(br.ReadBits(8) - 1)
    }
  }

  f.multiplier = int(br.ReadBits(2)) + 1
  rangebits := int(br.ReadBits(4))
  f.X = make([]int, 2)
  f.X[0] = 0
  f.X[1] = int(1 << uint(rangebits))
  for _,class := range f.partition_classes {
    for j := 0; j < f.classes[class].dimensions; j++ {
      f.X = append(f.X, int(br.ReadBits(rangebits)))
    }
  }
  return &f
}
