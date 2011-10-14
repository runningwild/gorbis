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
func HuffmanAssignmentspec(c gospec.Context) {
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

  c.Specify("Codebook with a single zero-bit entry", func() {
    var codebook vorbis.Codebook
    codebook.Entries = make([]vorbis.CodebookEntry, 1)
    codebook.Entries[0].Length = 0
    codebook.AssignCodewords()
    c.Expect(codebook.Entries[0].Codeword, Equals, uint32(0))
  })
}





















