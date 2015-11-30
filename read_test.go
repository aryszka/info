package keyval

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type infiniteBuffer struct {
	reader io.Reader
}

type errReader struct{ readCount int }

var (
	errExpectedFailingRead   = errors.New("expected failing read")
	errUnexpectedFailingRead = errors.New("unexpected failing read")
)

func (er *errReader) Read(b []byte) (int, error) {
	er.readCount++
	if er.readCount <= 1 {
		return 0, errExpectedFailingRead
	}

	return 0, errUnexpectedFailingRead
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

func TestReturnsSameErrorOnReadRepeatedCall(t *testing.T) {
	ir := &errReader{}
	r := NewReader(ir)
	var err error

	_, err = r.ReadEntry()
	if err != errExpectedFailingRead {
		t.Error("failed to fail")
	}

	_, err = r.ReadEntry()
	if err != errExpectedFailingRead || ir.readCount != 1 {
		t.Error("failed to store previous failure")
	}
}
