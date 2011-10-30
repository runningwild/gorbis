package vorbis_test

import (
  . "gospec"
  "gospec"
  "ogg/vorbis"
  "bytes"
  "testing"
)

func BenchmarkLookup1Values(b *testing.B) {
  for i := 0; i < b.N; i++ {
    vorbis.Lookup1Values(2048,8)
  }
}

func BenchmarkLookup1ValuesJava(b *testing.B) {
  for i := 0; i < b.N; i++ {
    vorbis.Lookup1ValuesJava(2048,8)
  }
}

func Lookup1Spec(c gospec.Context) {
  c.Expect(vorbis.Lookup1Values(48,2), Equals, 6)
  c.Expect(vorbis.Lookup1Values(49,2), Equals, 7)
  c.Expect(vorbis.Lookup1ValuesJava(48,2), Equals, 6)
  c.Expect(vorbis.Lookup1ValuesJava(49,2), Equals, 7)
}

func BitReaderSpec(c gospec.Context) {
  v := []uint8{1,1,3,0,0,9}
//  00001001 00000000 00000000 00000011 00000001 00000001
//                                                      1
//                                           001 0000000
//       001 00000000 00000000 00000011 00000
//  00001

  c.Specify("Bitreader reads from an io.Reader properly", func() {
    br := vorbis.MakeBitReader(bytes.NewBuffer(v))
    c.Expect(uint32(0x1), Equals, br.ReadBits(1))
    c.Expect(uint32(0x80), Equals, br.ReadBits(10))
    c.Expect(uint32(0x20000060), Equals, br.ReadBits(32))
    c.Expect(uint32(0x1), Equals, br.ReadBits(5))
  })
}

// TODO: Check that over- and under-specified codebooks raise an error
// Entry Length Codeword
//   0      2     00
//   1      4     0100
//   2      4     0101
//   3      4     0110
//   4      4     0111
//   5      2     10
//   6      3     110
//   7      3     111
//
// For testing, concatenating them all in order (notice they are reversed):
// 00 0010 1010 0110 1110 01 011 111
// 00 0010 1010 0110 1110 0101 1111
//  0    2    A    6    E    5    F = 0x02A6E5F

