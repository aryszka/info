package keyval

import (
	"bufio"
	"errors"
	"io"
)

const InnerReadBufferSize = 1 << 9

type readEntry struct {
	comment []byte
	section [][]byte
	key     [][]byte
	val     []byte
}

type ByteReader interface {
	ReadByte() (byte, error)
}

type EntryReader struct {
	reader         ByteReader
	state          readState
	entries        []*readEntry
	escape         bool
	escapeNext     bool
	comment        []byte
	section        [][]byte
	currentSection []byte
	key            [][]byte
	currentKey     []byte
	val            []byte
	whitespace     []byte
	commentApplied bool
	sectionApplied bool
	err            error
}

var EOFIncomplete = errors.New("EOF: incomplete data")

func escape(c byte) bool       { return c == EscapeChar }
func startComment(c byte) bool { return c == CommentChar }
func openSection(c byte) bool  { return c == OpenSectionChar }
func closeSection(c byte) bool { return c == CloseSectionChar }
func startValue(c byte) bool   { return c == StartValueChar }
func keySeparator(c byte) bool { return c == KeySeparatorChar }
func whitespace(c byte) bool   { return c == SpaceChar || c == TabChar }
func newline(c byte) bool      { return c == NewlineChar }

func NewEntryReader(r io.Reader) *EntryReader {
	if r == nil {
		return &EntryReader{}
	}

	er := &EntryReader{state: stateInitial}

	br, ok := r.(ByteReader)
	if !ok {
		br = bufio.NewReaderSize(r, InnerReadBufferSize)
	}

	er.reader = br
	return er
}

func (r *EntryReader) checkEscape(c byte) bool {
	r.escape = false
	if r.escapeNext {
		r.escape = true
		r.escapeNext = false
	} else if escape(c) {
		r.escapeNext = true
	}

	return r.escapeNext
}

func (r *EntryReader) appendWhitespace(c byte) { r.whitespace = append(r.whitespace, c) }
func (r *EntryReader) clearWhitespace()        { r.whitespace = nil }

func (r *EntryReader) clearComment()      { r.comment = nil }
func (r *EntryReader) commentWhitespace() { r.comment = append(r.comment, r.whitespace...) }

func (r *EntryReader) appendComment(c byte) {
	r.comment = append(r.comment, c)
	r.commentApplied = false
}

func (r *EntryReader) clearSection() {
	if !r.sectionApplied && len(r.section) > 0 {
		r.completeEntry()
	}

	r.section = nil
}

func (r *EntryReader) appendSection(c byte) { r.currentSection = append(r.currentSection, c) }
func (r *EntryReader) sectionWhitespace() {
	r.currentSection = append(r.currentSection, r.whitespace...)
}

func (r *EntryReader) completeSection() {
	if len(r.currentSection) > 0 {
		r.section = append(r.section, r.currentSection)
	}

	r.currentSection = nil
	r.sectionApplied = false
}

func (r *EntryReader) appendKey(c byte) { r.currentKey = append(r.currentKey, c) }

func (r *EntryReader) completeKey() {
	r.key = append(r.key, r.currentKey)
	r.currentKey = nil
}

func (r *EntryReader) keyWhitespace() {
	if len(r.currentKey) > 0 {
		r.currentKey = append(r.currentKey, r.whitespace...)
	}
}

func (r *EntryReader) appendValue(c byte) { r.val = append(r.val, c) }
func (r *EntryReader) valueWhitespace()   { r.val = append(r.val, r.whitespace...) }

func (r *EntryReader) completeEntry() {
	r.entries = append(r.entries, &readEntry{
		comment: r.comment,
		section: r.section,
		key:     r.key,
		val:     r.val})
	r.key = nil
	r.val = nil
	r.commentApplied = true
	r.sectionApplied = true
}

func mergeKey(section, key [][]byte) []string {
	skey := make([]string, len(section)+len(key))
	for i, k := range append(section, key...) {
		skey[i] = string(k)
	}

	return skey
}

func (r *EntryReader) fetchEntry() *Entry {
	if len(r.entries) == 0 {
		return nil
	}

	var next *readEntry
	next, r.entries = r.entries[0], r.entries[1:]

	return &Entry{
		Key:     mergeKey(next.section, next.key),
		Val:     string(next.val),
		Comment: string(next.comment)}
}

func (r *EntryReader) hasRemainderSection() bool {
	switch r.state {
	case
		stateSectionInitial,
		stateSection,
		stateSectionOrElse:
		return true
	default:
		return false
	}
}

func (r *EntryReader) hasIncompleteEntry() bool {
	return len(r.currentKey) > 0 ||
		len(r.key) > 0 ||
		len(r.val) > 0 ||
		(!r.commentApplied && len(r.comment) > 0) ||
		(!r.sectionApplied && len(r.section) > 0)
}

func (r *EntryReader) eofResult() (*Entry, error) {
	err := io.EOF
	hrs := r.hasRemainderSection()
	if hrs || r.escapeNext {
		err = EOFIncomplete
	}

	var last *Entry
	if r.hasIncompleteEntry() {
		if len(r.currentKey) > 0 {
			r.completeKey()
		}

		r.completeEntry()
		last = r.fetchEntry()
	}

	return last, err
}

func (r *EntryReader) ReadEntry() (*Entry, error) {
	if r.reader == nil {
		return nil, nil
	}

	if r.err != nil && r.err != io.EOF {
		return nil, r.err
	}

	next := r.fetchEntry()
	if next != nil {
		return next, nil
	}

	if r.err == io.EOF {
		return r.eofResult()
	}

	for {
		var c byte
		c, r.err = r.reader.ReadByte()

		if r.err != nil && r.err != io.EOF && r.err != io.ErrNoProgress {
			return nil, r.err
		}

		if r.err == io.EOF {
			return r.eofResult()
		}

		if r.err == io.ErrNoProgress {
			r.err = nil
			return nil, io.ErrNoProgress
		}

		r.acceptChar(c)
		next := r.fetchEntry()
		if next != nil {
			return next, nil
		}
	}
}
