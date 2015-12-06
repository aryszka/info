package keyval

import (
	"bytes"
	"io"
	"testing"
)

const testDoc = `
    # example keyval document
    [section1]
    key1 = val1
    key2 = val2
    [section2]
    key3 = val3
    key4 = val4`

func TestReadAllError(t *testing.T) {
	ibuf := bytes.NewBufferString(`
        # example keyval document
        [section1]
        key1 = val1
        key2 = val2
        [sectio`)
	buf := &Buffer{}
	err := buf.ReadAll(ibuf)
	if err != EOFIncomplete {
		t.Error(err)
	}
}

func TestReadAll(t *testing.T) {
	var rentries []*Entry
	rbuf := bytes.NewBufferString(testDoc)
	r := NewEntryReader(rbuf)
	for {
		entry, err := r.ReadEntry()

		if err != nil && err != io.EOF {
			t.Error(err)
			return
		}

		if entry == nil {
			break
		}

		rentries = append(rentries, entry)
	}

	ibuf := bytes.NewBufferString(testDoc)
	buf := &Buffer{}
	err := buf.ReadAll(ibuf)
	if err != nil && err != io.EOF {
		t.Error(err)
	}

	entries := buf.Entries()
	if len(entries) != len(rentries) {
		t.Error("failed to read all entries", len(entries), len(rentries))
	}

	for i, e := range entries {
		if e.Comment != rentries[i].Comment {
			t.Error(i, "failed to read comment", e.Comment, rentries[i].Comment)
			return
		}

		if len(e.Key) != len(rentries[i].Key) {
			t.Error(i, "failed to read keys", len(e.Key), len(rentries[i].Key))
			return
		}

		for j, k := range e.Key {
			if k != rentries[i].Key[j] {
				t.Error(i, j, "failed to read keys", k, rentries[i].Key[j])
				return
			}
		}

		if e.Val != rentries[i].Val {
			t.Error(i, "failed to read vals", e.Val, rentries[i].Val)
			return
		}
	}
}

func TestWriteAllError(t *testing.T) {
	rbuf := bytes.NewBufferString(testDoc)
	buf := &Buffer{}
	if err := buf.ReadAll(rbuf); err != nil && err != io.EOF {
		t.Error(err)
		return
	}

	ew := &errWriter{}
	if err := buf.WriteAll(ew); err != errExpectedFailingWrite {
		t.Error("failed to fail", err)
	}
}

func TestEntrySliceCloned(t *testing.T) {
	buf := &Buffer{}
	buf.AppendEntry(&Entry{Key: []string{"key1"}, Val: "val1"})
	buf.AppendEntry(&Entry{Key: []string{"key2"}, Val: "val2"})
	entries1 := buf.Entries()
	entries2 := buf.Entries()
	entries1[0] = nil
	if entries2[0] == nil {
		t.Error("failed copy entry slice")
	}
}
