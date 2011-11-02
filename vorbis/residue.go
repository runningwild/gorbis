package vorbis
import "fmt"
type Residue interface {
  Decode(br *BitReader, books []Codebook, ch int, do_not_decode []bool, n int) [][]float64
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

type residue0 struct {
  residueBase
}
func (r *residue0) Decode(br *BitReader, books []Codebook, ch int, do_not_decode []bool, n int) [][]float64 {
  fmt.Printf("starti ng format 0\n")
  return r.residueBase.decode(br, books, ch, do_not_decode, n, 0)
}

type residue1 struct {
  residueBase
}
func (r *residue1) Decode(br *BitReader, books []Codebook, ch int, do_not_decode []bool, n int) [][]float64 {
  print("starting format 1\n")
  return r.residueBase.decode(br, books, ch, do_not_decode, n, 1)
}

type residue2 struct {
  residueBase
}
func (r *residue2) Decode(br *BitReader, books []Codebook, ch int, do_not_decode []bool, n int) [][]float64 {
fmt.Printf("Starting format 2: %d, %v\n", ch, do_not_decode)
  decode := false
  for i := range do_not_decode {
    if !do_not_decode[i] {
      decode = true
      break
    }
  }

  var data []float64
  fmt.Printf("Decode: %t\n", decode)
  if !decode {
    data = make([]float64, ch * n)
  } else {
    data = r.decode(br, books, 1, []bool{ false }, ch * n, 1)[0]
  }

  // TODO: spec says to do this step even if we are using a blank array
  //       that just seems dumb
  output := make([][]float64, ch)
  for i := range output {
    output[i] = make([]float64, n)
  }
  for i := 0; i < n; i++ {
    for j := 0; j < ch; j++ {
      output[j][i] = data[i + ch * j]
    }
  }

  return output
}

func (r *residueBase) decode(br *BitReader, books []Codebook, ch int, do_not_decode []bool, n int, mode int) [][]float64 {
  limit_begin := r.begin
  if limit_begin > n {
    limit_begin = n
  }
  limit_end := r.end
  if limit_end > n {
    limit_end = n
  }

  book := books[r.classbook]
  classwords_per_codeword := book.Dimensions
  n_to_read := limit_end - limit_begin
  fmt.Printf("Residue Diff: %d %d %d %d\n", limit_begin, limit_end, n_to_read, n)
  fmt.Printf("Reside data: %d %d\n", r.classbook, r.partition_size)
  partitions_to_read := n_to_read / r.partition_size

  residue_vecs := make([][]float64, ch)
  for i := range residue_vecs {
    residue_vecs[i] = make([]float64, n)
  }

  // In any mode we cut out early if there is nowhere to put the data
  if n_to_read == 0 {
    fmt.Printf("Bailing: nothing to read\n")
    return residue_vecs
  }
  fmt.Printf("don't decode: %v\n", do_not_decode)

  classifications := make([][]int, ch)
  for i := range classifications {
    classifications[i] = make([]int, partitions_to_read)
  }

  for pass := 0; pass < 8; pass++ {
    partition_count := 0
    for partition_count < partitions_to_read {
      if pass == 0 {
        for j := 0; j < ch; j++ {
          if do_not_decode[j] { continue }
          temp := book.DecodeScalar(br)
          for i := classwords_per_codeword - 1; i >= 0; i-- {
            classifications[j][i + partition_count] = temp % r.num_classifications
            temp /= r.num_classifications
          }
        }
      }
      for i := 0; i < classwords_per_codeword && partition_count < partitions_to_read; i++ {
        for j := 0; j < ch; j++ {
          if do_not_decode[j] {
            partition_count++
            continue
          }
          vq_class := classifications[j][partition_count]
          vq_book := r.books[vq_class][pass]
          if vq_book == -1 {
            partition_count++
            continue
          }
          book := books[vq_book]

          n := r.partition_size
          v := residue_vecs[j]
          offset := limit_begin + partition_count * r.partition_size

          if mode == 0 {
            // format 0
            print("format 0\n")
            step := n / book.Dimensions
            for i := 0; i < step; i++ {
              temp := book.DecodeVector(br)
              for j := 0; j < book.Dimensions; j++ {
                v[offset + i + j * step] += temp[j]
                print("temp: ", temp[j])
              }
            }
          } else {
            // format 1 (used by format 2)
            print("format 1\n")
            i := 0
            for i < n {
              fmt.Printf("i: %d\n", i)
              temp := book.DecodeVector(br)
              for j := 0; j < book.Dimensions; j++ {
                v[offset + i] += temp[j]
                print("temp: ", temp[j])
                i++
              }
            }
          }
          partition_count++
        }
      }
    }
  }

  return residue_vecs
}

func readResidue(br *BitReader) Residue {
  var residue Residue

  var base residueBase
  residue_type := int(br.ReadBits(16))
  if residue_type < 0 || residue_type > 2 {
    panic("Unknown residue type.")
  }
  base.read(br)
  switch residue_type {
    case 0:
      residue = &residue0{ base }
    case 1:
      residue = &residue1{ base }
    case 2:
      residue = &residue2{ base }
  }

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
      } else {
        r.books[i][j] = -1
      }
    }
  }
}
