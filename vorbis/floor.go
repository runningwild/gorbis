package vorbis

type Floor interface {}

type Floor0 struct {
  order            int
  rate             int
  bark_map_size    int
  amplitude_bits   int
  amplitude_offset int
  books []int
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
  
}

func decodeFloor1(br *BitReader) Floor {
  var f Floor1
  num_partitions := int(br.ReadBits(5))
  max_class := -1
  partition_class_list := make([]int, num_partitions)
  for i := range partition_class_list {
    partition_class_list[i] = int(br.ReadBits(4))
    if partition_class_list[i] > max_class {
      max_class = partition_class_list[i]
    }
  }
  class_dims := make([]int, max_class + 1)
  class_subclasses := make([]int, max_class + 1)
  class_masterbooks := make([]int, max_class + 1)
  subclass_books := make([][]int, max_class + 1)
  for i := 0; i <= max_class; i++ {
    class_dims[i] = int(br.ReadBits(3) + 1)
    class_subclasses[i] = int(br.ReadBits(2))
    if class_subclasses[i] > 0 {
      class_masterbooks[i] = int(br.ReadBits(8))
    }
    subclass_books[i] = make([]int, int(1 << uint(class_subclasses[i])))
    for j := 0; j < int(1 << uint(class_subclasses[i])); j++ {
      // 12
      subclass_books[i][j] = int(br.ReadBits(8) - 1)
    }
  }
  multiplier := int(br.ReadBits(2)) + 1
  _ = multiplier
  rangebits := int(br.ReadBits(4))
  xvals := make([]int, 2)
  xvals[0] = 0
  xvals[1] = int(1 << uint(rangebits))
  for i := 0; i < num_partitions; i++ {
    current_class_number := partition_class_list[i]
    for j := 0; j < class_dims[current_class_number]; j++ {
      xvals = append(xvals, int(br.ReadBits(rangebits)))
    }
  }
  return &f
}
