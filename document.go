package keyval

import "io"

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

func (d *Document) truncRange(at, n int) (int, int) {
	if at < 0 {
		at = 0
	}

	if n < 0 {
		n = 0
	}

	l := d.Len()

	if at > l {
		at = l
	}

	if at+n > l {
		n = l - at
	}

	return at, n
}

func (d *Document) KeysOf(key ...string) [][]string {
	var keys [][]string
	for _, e := range d.Entries() {
		if e == nil || len(e.Key) <= len(key) {
			continue
		}

		if KeyEq(key, e.Key[:len(key)]) {
			keys = append(keys, e.Key[len(key):])
		}
	}

	return keys
}

func (d *Document) Keys() [][]string {
	return d.KeysOf()
}

func (d *Document) Len() int {
	return len(d.Entries())
}

func (d *Document) Index(e *Entry) int {
	for i, ei := range d.Entries() {
		if ei == e {
			return i
		}
	}

	return -1
}

func (d *Document) EntryOf(key ...string) (int, *Entry) {
	entries := d.Entries()
	for i := len(entries) - 1; i >= 0; i-- {
		e := entries[i]
		if e != nil && KeyEq(e.Key, key) {
			return i, e
		}
	}

	return -1, nil
}

func (d *Document) EntriesOf(key ...string) []*Entry {
	var entries []*Entry
	for _, e := range d.Entries() {
		if e != nil && KeyEq(e.Key, key) {
			entries = append(entries, e)
		}
	}

	return entries
}

func (d *Document) EntryAt(i int) *Entry {
	if i < 0 || i >= d.Len() {
		return nil
	}

	return d.Entries()[i]
}

func (d *Document) Entries() []*Entry {
	return d.entries
}

func (d *Document) ReplaceEntry(at, n int, e ...*Entry) {
	at, n = d.truncRange(at, n)
	d.entries = append(d.entries[:at], append(e, d.entries[at+n:]...)...)
}

func (d *Document) InsertEntry(at int, e ...*Entry) {
	d.ReplaceEntry(at, 0, e...)
}

func (d *Document) AppendEntry(e ...*Entry) {
	d.InsertEntry(d.Len(), e...)
}

// this makes not much sense this way
func (d *Document) DeleteAt(at, n int, key ...string) {
	at, n = d.truncRange(at, n)
	for i := 0; i < n; {
		e := d.EntryAt(i)
		if e != nil && KeyEq(e.Key, key) {
			d.ReplaceEntry(i, 1)
			n--
		} else {
			i++
		}
	}
}

func (d *Document) DeleteEntry(e ...*Entry) {
	for i := 0; len(e) > 0 && i < d.Len(); {
		if d.EntryAt(i) == e[0] {
			d.ReplaceEntry(i, 1)
			e = e[1:]
		} else {
			i++
		}
	}
}

func (d *Document) ValOf(key ...string) string {
	_, e := d.EntryOf(key...)
	if e == nil {
		return ""
	}

	return e.Val
}

func (d *Document) ValsOf(key ...string) []string {
	var vals []string
	for _, e := range d.EntriesOf(key...) {
		if e != nil {
			vals = append(vals, e.Val)
		}
	}

	return vals
}

func (d *Document) SetValOf(key []string, val string) {
	for _, e := range d.EntriesOf(key...) {
		if e != nil {
			e.Val = val
		}
	}
}

func (d *Document) InsertVal(at int, key []string, val string) {
	d.InsertEntry(at, &Entry{Key: key, Val: val})
}

func (d *Document) AppendVal(key []string, val string) {
	d.InsertVal(d.Len(), key, val)
}

func (d *Document) DeleteOf(key ...string) {
	d.DeleteAt(0, d.Len(), key...)
}

func (d *Document) CommentOf(key ...string) string {
	_, e := d.EntryOf(key...)
	if e == nil {
		return ""
	}

	return e.Comment
}

func (d *Document) SetCommentOf(key []string, comment string) {
	entries := d.EntriesOf(key...)
	for _, e := range entries {
		e.Comment = comment
	}
}

func (d *Document) Val(key string) string {
	return d.ValOf(SplitKey(key)...)
}

func (d *Document) Vals(key string) []string {
	return d.ValsOf(SplitKey(key)...)
}

func (d *Document) SetVal(key, val string) {
	d.SetValOf(SplitKey(key), val)
}

func (d *Document) Insert(at int, key string, val string) {
	d.InsertVal(at, SplitKey(key), val)
}

func (d *Document) Append(key string, val string) {
	d.Insert(d.Len(), key, val)
}

func (d *Document) Delete(key string) {
	d.DeleteOf(SplitKey(key)...)
}

func (d *Document) Comment(key string) string {
	return d.CommentOf(SplitKey(key)...)
}

func (d *Document) SetComment(key string, comment string) {
	d.SetCommentOf(SplitKey(key), comment)
}

// maybe better if this is not here?
// it should be, because it should be the default decode
func (d *Document) Map() map[string][]interface{} {
	return nil
}

func (d *Document) SortFunc(less CompareFunc) {
}

func (d *Document) Sort() {
	d.SortFunc(DefaultCompare)
}

func (d *Document) TruncateStart(n int) {
	d.ReplaceEntry(0, n)
}

func (d *Document) TruncateEnd(n int) {
	d.ReplaceEntry(d.Len()-n, n)
}

func (d *Document) Copy() *Document {
	c := &Document{}
	c.entries = make([]*Entry, len(d.entries))
	copy(c.entries, d.entries)
	return c
}

func (d *Document) TruncateEffective() {
	found := make(map[string]bool)
	for i := d.Len() - 1; i >= 0; i-- {
		e := d.EntryAt(i)
		if e == nil {
			d.ReplaceEntry(i, 1)
			continue
		}

		ks := JoinKey(e.Key)
		if found[ks] {
			d.ReplaceEntry(i, 1)
		} else {
			found[ks] = true
		}
	}
}

func (d *Document) Reset() {
	d.ReplaceEntry(0, d.Len())
}

func (d *Document) Bytes() []byte {
	return nil
}

func (d *Document) String() string {
	return ""
}

func (d *Document) Json() []byte {
	return nil
}

func (d *Document) Yaml() []byte {
	return nil
}

func (d *Document) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func (d *Document) MarshalYAML() ([]byte, error) {
	return nil, nil
}

func (d *Document) ReadAllEntries(r *EntryReader) error {
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

		d.AppendEntry(entry)

		if err != nil {
			return err
		}
	}
}

func (d *Document) ReadAll(r io.Reader) error {
	return d.ReadAllEntries(NewEntryReader(r))
}

func (d *Document) WriteAllEntries(w *EntryWriter) error {
	if w == nil {
		return nil
	}

	for _, e := range d.Entries() {
		if err := w.WriteEntry(e); err != nil {
			return err
		}
	}

	return nil
}

func (d *Document) WriteAll(w io.Writer) error {
	return d.WriteAllEntries(NewEntryWriter(w))
}
