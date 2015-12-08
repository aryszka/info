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
	buf := &Document{}
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
	buf := &Document{}
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
	buf := &Document{}
	if err := buf.ReadAll(rbuf); err != nil && err != io.EOF {
		t.Error(err)
		return
	}

	ew := &errWriter{}
	if err := buf.WriteAll(ew); err != errExpectedFailingWrite {
		t.Error("failed to fail", err)
	}
}

func TestWriteAll(t *testing.T) {
	buf := &Document{}
	buf.AppendEntry(&Entry{Key: []string{"section1", "key1"}, Val: "val1"})
	buf.AppendEntry(&Entry{Key: []string{"section1", "key2"}, Val: "val2"})
	buf.AppendEntry(&Entry{Key: []string{"section2", "key3"}, Val: "val3"})
	buf.AppendEntry(&Entry{Key: []string{"section2", "key4"}, Val: "val4"})
	ibuf := bytes.NewBuffer(nil)
	if err := buf.WriteAll(ibuf); err != nil {
		t.Error(err)
		return
	}

	if ibuf.String() != "[section1]\nkey1 = val1\nkey2 = val2\n\n"+
		"[section2]\nkey3 = val3\nkey4 = val4\n" {
		t.Error("failed to write all")
	}
}

func TestGetKeys(t *testing.T) {
	buf := &Document{}
	buf.AppendEntry(&Entry{Key: []string{"some key"}})
	buf.AppendEntry(&Entry{Key: []string{"some key", "some sub key"}})
	buf.AppendEntry(&Entry{Key: []string{"some key", "some sub key"}})

	keys := buf.Keys()

	if len(keys) != 2 {
		t.Error("invalid number of keys")
	}

	if len(keys[0]) != 1 || keys[0][0] != "some key" {
		t.Error("invalid key")
	}

	if len(keys[1]) != 2 || keys[1][0] != "some key" || keys[1][1] != "some sub key" {
		t.Error("invalid key")
	}
}

func TestKeysCloned(t *testing.T) {
	buf := &Document{}
	buf.AppendEntry(&Entry{Key: []string{"some key"}})
	buf.AppendEntry(&Entry{Key: []string{"some key", "some sub key"}})
	buf.AppendEntry(&Entry{Key: []string{"some key", "some sub key"}})

	keys := buf.Keys()
	keys[0] = nil
	keys = buf.Keys()
	if len(keys[0]) != 1 || keys[0][0] != "some key" {
		t.Error("keys not cloned")
	}
}

func TestGetEntries(t *testing.T) {
	entries := []*Entry{
		{Key: []string{"some key"}},
		{Key: []string{"some key", "some sub key"}},
		{Key: []string{"some key", "some sub key"}}}
	buf := &Document{}
	buf.AppendEntry(entries...)

	entriesBack := buf.Entries()

	if len(entriesBack) != len(entries) {
		t.Error("failed to return entries")
	}

	for i, e := range entriesBack {
		if e != entries[i] {
			t.Error("failed to return entries")
		}
	}
}

func TestEntrySliceCloned(t *testing.T) {
	buf := &Document{}
	buf.AppendEntry(&Entry{Key: []string{"key1"}, Val: "val1"})
	buf.AppendEntry(&Entry{Key: []string{"key2"}, Val: "val2"})
	entries1 := buf.Entries()
	entries2 := buf.Entries()
	entries1[0] = nil
	if entries2[0] == nil {
		t.Error("failed copy entry slice")
	}
}

func TestAppendEntry(t *testing.T) {
	buf := &Document{}
	entry := &Entry{Key: []string{"key"}}
	buf.AppendEntry(entry)
	if buf.Entry(0) != entry {
		t.Error("failed to append entry")
	}
}
