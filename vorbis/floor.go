package vorbis

import "sort"

var inverse_db_table []float64
func init() {
  inverse_db_table = []float64 {
    1.0649863e-07,
    1.1341951e-07,
    1.2079015e-07,
    1.2863978e-07, 
    1.3699951e-07,
    1.4590251e-07,
    1.5538408e-07,
    1.6548181e-07, 
    1.7623575e-07,
    1.8768855e-07,
    1.9988561e-07,
    2.1287530e-07, 
    2.2670913e-07,
    2.4144197e-07,
    2.5713223e-07,
    2.7384213e-07, 
    2.9163793e-07,
    3.1059021e-07,
    3.3077411e-07,
    3.5226968e-07, 
    3.7516214e-07,
    3.9954229e-07,
    4.2550680e-07,
    4.5315863e-07, 
    4.8260743e-07,
    5.1396998e-07,
    5.4737065e-07,
    5.8294187e-07, 
    6.2082472e-07,
    6.6116941e-07,
    7.0413592e-07,
    7.4989464e-07, 
    7.9862701e-07,
    8.5052630e-07,
    9.0579828e-07,
    9.6466216e-07, 
    1.0273513e-06,
    1.0941144e-06,
    1.1652161e-06,
    1.2409384e-06, 
    1.3215816e-06,
    1.4074654e-06,
    1.4989305e-06,
    1.5963394e-06, 
    1.7000785e-06,
    1.8105592e-06,
    1.9282195e-06,
    2.0535261e-06, 
    2.1869758e-06,
    2.3290978e-06,
    2.4804557e-06,
    2.6416497e-06, 
    2.8133190e-06,
    2.9961443e-06,
    3.1908506e-06,
    3.3982101e-06, 
    3.6190449e-06,
    3.8542308e-06,
    4.1047004e-06,
    4.3714470e-06, 
    4.6555282e-06,
    4.9580707e-06,
    5.2802740e-06,
    5.6234160e-06, 
    5.9888572e-06,
    6.3780469e-06,
    6.7925283e-06,
    7.2339451e-06, 
    7.7040476e-06,
    8.2047000e-06,
    8.7378876e-06,
    9.3057248e-06, 
    9.9104632e-06,
    1.0554501e-05,
    1.1240392e-05,
    1.1970856e-05, 
    1.2748789e-05,
    1.3577278e-05,
    1.4459606e-05,
    1.5399272e-05, 
    1.6400004e-05,
    1.7465768e-05,
    1.8600792e-05,
    1.9809576e-05, 
    2.1096914e-05,
    2.2467911e-05,
    2.3928002e-05,
    2.5482978e-05, 
    2.7139006e-05,
    2.8902651e-05,
    3.0780908e-05,
    3.2781225e-05, 
    3.4911534e-05,
    3.7180282e-05,
    3.9596466e-05,
    4.2169667e-05, 
    4.4910090e-05,
    4.7828601e-05,
    5.0936773e-05,
    5.4246931e-05, 
    5.7772202e-05,
    6.1526565e-05,
    6.5524908e-05,
    6.9783085e-05, 
    7.4317983e-05,
    7.9147585e-05,
    8.4291040e-05,
    8.9768747e-05, 
    9.5602426e-05,
    0.00010181521,
    0.00010843174,
    0.00011547824, 
    0.00012298267,
    0.00013097477,
    0.00013948625,
    0.00014855085, 
    0.00015820453,
    0.00016848555,
    0.00017943469,
    0.00019109536, 
    0.00020351382,
    0.00021673929,
    0.00023082423,
    0.00024582449, 
    0.00026179955,
    0.00027881276,
    0.00029693158,
    0.00031622787, 
    0.00033677814,
    0.00035866388,
    0.00038197188,
    0.00040679456, 
    0.00043323036,
    0.00046138411,
    0.00049136745,
    0.00052329927, 
    0.00055730621,
    0.00059352311,
    0.00063209358,
    0.00067317058, 
    0.00071691700,
    0.00076350630,
    0.00081312324,
    0.00086596457, 
    0.00092223983,
    0.00098217216,
    0.0010459992,
    0.0011139742, 
    0.0011863665,
    0.0012634633,
    0.0013455702,
    0.0014330129, 
    0.0015261382,
    0.0016253153,
    0.0017309374,
    0.0018434235, 
    0.0019632195,
    0.0020908006,
    0.0022266726,
    0.0023713743, 
    0.0025254795,
    0.0026895994,
    0.0028643847,
    0.0030505286, 
    0.0032487691,
    0.0034598925,
    0.0036847358,
    0.0039241906, 
    0.0041792066,
    0.0044507950,
    0.0047400328,
    0.0050480668, 
    0.0053761186,
    0.0057254891,
    0.0060975636,
    0.0064938176, 
    0.0069158225,
    0.0073652516,
    0.0078438871,
    0.0083536271, 
    0.0088964928,
    0.009474637,
    0.010090352,
    0.010746080, 
    0.011444421,
    0.012188144,
    0.012980198,
    0.013823725, 
    0.014722068,
    0.015678791,
    0.016697687,
    0.017782797, 
    0.018938423,
    0.020169149,
    0.021479854,
    0.022875735, 
    0.024362330,
    0.025945531,
    0.027631618,
    0.029427276, 
    0.031339626,
    0.033376252,
    0.035545228,
    0.037855157, 
    0.040315199,
    0.042935108,
    0.045725273,
    0.048696758, 
    0.051861348,
    0.055231591,
    0.058820850,
    0.062643361, 
    0.066714279,
    0.071049749,
    0.075666962,
    0.080584227, 
    0.085821044,
    0.091398179,
    0.097337747,
    0.10366330, 
    0.11039993,
    0.11757434,
    0.12521498,
    0.13335215, 
    0.14201813,
    0.15124727,
    0.16107617,
    0.17154380, 
    0.18269168,
    0.19456402,
    0.20720788,
    0.22067342, 
    0.23501402,
    0.25028656,
    0.26655159,
    0.28387361, 
    0.30232132,
    0.32196786,
    0.34289114,
    0.36517414, 
    0.38890521,
    0.41417847,
    0.44109412,
    0.46975890, 
    0.50028648,
    0.53279791,
    0.56742212,
    0.60429640, 
    0.64356699,
    0.68538959,
    0.72993007,
    0.77736504, 
    0.82788260,
    0.88168307,
    0.9389798,
    1.0,
  }
}

