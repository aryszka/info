package keyval

import (
	"bytes"
	"io"
	"testing"
)

type infiniteBuffer struct {
	buffer *bytes.Buffer
}

func (b *infiniteBuffer) Read(data []byte) (int, error) {
	l, err := b.buffer.Read(data)
	if err == io.EOF {
		err = nil
	}

	return l, err
}

func TestNothingToRead(t *testing.T) {
	r := New(nil)
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read")
	}
}

func TestEmptyReader(t *testing.T) {
	r := New(&infiniteBuffer{bytes.NewBuffer(nil)})
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read")
	}
}

func TestEmptyReaderEof(t *testing.T) {
	r := New(bytes.NewBuffer(nil))
	e, err := r.ReadEntry()
	if e != nil || err != io.EOF {
		t.Error("failed to read eof")
	}
}