func HuffmanAssignmentSpec(c gospec.Context) {
  c.Specify("Basic huffman assignment", func() {
    var codebook vorbis.Codebook
    codebook.Entries = make([]vorbis.CodebookEntry, 8)
    codebook.Entries[0].Length = 2
    codebook.Entries[1].Length = 4
    codebook.Entries[2].Length = 4
    codebook.Entries[3].Length = 4
    codebook.Entries[4].Length = 4
    codebook.Entries[5].Length = 2
    codebook.Entries[6].Length = 3
    codebook.Entries[7].Length = 3
    codebook.AssignCodewords()
    c.Expect(codebook.Entries[0].Codeword, Equals, uint32(0))
    c.Expect(codebook.Entries[1].Codeword, Equals, uint32(4))
    c.Expect(codebook.Entries[2].Codeword, Equals, uint32(5))
    c.Expect(codebook.Entries[3].Codeword, Equals, uint32(6))
    c.Expect(codebook.Entries[4].Codeword, Equals, uint32(7))
    c.Expect(codebook.Entries[5].Codeword, Equals, uint32(2))
    c.Expect(codebook.Entries[6].Codeword, Equals, uint32(6))
    c.Expect(codebook.Entries[7].Codeword, Equals, uint32(7))
  })

  c.Specify("Another huffman assignment", func() {
    var codebook vorbis.Codebook
    codebook.Entries = make([]vorbis.CodebookEntry, 8)
    codebook.Entries[0].Length = 1
    codebook.Entries[1].Length = 3
    codebook.Entries[2].Length = 4
    codebook.Entries[3].Length = 7
    codebook.Entries[4].Length = 2
    codebook.Entries[5].Length = 5
    codebook.Entries[6].Length = 6
    codebook.Entries[7].Length = 7
    codebook.AssignCodewords()
    c.Expect(codebook.Entries[0].Codeword, Equals, uint32(0))
    c.Expect(codebook.Entries[1].Codeword, Equals, uint32(4))
    c.Expect(codebook.Entries[2].Codeword, Equals, uint32(10))
    c.Expect(codebook.Entries[3].Codeword, Equals, uint32(0x58))
    c.Expect(codebook.Entries[4].Codeword, Equals, uint32(3))
    c.Expect(codebook.Entries[5].Codeword, Equals, uint32(0x17))
    c.Expect(codebook.Entries[6].Codeword, Equals, uint32(0x2D))
    c.Expect(codebook.Entries[7].Codeword, Equals, uint32(0x59))
  })

  c.Specify("Large huffman assignment", func() {
    var codebook vorbis.Codebook
    codebook.Entries = make([]vorbis.CodebookEntry, 100)
    lengths := []int{ 3, 8, 9, 13, 10, 12, 12, 12, 12, 12, 6, 4, 6, 8, 6, 8,
        10, 10, 11, 12, 8, 5, 4, 10, 4, 7, 8, 9, 10, 11, 13, 8, 10, 8, 9, 9,
        11, 12, 13, 14, 10, 6, 4, 9, 3, 5, 6, 8, 10, 11, 11, 8, 6, 9, 5, 5, 6,
        7, 9, 11, 12, 9, 7, 11, 6, 6, 6, 7, 8, 10, 12, 11, 9, 12, 7, 7, 6, 6,
        7, 9, 13, 12, 10, 13, 9, 8, 7, 7, 7, 8, 11, 15, 11, 15, 11, 10, 9, 8,
        7, 7 }
    for i := range codebook.Entries {
      codebook.Entries[i].Length = lengths[i]
    }
    codebook.AssignCodewords()
    codewords := []uint32{ 0x0, 0x20, 0x42, 0x430, 0x87, 0x219, 0x21a, 0x21b,
        0x220, 0x221, 0x9, 0x3, 0xa, 0x23, 0xb, 0x40, 0x89, 0x8a, 0x111,
        0x22c, 0x41, 0x9, 0x5, 0x108, 0x6, 0x22, 0x43, 0x85, 0x109, 0x117,
        0x431, 0x46, 0x11c, 0x70, 0x8f, 0xe2, 0x23a, 0x22d, 0x8ec, 0x11da,
        0x1c6, 0x1d, 0x8, 0xe4, 0x5, 0xf, 0x24, 0x73, 0x1c7, 0x394, 0x395,
        0x94, 0x26, 0x12a, 0x18, 0x19, 0x27, 0x4b, 0x12b, 0x396, 0x477,
        0x1a0, 0x69, 0x397, 0x35, 0x36, 0x37, 0x70, 0xd1, 0x342, 0xd0c, 0x687,
        0x1c4, 0xd0d, 0x72, 0x73, 0x3a, 0x3b, 0x78, 0x1c5, 0x1c60, 0xe31,
        0x38d, 0x1c61, 0x1c7, 0xf2, 0x7a, 0x7b, 0x7c, 0xf3, 0x719, 0x23b6,
        0x7d0, 0x23b7, 0x7d1, 0x3e9, 0x1f5, 0xfb, 0x7e, 0x7f }

    for i := range codewords {
      c.Expect(codebook.Entries[i].Codeword, Equals, codewords[i])
    }
  })

  c.Specify("Codebook with a single zero-bit entry", func() {
    var codebook vorbis.Codebook
    codebook.Entries = make([]vorbis.CodebookEntry, 1)
    codebook.Entries[0].Length = 0
    codebook.AssignCodewords()
    c.Expect(codebook.Entries[0].Codeword, Equals, uint32(0))
  })
}

func HuffmanDecodeSpec(c gospec.Context) {
  c.Specify("Basic huffman decode", func() {
    var codebook vorbis.Codebook
    codebook.Entries = make([]vorbis.CodebookEntry, 8)
    codebook.Entries[0].Length = 2
    codebook.Entries[1].Length = 4
    codebook.Entries[2].Length = 4
    codebook.Entries[3].Length = 4
    codebook.Entries[4].Length = 4
    codebook.Entries[5].Length = 2
    codebook.Entries[6].Length = 3
    codebook.Entries[7].Length = 3
    codebook.AssignCodewords()

    v := []uint8{0x5F, 0x6E, 0x2A, 0x00}
    br := vorbis.MakeBitReader(bytes.NewBuffer(v))

    c.Expect(codebook.DecodeScalar(br), Equals, 7)
    c.Expect(codebook.DecodeScalar(br), Equals, 6)
    c.Expect(codebook.DecodeScalar(br), Equals, 5)
    c.Expect(codebook.DecodeScalar(br), Equals, 4)
    c.Expect(codebook.DecodeScalar(br), Equals, 3)
    c.Expect(codebook.DecodeScalar(br), Equals, 2)
    c.Expect(codebook.DecodeScalar(br), Equals, 1)
    c.Expect(codebook.DecodeScalar(br), Equals, 0)
  })
}





















