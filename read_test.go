package keyval

import (
	"bytes"
	"io"
	"testing"
)

type infiniteBuffer struct {
	reader io.Reader
}

func (b *infiniteBuffer) Read(data []byte) (int, error) {
	l, err := b.reader.Read(data)
	if err == io.EOF {
		err = nil
	}

	return l, err
}

func TestNothingToRead(t *testing.T) {
	r := NewReader(nil)
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read")
	}
}

func TestEmptyReader(t *testing.T) {
	r := NewReader(&infiniteBuffer{bytes.NewBuffer(nil)})
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read")
	}
}

func TestEmptyReaderEof(t *testing.T) {
	r := NewReader(bytes.NewBuffer(nil))
	e, err := r.ReadEntry()
	if e != nil || err != io.EOF {
		t.Error("failed to read eof")
	}
}

func TestSmallBufferSize(t *testing.T) {
	buf := bytes.NewBufferString(`
        just a long piece of text\
        spanning multiple lines,\
        more lines and bytes than\
        the reader buffer size`)
	r := NewReader(buf)
	r.BufferSize = 2
	for {
		entry, err := r.ReadEntry()
		if err != nil && err != io.EOF {
			t.Error(err)
			return
		}

		if entry == nil && err != io.EOF {
			t.Error("failed to read long entry")
			return
		}

		if err == io.EOF {
			return
		}
	}
}
