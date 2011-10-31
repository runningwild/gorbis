package vorbis

import (
  "fmt"
  "ogg"
  "bytes"
  "os"
  "math"
)

var magic_string string = "\x01vorbis"

func check(err os.Error) {
  if err != nil {
    panic(err.String())
  }
}

func (v *vorbisDecoder) readAudioPacket(page ogg.Page, num_channels int) {
  br := MakeBitReader(v.buffer)
  if br.ReadBits(1) != 0 {
    fmt.Printf("Warning: Not an audio packet")
    return
  }

  mode_number := ilog(uint32(len(v.Mode_configs)) - 1)
  mode := v.Mode_configs[mode_number]
  mapping := v.Mapping_configs[mode.mapping]

  window := v.generateWindow(br, mode)

  // Floor curves
  for i := 0; i < num_channels; i++ {
    submap_number := mapping.muxs[i]
    floor_number := mapping.submaps[submap_number].floor
    floor := v.Floor_configs[floor_number]
    floor.Decode(br, v.Codebooks, len(window))
  }
}

func (v *vorbisDecoder) generateWindow(br *BitReader, mode Mode) []float64 {
  var n int
  if mode.block_flag {
    n = v.Blocksize_1
  } else {
    n = v.Blocksize_0
  }

  // window selection and setup
  var prev_window_flag, next_window_flag bool
  if mode.block_flag {
    prev_window_flag = br.ReadBits(1) == 1
    next_window_flag = br.ReadBits(1) == 1
  }

  // An end of stream error is possible here, just bail on this packet
  // TODO: Need to make it possible to reset the bitreader, or just need to make
  // a new one whenever it encounters an error
  // TODO: Mayeb just panic?  Perhaps after successfully reading the headers all
  // panics can just indicate the current packet is bad and further decoding
  // can go ahead as normal
  if br.CheckError() != nil {
    return nil
  }



  // generate window
  window_center := n / 2
  var left_window_start, left_window_end, left_n int
  var right_window_start, right_window_end, right_n int
  if mode.block_flag {
    if prev_window_flag {
      left_window_start = 0
      left_window_end = window_center
      left_n = n / 2
    } else {
      left_window_start = n / 4 - v.Blocksize_0 / 4
      left_window_end = n / 4 + v.Blocksize_0 / 4
      left_n = v.Blocksize_0 / 2
    }
    if next_window_flag {
      left_window_start = window_center
      left_window_end = n
      left_n = n / 2
    } else {
      left_window_start = (n * 3) / 4 - v.Blocksize_0 / 4
      left_window_end = (n * 3) / 4 + v.Blocksize_0 / 4
      left_n = v.Blocksize_0 / 2
    }
  }

  window := make([]float64, n)
  const pi_over_2 = math.Pi / 2
  for i := left_window_start; i < left_window_end; i++ {
    base := (float64(i - left_window_start) + 0.5) / float64(left_n) * pi_over_2
    window[i] = math.Sin(pi_over_2 * math.Pow(math.Sin(base), 2))
  }
  for i := left_window_end; i < right_window_start; i++ {
    window[i] = 1
  }
  for i := right_window_start; i < right_window_end; i++ {
    base := (float64(i - right_window_start) + 0.5) / float64(right_n) * pi_over_2 + pi_over_2
    window[i] = math.Sin(pi_over_2 * math.Pow(math.Sin(base), 2))
  }

  return window
}

func init() {
  ogg.RegisterFormat(magic_string, makeVorbisDecoder)
}

func makeVorbisDecoder() ogg.Codec {
  var v vorbisDecoder
  v.buffer = bytes.NewBuffer(make([]byte, 0, 256))
  return &v
}

type codecMode int
const (
  readId codecMode = iota
  readComment
  readSetup
  readData
)

type vorbisDecoder struct {
  mode codecMode
  idHeader
  commentHeader
  setupHeader
  buffer *bytes.Buffer
}
func (v *vorbisDecoder) Add(page ogg.Page) {
  // Might neet to paste a few packets together before we start reading
  v.buffer.Write(page.Data)
  if len(page.Segment_table) > 0 && page.Segment_table[len(page.Segment_table) - 1] == 255 {
    return
  }

  switch v.mode {
    case readId:
      fmt.Printf("Read id\n")
      v.idHeader.read(v.buffer)
      fmt.Printf("After id: %d\n", v.buffer.Len())
      v.mode++
      fallthrough

    case readComment:
      if v.buffer.Len() == 0 {
        // This could happen if the id and comment headers aren't in the
        // same packet.  The spec really doesn't specify how it should be.
        // TODO: For this pair of headers this might be specified to never
        //       happen, so remove this if statement if that's the case.
        return
      }
      fmt.Printf("Read comment\n")
      v.commentHeader.read(v.buffer)
      fmt.Printf("After comment: %d\n", v.buffer.Len())
      v.mode++
      fallthrough

    case readSetup:
      if v.buffer.Len() == 0 {
        // This could happen if the comment and setup headers aren't in the
        // same packet.  The spec really doesn't specify how it should be.
        return
      }
      v.setupHeader.read(v.buffer, int(v.Channels))
      v.mode++

    case readData:
      v.readAudioPacket(page, int(v.Channels))
      v.buffer.Truncate(0)
  }
}
func (v *vorbisDecoder) Finish() {
  fmt.Printf("Finished stream\n")
}


