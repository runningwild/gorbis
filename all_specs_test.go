package ogg_test

import (
  "gospec"
  "testing"
)


func TestAllSpecs(t *testing.T) {
  r := gospec.NewRunner()
  r.AddSpec(OggSpec)
  gospec.MainGoTest(r, t)
}

