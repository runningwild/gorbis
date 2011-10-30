package vorbis

type Mode struct {
  block_flag bool
  mapping    int
}

func readMode(br *BitReader, num_mappings int) Mode {
  var m Mode

  m.block_flag = br.ReadBits(1) == 1

  // Don't bother storing this, we know it has to be zero
  window_type := int(br.ReadBits(16))
  if window_type != 0 {
    panic("Found non-zero window type while reading modes.")
  }

  // Don't bother storing this, we know it has to be zero
  transform_type := int(br.ReadBits(16))
  if transform_type != 0 {
    panic("Found non-zero transform type while reading modes.")
  }

  m.mapping = int(br.ReadBits(8))
  if m.mapping >= num_mappings {
    panic("Mode mapping value is out of range.")
  }

  return m
}
