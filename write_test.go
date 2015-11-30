package keyval

import (
	"errors"
	"testing"
)

type errWriter struct{ writeCount int }

var (
	errExpectedFailingWrite   = errors.New("expected failing write")
	errUnexpectedFailingWrite = errors.New("unexpected failing write")
)

func (er *errWriter) Write(b []byte) (int, error) {
	er.writeCount++
	if er.writeCount <= 1 {
		return 0, errExpectedFailingRead
	}

	return 0, errUnexpectedFailingRead
}

func TestEscapeWrite(t *testing.T) {
	for i, ti := range []struct{ escaped, in, out string }{
		{"", "abc", "abc"},
		{"a", "abc", "\\abc"},
		{"b", "abc", "a\\bc"},
		{"c", "abc", "ab\\c"},
		{"ab", "abc", "\\a\\bc"},
		{"ac", "abc", "\\ab\\c"},
		{"bc", "abc", "a\\b\\c"},
		{"abc", "abc", "\\a\\b\\c"},
		{" \n", "some longer text with\nnew line ",
			"some\\ longer\\ text\\ with\\\nnew\\ line\\ "},
	} {
		out := string(escapeWrite([]byte(ti.in), []byte(ti.escaped)))
		if out != ti.out {
			t.Error(i, ti.escaped, ti.in, ti.out, out)
		}
	}
}

func TestReturnSameErrorOnRepeatedWriteCall(t *testing.T) {
	iw := &errWriter{}
	w := NewWriter(iw)
	var err error

	err = w.WriteEntry(&Entry{Key: []string{"a key"}})
	if err != errExpectedFailingRead {
		t.Error("failed to fail")
	}

	err = w.WriteEntry(&Entry{Key: []string{"a key"}})
	if err != errExpectedFailingRead || iw.writeCount != 1 {
		t.Error("failed to store previous failure")
	}
}

func TestReturnSameErrorOnRepeatedWriteCallBuffered(t *testing.T) {
	iw := &errWriter{}
	w := NewWriter(iw)
	w.BufferSize = 1 << 2
	var err error

	err = w.WriteEntry(&Entry{Key: []string{"a key"}})
	if err != errExpectedFailingRead {
		t.Error("failed to fail")
	}

	err = w.WriteEntry(&Entry{Key: []string{"a key"}})
	if err != errExpectedFailingRead || iw.writeCount != 1 {
		t.Error("failed to store previous failure")
	}

	err = w.Flush()
	if err != errExpectedFailingRead || iw.writeCount != 1 {
		t.Error("failed to store previous failure")
	}
}
