package keyval

import (
	"io"
	"strings"
)

// ???
// rather no
type MapOptions int

const (
	NoList MapOptions = iota
	ListAll
	ListBreadthFirst
	ListDepthFirst
)

type Document struct{ entries []*Entry }

type CompareFunc func(*Entry, *Entry) bool

func DefaultCompare(left, right *Entry) bool { return false }

func JoinKey(key []string) string {
	return strings.Join(key, ".")
}

func SplitKey(key string) []string {
	return strings.Split(key, ".")
}

func KeyEq(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for i, k := range left {
		if k != right[i] {
			return false
		}
	}

	return true
}

func (b *Document) truncRange(at, n int) (int, int) {
	if at < 0 {
		at = 0
	}

	if n < 0 {
		n = 0
	}

	l := b.Len()

	if at > l {
		at = l
	}

	if at+n > l {
		n = l - at
	}

	return at, n
}

func (b *Document) KeysOf(key ...string) [][]string {
	var keys [][]string
	for _, e := range b.Entries() {
		if e == nil || len(e.Key) <= len(key) {
			continue
		}

		if KeyEq(key, e.Key[:len(key)]) {
			keys = append(keys, e.Key[len(key):])
		}
	}

	return keys
}

func (b *Document) Keys() [][]string {
	return b.KeysOf()
}

func (b *Document) Len() int {
	return len(b.Entries())
}

func (b *Document) Index(e *Entry) int {
	for i, ei := range b.Entries() {
		if ei == e {
			return i
		}
	}

	return -1
}

func (b *Document) EntryOf(key ...string) (int, *Entry) {
	entries := b.Entries()
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e != nil && KeyEq(e.Key, key) {
			return i, e
		}
	}

	return -1, nil
}

func (b *Document) EntriesOf(key ...string) []*Entry {
	var entries []*Entry
	for _, e := range b.Entries() {
		if e != nil && KeyEq(e.Key, key) {
			entries = append(entries, e)
		}
	}

	return entries
}

func (b *Document) EntryAt(i int) *Entry {
	if i < 0 || i >= b.Len() {
		return nil
	}

	return b.Entries()[i]
}

func (b *Document) Entries() []*Entry {
	return b.entries
}

func (b *Document) ReplaceEntry(at, n int, e ...*Entry) {
	at, n = b.truncRange(at, n)
	b.entries = append(b.entries[:at], append(e, b.entries[at+n:]...)...)
}

func (b *Document) InsertEntry(at int, e ...*Entry) {
	b.ReplaceEntry(at, 0, e...)
}

func (b *Document) AppendEntry(e ...*Entry) {
	b.InsertEntry(b.Len(), e...)
}

func (b *Document) DeleteAt(at, n int, key ...string) {
	at, n = b.truncRange(at, n)
	for i := 0; i < n; {
		e := b.EntryAt(i)
		if e != nil && KeyEq(e.Key, key) {
			b.ReplaceEntry(i, 1)
			n--
		} else {
			i++
		}
	}
}

func (b *Document) DeleteEntry(e ...*Entry) {
	for i := 0; len(e) > 0 && i < b.Len(); {
		if b.EntryAt(i) == e[0] {
			b.ReplaceEntry(i, 1)
			e = e[1:]
		} else {
			i++
		}
	}
}

func (b *Document) ValOf(key ...string) string {
	_, e := b.EntryOf(key...)
	if e == nil {
		return ""
	}

	return e.Val
}

func (b *Document) ValsOf(key ...string) []string {
	var vals []string
	for _, e := range b.EntriesOf(key...) {
		if e != nil {
			vals = append(vals, e.Val)
		}
	}

	return vals
}

func (b *Document) SetValOf(key []string, val string) {
	for _, e := range b.EntriesOf(key...) {
		if e != nil {
			e.Val = val
		}
	}
}

func (b *Document) InsertVal(at int, key []string, val string) {
	b.InsertEntry(at, &Entry{Key: key, Val: val})
}

func (b *Document) AppendVal(key []string, val string) {
	b.InsertVal(b.Len(), key, val)
}

func (b *Document) DeleteOf(key ...string) {
	b.DeleteAt(0, b.Len(), key...)
}

func (b *Document) CommentOf(key ...string) string {
	_, e := b.EntryOf(key...)
	if e == nil {
		return ""
	}

	return e.Comment
}

func (b *Document) SetCommentOf(key []string, comment string) {
	entries := b.EntriesOf(key...)
	for _, e := range entries {
		e.Comment = comment
	}
}

func (b *Document) Val(key string) string {
	return b.ValOf(SplitKey(key)...)
}

func (b *Document) Vals(key string) []string {
	return b.ValsOf(SplitKey(key)...)
}

func (b *Document) SetVal(key, val string) {
	b.SetValOf(SplitKey(key), val)
}

func (b *Document) Insert(at int, key string, val string) {
	b.InsertVal(at, SplitKey(key), val)
}

func (b *Document) Append(key string, val string) {
	b.Insert(b.Len(), key, val)
}

func (b *Document) Delete(key string) {
	b.DeleteOf(SplitKey(key)...)
}

func (b *Document) Comment(key string) string {
	return b.CommentOf(SplitKey(key)...)
}

func (b *Document) SetComment(key string, comment string) {
	b.SetCommentOf(SplitKey(key), comment)
}

// maybe better if this is not here?
// it should be, because it should be the default decode
func (b *Document) Map() map[string][]interface{} { return nil }
func (b *Document) SortFunc(less CompareFunc)     {}

func (b *Document) Sort() {
	b.SortFunc(DefaultCompare)
}

func (b *Document) TruncateStart(n int) {
	b.ReplaceEntry(0, n)
}

func (b *Document) TruncateEnd(n int) {
	b.ReplaceEntry(b.Len()-n, n)
}

func (b *Document) Copy() *Document              { return nil }
func (b *Document) TruncateEffective()           {}
func (b *Document) Reset()                       { b.ReplaceEntry(0, b.Len()) }
func (b *Document) Bytes() []byte                { return nil }
func (b *Document) String() string               { return "" }
func (b *Document) Json() []byte                 { return nil }
func (b *Document) Yaml() []byte                 { return nil }
func (b *Document) MarshalJSON() ([]byte, error) { return nil, nil }
func (b *Document) MarshalYAML() ([]byte, error) { return nil, nil }

func (b *Document) ReadAllEntries(r *EntryReader) error {
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

func (b *Document) ReadAll(r io.Reader) error {
	return b.ReadAllEntries(NewEntryReader(r))
}

func (b *Document) WriteAllEntries(w *EntryWriter) error {
	if w == nil {
		return nil
	}

	for _, e := range b.Entries() {
		if err := w.WriteEntry(e); err != nil {
			return err
		}
	}

	return nil
}

func (b *Document) WriteAll(w io.Writer) error {
	return b.WriteAllEntries(NewEntryWriter(w))
}
