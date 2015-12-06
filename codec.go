package keyval

import "io"

type Encoder struct{}

type Decoder struct{}

func NewEncoder(w io.Writer) *Encoder         { return nil }
func (e *Encoder) Encode(v interface{}) error { return nil }
func NewDecoder(r io.Reader) *Decoder         { return nil }
func (d *Decoder) Decode(v interface{}) error { return nil }
