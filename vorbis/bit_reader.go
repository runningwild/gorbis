package vorbis

import "io"

type BitReader struct {
  in      io.ByteReader
  current byte
  bit_pos int
  err     error
}

func MakeBitReader(in io.ByteReader) *BitReader {
  var br BitReader
  br.in = in
  br.bit_pos = 8
  return &br
}

func (br *BitReader) CheckError() error {
  return br.err
}

// 0 <= n < 8
func (br *BitReader) readAtMost(n int) (read int, bits uint32) {
  bits = uint32(br.current)
  bits = bits >> uint(br.bit_pos)
  bits = bits & ((1 << uint(n)) - 1)
  read = 8 - br.bit_pos
  if read > n {
    read = n
  }
  br.bit_pos += read
  if br.bit_pos == 8 {
    br.bit_pos = 0
    var err error
    br.current, err = br.in.ReadByte()
    if err != nil {
      br.err = err
    }
  }
  return
}

var total int

// 0 <= n < 32
func (br *BitReader) ReadBits(n int) uint32 {
  total += n
  if br.err != nil {
    return 0
  }
  var bits uint32
  pos := 0
  for n > 0 {
    read, next := br.readAtMost(n)
    bits = bits | (next << uint(pos))
    pos += read
    n -= read
  }
  return bits
}
