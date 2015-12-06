package keyval

import (
	"io"
	"strings"
)

type Buffer struct {
	entries  []*Entry
	entryMap map[string][]*Entry
	keys     [][]string
}

type CompareFunc func(*Entry, *Entry) bool

func DefaultCompare(left, right *Entry) bool { return false }

func joinKey(key []string) string {
	return strings.Join(key, ".")
}

func (b *Buffer) read(r *EntryReader) error {
	if r == nil {
		return nil
	}

	for {
		entry, err := r.ReadEntry()
		if err != nil && err != io.EOF {
			return err
		}

		if entry == nil {
			return err
		}

		b.entries = append(b.entries, entry)

		if err != nil {
			return err
		}
	}
}

func (b *Buffer) readCached(r *EntryReader) error {
	if r == nil {
		return nil
	}

	for {
		entry, err := r.ReadEntry()
		if err != nil && err != io.EOF {
			return err
		}

		if entry == nil {
			return err
		}

		b.AppendEntry(entry)

		if err != nil {
			return err
		}
	}
}

func (b *Buffer) ReadAllEntry(r *EntryReader) error {
	return b.readCached(r)
}

func (b *Buffer) ReadAll(r io.Reader) error {
	return b.ReadAllEntry(NewEntryReader(r))
}

func (b *Buffer) WriteAllEntry(w *EntryWriter) error {
	if w == nil {
		return nil
	}

	for _, e := range b.entries {
		if err := w.WriteEntry(e); err != nil {
			return err
		}
	}

	return nil
}

func (b *Buffer) WriteAll(w io.Writer) error {
	return b.WriteAllEntry(NewEntryWriter(w))
}

func (b *Buffer) lookupString(key []string) []*Entry {
	keystr := joinKey(key)
	var vals []*Entry
	for _, e := range b.entries {
		if len(e.Key) != len(key) {
			continue
		}

		if e.Val != "" && joinKey(e.Key) == keystr {
			vals = append(vals, e)
		}
	}

	return vals
}

func (b *Buffer) lookupSlice(key []string) []*Entry {
	var vals []*Entry
	for _, e := range b.entries {
		if e.Val == "" {
			continue
		}

		if len(e.Key) != len(key) {
			continue
		}

		keyeq := true
		for i, kn := range key {
			if kn != e.Key[i] {
				keyeq = false
				break
			}
		}

		if keyeq {
			vals = append(vals, e)
		}
	}

	return vals
}

func (b *Buffer) lookupMap(key []string) []*Entry {
	return b.entryMap[joinKey(key)]
}

func (b *Buffer) Get(key string) string                      { return "" }
func (b *Buffer) Set(key, val string)                        {}
func (b *Buffer) GetAll(key []string) []string               { return nil }
func (b *Buffer) SetAll(key, vals []string)                  {}
func (b *Buffer) GetFirst(key []string) string               { return "" }
func (b *Buffer) SetFirst(key []string, val string)          {}
func (b *Buffer) Add(key []string, val string)               {}
func (b *Buffer) CommentAll(key []string) []string           { return nil }
func (b *Buffer) SetCommentAll(key []string, comment string) {}

func (b *Buffer) Keys() [][]string {
	keys := make([][]string, len(b.keys))
	copy(keys, b.keys)
	return keys
}

func (b *Buffer) KeysOf(key []string) [][]string { return nil }

func (b *Buffer) Len() int {
	return len(b.entries)
}

func (b *Buffer) Entry(i int) *Entry { return nil }

func (b *Buffer) Entries() []*Entry {
	ec := make([]*Entry, b.Len())
	copy(ec, b.entries)
	return ec
}

func (b *Buffer) AppendEntry(e ...*Entry) {
	if b.entryMap == nil {
		b.entryMap = make(map[string][]*Entry)
	}

	b.entries = append(b.entries, e...)
	for _, ei := range e {
		keystr := joinKey(ei.Key)
		entries, hasKey := b.entryMap[keystr]
		if !hasKey {
			b.keys = append(b.keys, ei.Key)
		}

		b.entryMap[keystr] = append(entries, ei)
	}
}

func (b *Buffer) Map() map[string]interface{}     { return nil }
func (b *Buffer) AddMap(m map[string]interface{}) {}
func (b *Buffer) SortFunc(less CompareFunc)       {}

func (b *Buffer) Sort() {
	b.SortFunc(DefaultCompare)
}

func (b *Buffer) Truncate(n int)               {}
func (b *Buffer) Reset()                       { b.Truncate(0) }
func (b *Buffer) Bytes() []byte                { return nil }
func (b *Buffer) String() string               { return "" }
func (b *Buffer) Json() []byte                 { return nil }
func (b *Buffer) MarshalJSON() ([]byte, error) { return nil, nil }