type Floor interface {
  // If Decode returns nil it indicates that this floor curve is unused.
  Decode(*BitReader, []Codebook, int) []float64
}

type Floor0 struct {
  order            int
  rate             int
  bark_map_size    int
  amplitude_bits   int
  amplitude_offset int

  books []int
}

// TODO: Floor0 is not done, but should be completed prior to release
// TODO: Need to find a vorbis file that actually uses floor0 for testing
func (f *Floor0) Decode(br *BitReader, codebooks []Codebook, n int) []float64 {
  panic("Floor0 not complete: notify devs")
  amplitude := int(br.ReadBits(f.amplitude_bits))
  if amplitude > 0 {
    var coefficient []float64
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

    // This is acceptable according to the spec, so we try to proceed and
    // avoid panicing
    if br.CheckError() != nil {
      return nil
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
  Xs []int

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

func (f *Floor1) Decode(br *BitReader, codebooks []Codebook, n int) []float64 {
  // Check the non-zero bit
  if br.ReadBits(1) == 0 {
    return nil
  }

  // Decode Y values
  Ys := f.decodeYs(br, codebooks)
  if Ys == nil {
    return nil
  }

  // Amplitude value synthesis
  return f.computeCurve(br, Ys, codebooks, n)
}

func (f *Floor1) decodeYs(br *BitReader, codebooks []Codebook) []int {
  var rnge uint32
  switch f.multiplier - 1 {
    case 0:
      rnge = 256
    case 1:
      rnge = 128
    case 2:
      rnge = 86
    case 3:
      rnge = 64
  }
  var Ys []int
  Ys = append(Ys, int(br.ReadBits(ilog(rnge - 1))))
  Ys = append(Ys, int(br.ReadBits(ilog(rnge - 1))))
  for _,class_index := range f.partition_classes {
    class := f.classes[class_index]
    cdim := class.dimensions
    cbits := uint(class.subclass)
    csub := (1 << cbits) - 1
    cval := 0
    if cbits > 0 {
      cval = codebooks[class.masterbook].DecodeScalar(br)
    }
    for j := 0; j < cdim; j++ {
      book := class.subclass_books[cval & csub]
      cval = cval >> cbits
      if book >= 0 {
        Ys = append(Ys, codebooks[book].DecodeScalar(br))
      } else {
        Ys = append(Ys, 0)
      }
    }
  }
  return Ys
}

func (f *Floor1) computeCurve(br *BitReader, Ys []int, codebooks []Codebook, n int) []float64 {
  var rnge int
  switch f.multiplier - 1 {
    case 0:
      rnge = 256
    case 1:
      rnge = 128
    case 2:
      rnge = 86
    case 3:
      rnge = 64
  }

  step_2 := make([]bool, len(f.Xs))
  step_2[0] = true
  step_2[1] = true

  final_Ys := make([]int, len(f.Xs))
  final_Ys[0] = Ys[0]
  final_Ys[1] = Ys[1]

  // Amplitude Value Synthesis
  for i := 2; i < len(f.Xs); i++ {
    low := lowNeighbor(f.Xs, i)
    high := highNeighbor(f.Xs, i)
    predicted := renderPoint(f.Xs[low], final_Ys[low], f.Xs[high], final_Ys[high], f.Xs[i])
    val := Ys[i]

    high_room := rnge - predicted
    low_room := predicted
    var room int
    if high_room < low_room {
      room = high_room * 2
    } else {
      room = low_room * 2
    }

    if val == 0 {
      step_2[i] = false
      final_Ys[i] = predicted
    } else {
      step_2[low] = true
      step_2[high] = true
      step_2[i] = true
      if val >= room {
        if high_room > low_room {
          final_Ys[i] = val - low_room + predicted
        } else {
          final_Ys[i] = predicted - val + high_room - 1
        }
      } else {
        if val % 2 == 1 {
          final_Ys[i] = predicted - (val + 1) / 2
        } else {
          final_Ys[i] = predicted + val / 2
        }
      }
    }
  }

  // Curve Synthesis
  Xs := make([]int, len(f.Xs))
  copy(Xs, f.Xs)

  sort.Sort(&curveVals{Xs, final_Ys, step_2})

  hx := 0
  lx := 0
  ly := final_Ys[0] * f.multiplier

  floor := make([]int, n)
  var hy int
  for i := 1; i < len(final_Ys); i++ {
    if step_2[i] {
      hy = final_Ys[i] * f.multiplier
      hx = Xs[i]
      renderLine(lx, ly, hx, hy, floor)
      lx = hx
      ly = hy
    }
  }

  if hx < n {
    // TODO: This is silly, it's just a horizontal line
    renderLine(hx, hy, n, hy, floor)
  } else if hx > n {
    floor = floor[0 : n]
  }
  amps := make([]float64, n)
  for i := range amps {
    amps[i] = inverse_db_table[floor[i]]
  }
  return amps
}

type curveVals struct {
  X,Y []int
  step2 []bool
}
func (c *curveVals) Len() int {
  return len(c.X)
}
func (c *curveVals) Swap(i,j int) {
  c.X[i],c.X[j] = c.X[j],c.X[i]
  c.Y[i],c.Y[j] = c.Y[j],c.Y[i]
  c.step2[i],c.step2[j] = c.step2[j],c.step2[i]
}
func (c *curveVals) Less(i,j int) bool {
  return c.X[i] < c.X[j]
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
  f.Xs = make([]int, 2)
  f.Xs[0] = 0
  f.Xs[1] = int(1 << uint(rangebits))
  for _,class := range f.partition_classes {
    for j := 0; j < f.classes[class].dimensions; j++ {
      f.Xs = append(f.Xs, int(br.ReadBits(rangebits)))
    }
  }
  return &f
}
