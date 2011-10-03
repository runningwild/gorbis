package vorbis

type bitReader struct {
  data []byte
  byte_pos,bit_pos int
}
func makeBitReader(data []byte) *bitReader {
  var br bitReader
  br.data = data
  return &br
}

// 0 <= n < 8
func (br *bitReader) readAtMost(n int) (read int, bits uint32) {
  bits = uint32(br.data[br.byte_pos])
  bits = bits >> uint(br.bit_pos)
  bits = bits & ((1 << uint(n)) - 1)
  read = 8 - br.bit_pos
  if read > n {
    read = n
  }
  br.bit_pos += read
  if br.bit_pos == 8 {
    br.bit_pos = 0
    br.byte_pos++
  }
  return
}

// 0 <= n < 32
func (br *bitReader) ReadBits(n int) (bits uint32) {
  pos := 0
  for n > 0 {
    read,next := br.readAtMost(n)
    bits = bits | (next << uint(pos))
    pos += read
    n -= read
  }
  return
}

