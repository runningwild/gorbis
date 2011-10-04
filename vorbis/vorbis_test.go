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

