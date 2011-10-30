package vorbis

import (
  "fmt"
  "ogg"
  "bytes"
  "os"
)

var magic_string string = "\x01vorbis"

func check(err os.Error) {
  if err != nil {
    panic(err.String())
  }
}

func (v *vorbisDecoder) readAudioPacket(page ogg.Page) {
  br := MakeBitReader(v.buffer)
  if br.ReadBits(1) != 0 {
    fmt.Printf("Warning: Not an audio packet")
  }
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
      v.readAudioPacket(page)
      v.buffer.Truncate(0)
  }
}
func (v *vorbisDecoder) Finish() {
  fmt.Printf("Finished stream\n")
}


