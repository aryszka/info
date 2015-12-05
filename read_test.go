package keyval

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

type (
	infiniteBuffer struct{ reader io.Reader }
	errReader      struct{ readCount int }
	measureReader  struct{ lastReadSize int }
)

var (
	errExpectedFailingRead   = errors.New("expected failing read")
	errUnexpectedFailingRead = errors.New("unexpected failing read")
)

func (b *infiniteBuffer) Read(data []byte) (int, error) {
	l, err := b.reader.Read(data)
	if err == io.EOF {
		err = nil
	}

	return l, err
}

func (er *errReader) Read(b []byte) (int, error) {
	er.readCount++
	if er.readCount <= 1 {
		return 0, errExpectedFailingRead
	}

	return 0, errUnexpectedFailingRead
}

func (mr *measureReader) Read(b []byte) (int, error) {
	mr.lastReadSize = len(b)
	return 0, nil
}

func TestNothingToRead(t *testing.T) {
	r := NewEntryReader(nil)
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read")
	}
}

func TestEmptyReader(t *testing.T) {
	r := NewEntryReader(&infiniteBuffer{bytes.NewBuffer(nil)})
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read", e == nil, err == nil)
	}
}

func TestEmptyReaderEof(t *testing.T) {
	r := NewEntryReader(bytes.NewBuffer(nil))
	e, err := r.ReadEntry()
	if e != nil || err != io.EOF {
		t.Error("failed to read eof")
	}
}

func TestReturnsSameErrorOnReadRepeatedCall(t *testing.T) {
	ir := &errReader{}
	r := NewEntryReader(ir)
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

func TestUninitializedReader(t *testing.T) {
	r := &EntryReader{}
	entry, err := r.ReadEntry()
	if entry != nil || err != nil {
		t.Error("failed noop")
	}
}

func TestReadMultipleAndEof(t *testing.T) {
	buf := bytes.NewBufferString(`
        key one = value one
        key two = value two
        key three = value three
        `)
	r := NewEntryReader(buf)

	var (
		entry *Entry
		err   error
	)

	entry, err = r.ReadEntry()
	if err != nil || entry == nil || strings.Join(entry.Key, ".") != "key one" || entry.Val != "value one" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != nil || entry == nil || strings.Join(entry.Key, ".") != "key two" || entry.Val != "value two" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != nil && err != io.EOF ||
		entry == nil || strings.Join(entry.Key, ".") != "key three" || entry.Val != "value three" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != io.EOF || entry != nil {
		t.Error("failed to read eof")
	}
}

func TestReadMultipleAndHang(t *testing.T) {
	buf := &infiniteBuffer{bytes.NewBufferString(`
        key one = value one
        key two = value two
        key three = value three
        `)}
	r := NewEntryReader(buf)

	var (
		entry *Entry
		err   error
	)

	entry, err = r.ReadEntry()
	if err != nil || entry == nil || strings.Join(entry.Key, ".") != "key one" || entry.Val != "value one" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != nil || entry == nil || strings.Join(entry.Key, ".") != "key two" || entry.Val != "value two" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != nil || entry == nil || strings.Join(entry.Key, ".") != "key three" || entry.Val != "value three" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != nil || entry != nil {
		t.Error("failed to hang")
	}
}

func TestReadMultipleAndHangIncomplete(t *testing.T) {
	buf := &infiniteBuffer{bytes.NewBufferString(`
        key one = value one
        key two = value two
        key three = value three`)}
	r := NewEntryReader(buf)

	var (
		entry *Entry
		err   error
	)

	entry, err = r.ReadEntry()
	if err != nil || entry == nil || strings.Join(entry.Key, ".") != "key one" || entry.Val != "value one" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != nil || entry == nil || strings.Join(entry.Key, ".") != "key two" || entry.Val != "value two" {
		t.Error("failed to read")
	}

	entry, err = r.ReadEntry()
	if err != nil || entry != nil {
		t.Error("failed to hang")
	}
}
