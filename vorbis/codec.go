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

func (v *vorbisDecoder) readAudioPacket(buffer io.ByteReader, num_channels int) {
  br := MakeBitReader(buffer)

  fmt.Printf("Start packet\n")
  if br.ReadBits(1) != 0 {
    fmt.Printf("Warning: Not an audio packet")
    return
  }

  mode_number := ilog(uint32(len(v.Mode_configs)) - 1)
  mode := v.Mode_configs[mode_number]
  mapping := v.Mapping_configs[mode.mapping]

  window := v.generateWindow(br, mode)

  // Floor curves
  // If the output for a floor for a particular channel is 'unused' that
  // element of the array will be nil
  floor_outputs := make([][]float64, num_channels)
  for i := 0; i < num_channels; i++ {
    submap_number := mapping.muxs[i]
    floor_number := mapping.submaps[submap_number].floor
    floor := v.Floor_configs[floor_number]

    // TODO: Not entirely sure we should be dividing by two here...
    floor_outputs[i] = floor.Decode(br, v.Codebooks, len(window) / 2)
  }

  if br.CheckError() != nil {
    // TODO: Need to handle an EOF error by zeroing channel data and skipping
    // to the add/overlap output stage
    panic("Not implemented")
  }

  // non-zero vector propagate
  // If any coupling has either angle or magnitude unused, then we can't use
  // either.  The spec doesn't seem to specify if channels can be used by more
  // than one coupling, but I suspect not, so this should be an ok way to
  // handle this.
  for _,coupling := range mapping.couplings {
    if floor_outputs[coupling.angle] == nil || floor_outputs[coupling.magnitude] == nil {
      floor_outputs[coupling.angle] = nil
      floor_outputs[coupling.magnitude] = nil
    }
  }

  // residue decode
  do_not_decode := make([]bool, num_channels)
  residue_outputs := make([][]float64, num_channels)
  for i,submap := range mapping.submaps {
    ch := 0
    for j := 0; j < num_channels; j++ {
      if mapping.muxs[j] == i {
        do_not_decode[ch] = floor_outputs[j] == nil
        ch++
      }
    }

    residues := v.Residue_configs[submap.residue].Decode(br, v.Codebooks, ch, do_not_decode, len(window)/2)

    ch = 0
    for j := 0; j < num_channels; j++ {
      if mapping.muxs[j] == i {
        residue_outputs[j] = residues[ch]
        ch++
      }
    }
  }

  // inverse coupling
  for i := len(mapping.couplings) - 1; i >= 0; i-- {
    mag := residue_outputs[mapping.couplings[i].magnitude]
    ang := residue_outputs[mapping.couplings[i].angle]
    var nM, nA float64
    for j := range mag {
      M := mag[j]
      A := ang[j]
      if M > 0 {
        if A > 0 {
          nM = M
          nA = M - A
        } else {
          nA = M
          nM = M + A
        }
      } else {
        if A > 0 {
          nM = M
          nA = M + A
        } else {
          nA = M
          nM = M - A
        }
      }
      mag[j] = nM
      ang[j] = nA
    }
  }

  // dot product
  fmt.Printf("%d %d\n", len(floor_outputs), len(residue_outputs))
  for i := range floor_outputs {
  fmt.Printf("%d %d\n", len(floor_outputs[i]), len(residue_outputs[i]))
    for j := range floor_outputs[i] {
      floor_outputs[i][j] *= residue_outputs[i][j]
    }
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
  v.input = make(chan ogg.Packet, 25)
  go v.routine()
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

  input chan ogg.Packet
}
func (v *vorbisDecoder) Input() chan<- ogg.Packet {
  return v.input
}
func (v *vorbisDecoder) routine() {
  for packet := range v.input {
    buffer := bytes.NewBuffer(packet.Data)
    switch v.mode {
      case readId:
        v.idHeader.read(buffer)
        v.mode++
        fallthrough

      case readComment:
        // TODO: EOF during this packet is acceptable
        if buffer.Len() == 0 {
          // This could happen if the id and comment headers aren't in the
          // same packet.  The spec really doesn't specify how it should be.
          // TODO: For this pair of headers this might be specified to never
          //       happen, so remove this if statement if that's the case.
          continue
        }
        v.commentHeader.read(buffer)
        v.mode++
        fallthrough

      case readSetup:
        if buffer.Len() == 0 {
          // This could happen if the comment and setup headers aren't in the
          // same packet.  The spec really doesn't specify how it should be.
          continue
        }
        v.setupHeader.read(buffer, int(v.Channels))
        v.mode++
        total = 0

      case readData:
        v.readAudioPacket(buffer, int(v.Channels))
    }
  }
}

