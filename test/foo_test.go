package foo_test

import (
  . "gospec"
  "gospec"
  "ogg"
  _ "ogg/vorbis"
  "os"
)

func OggSpec(c gospec.Context) {
  //  v := []uint8{1,1,3,0,0,9}
  //  00001001 00000000 00000000 00000011 00000001 00000001
  //                                                      1
  //                                           001 0000000
  //       001 00000000 00000000 00000011 00000
  //  00001

  //  c.Specify("Bitreader reads from an io.Reader properly", func() {
  //    br := ogg.MakeBitReader(v)
  //    c.Expect(uint32(0x1), Equals, br.ReadBits(1))
  //    c.Expect(uint32(0x80), Equals, br.ReadBits(10))
  //    c.Expect(uint32(0x20000060), Equals, br.ReadBits(32))
  //    c.Expect(uint32(0x1), Equals, br.ReadBits(5))
  //  })

  c.Specify("Read a file", func() {
    f, err := os.Open("metroid.ogg")
    c.Assume(err, Equals, nil)
    err = ogg.Decode(f)
    c.Assume(err, Equals, nil)
  })
}
