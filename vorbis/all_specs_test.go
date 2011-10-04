package vorbis_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
//  r.AddSpec(BitReaderSpec)
  r.AddSpec(Lookup1Spec)
  gospec.MainGoTest(r, t)
}

