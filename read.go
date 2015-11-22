package keyval

import (
	"errors"
	"io"
)

type readState int

const (
	stateWhitespace readState = iota
	stateCommentWhitespace
	stateComment
	stateCommentOrWhitespace
	stateContinueCommentOrWhitespace
	stateSectionWhitespace
	stateSection
	stateSectionOrWhitespace
	stateKey
	stateKeyOrWhitespace
	stateValueWhitespace
	stateValue
	stateValueOrWhitespace
)

const DefaultBufferSize = 1 << 18

type readEntry struct {
	key     [][]byte
	val     []byte
	comment []byte
}

type KeyVal struct {
	Key     []string
	Val     string
	Comment string
}

type Reader struct {
	BufferSize int
	buffer     []byte
	state      readState
	reader     io.Reader
	eof        bool
	entries    []*readEntry
	current    *readEntry
	escape     bool
	escapeNext bool
	comment    []byte
	section    []byte
	whitespace []byte
}

var EOFRemainder = errors.New("EOF: incomplete remainder data")

func escape(c byte) bool       { return c == '\\' }
func startComment(c byte) bool { return c == ';' || c == '#' }
func openSection(c byte) bool  { return c == '[' }
func closeSection(c byte) bool { return c == ']' }
func startValue(c byte) bool   { return c == '=' }
func whitespace(c byte) bool   { return c == '\t' || c == ' ' }
func newline(c byte) bool      { return c == '\n' || c == '\r' }
func whitespacenl(c byte) bool { return whitespace(c) || newline(c) }

func New(r io.Reader) *Reader {
	return &Reader{
		BufferSize: DefaultBufferSize,
		state:      stateWhitespace,
		reader:     r}
}

func (r *Reader) startComment()           { r.comment = []byte{} }
func (r *Reader) appendComment(c byte)    { r.comment = append(r.comment, c) }
func (r *Reader) startSection()           { r.section = []byte{} }
func (r *Reader) appendSection(c byte)    { r.section = append(r.section, c) }
func (r *Reader) appendWhitespace(c byte) { r.whitespace = append(r.whitespace, c) }
func (r *Reader) clearWhitespace()        { r.whitespace = []byte{} }

func (r *Reader) checkEscape(c byte) bool {
	r.escape = false
	if r.escapeNext {
		r.escape = true
		r.escapeNext = false
	} else if escape(c) {
		r.escapeNext = true
	}

	return r.escapeNext
}

func (r *Reader) newEntry() {
	r.current = &readEntry{comment: r.comment}
	if len(r.section) == 0 {
		r.current.key = [][]byte{[]byte{}}
	} else {
		r.current.key = [][]byte{r.section, []byte{}}
	}
}

func (r *Reader) completeEntry() {
	r.entries = append(r.entries, r.current)
	r.current = nil
}

func (r *Reader) appendKey(c byte) {
	le := r.entries[len(r.entries)-1]
	lki := len(le.key) - 1
	le.key = append(le.key[:lki], append(le.key[lki], c))
}

func (r *Reader) keyWhitespace() {
	le := r.entries[len(r.entries)-1]
	lki := len(le.key) - 1
	le.key = append(le.key[:lki], append(le.key[lki], r.whitespace...))
	r.clearWhitespace()
}

func (r *Reader) appendValue(c byte) {
	last := r.entries[len(r.entries)-1]
	last.val = append(last.val, c)
}

func (r *Reader) valueWhitespace() {
	le := r.entries[len(r.entries)-1]
	le.val = append(le.val, r.whitespace...)
	r.clearWhitespace()
}

func (r *Reader) commentWhitespace() {
	if len(r.comment) > 0 {
		r.comment = append(r.comment, r.whitespace...)
	}

	r.clearWhitespace()
}

func (r *Reader) sectionWhitespace() {
	if len(r.section) > 0 {
		r.section = append(r.section, r.whitespace...)
	}

	r.clearWhitespace()
}

func keysToString(keys [][]byte) []string {
	s := make([]string, len(keys))
	for i, k := range keys {
		s[i] = string(k)
	}

	return s
}

func (r *Reader) fetchEntry() *KeyVal {
	if len(r.entries) == 0 {
		return nil
	}

	var next *readEntry
	next, r.entries = r.entries[0], r.entries[1:]

	return &KeyVal{
		Key:     keysToString(next.key),
		Val:     string(next.val),
		Comment: string(next.comment)}
}

func (r *Reader) hasRemainder() bool {
	switch r.state {
	case
		stateSectionWhitespace,
		stateSection,
		stateSectionOrWhitespace:
		return true
	default:
		return false
	}
}

func (r *Reader) returnEof() (*KeyVal, error) {
	err := io.EOF
	hr := r.hasRemainder()
	if hr {
		err = EOFRemainder
	}

	var last *KeyVal
	if r.current != nil {
		r.completeEntry()
		last = r.fetchEntry()
	} else if len(r.comment) > 0 || (!hr && len(r.section) > 0) {
		r.newEntry()
		r.completeEntry()
		r.comment = nil
		r.section = nil
		last = r.fetchEntry()
	}

	return last, err
}

func (r *Reader) ReadEntry() (*KeyVal, error) {
	if r.reader == nil {
		return nil, nil
	}

	next := r.fetchEntry()
	if next != nil {
		return next, nil
	}

	if r.eof {
		return r.returnEof()
	}

	if len(r.buffer) != r.BufferSize {
		r.buffer = make([]byte, r.BufferSize)
	}

	l, err := r.reader.Read(r.buffer)
	if err != io.EOF {
		return nil, err
	}

	r.eof = err == io.EOF
	if r.eof && l == 0 {
		return r.returnEof()
	}

	for i := 0; i < l; i++ {
		c := r.buffer[i]

		if r.checkEscape(c) {
			continue
		}

		r.appendChar(c)
	}

	next = r.fetchEntry()
	if next != nil {
		return next, nil
	}

	if r.eof {
		return r.returnEof()
	}

	return nil, nil
}

func (r *Reader) appendChar(c byte) {
	switch r.state {
	case stateWhitespace:
		switch {
		case r.escape:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case whitespace(c):
			r.state = stateWhitespace
		case newline(c):
			r.state = stateWhitespace
		case startComment(c):
			r.state = stateCommentWhitespace
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionWhitespace
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueWhitespace
			r.newEntry()
		default:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		}

	case stateCommentWhitespace:
		switch {
		case r.escape:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentWhitespace
		case newline(c):
			r.state = stateContinueCommentOrWhitespace
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case openSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case closeSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case startValue(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		default:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		}

	case stateComment:
		switch {
		case r.escape:
			r.state = stateComment
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentOrWhitespace
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrWhitespace
			r.clearWhitespace()
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateComment
			r.appendComment(c)
		case openSection(c):
			r.state = stateComment
			r.appendComment(c)
		case closeSection(c):
			r.state = stateComment
			r.appendComment(c)
		case startValue(c):
			r.state = stateComment
			r.appendComment(c)
		default:
			r.state = stateComment
			r.appendComment(c)
		}

	case stateCommentOrWhitespace:
		switch {
		case r.escape:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentOrWhitespace
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrWhitespace
			r.clearWhitespace()
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case openSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case closeSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case startValue(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		default:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		}

	case stateContinueCommentOrWhitespace:
		switch {
		case r.escape:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case whitespace(c):
			r.state = stateContinueCommentOrWhitespace
		case newline(c):
			r.state = stateContinueCommentOrWhitespace
		case startComment(c):
			r.state = stateCommentWhitespace
		case openSection(c):
			r.state = stateSectionWhitespace
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueWhitespace
			r.newEntry()
		default:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		}

	case stateSectionWhitespace:
		switch {
		case r.escape:
			r.state = stateSection
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionWhitespace
		case newline(c):
			r.state = stateSectionWhitespace
		case startComment(c):
			r.state = stateSection
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.appendSection(c)
		case closeSection(c):
			r.state = stateWhitespace
		case startValue(c):
			r.state = stateSection
			r.appendSection(c)
		default:
			r.state = stateSection
			r.appendSection(c)
		}

	case stateSection:
		switch {
		case r.escape:
			r.state = stateSection
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionOrWhitespace
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateSectionOrWhitespace
			r.clearWhitespace()
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateSection
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.appendSection(c)
		case closeSection(c):
			r.state = stateWhitespace
		case startValue(c):
			r.state = stateSection
			r.appendSection(c)
		default:
			r.state = stateSection
			r.appendSection(c)
		}

	case stateSectionOrWhitespace:
		switch {
		case r.escape:
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionOrWhitespace
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateSectionOrWhitespace
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case closeSection(c):
			r.state = stateWhitespace
		case startValue(c):
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		default:
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		}

	case stateKey:
		switch {
		case r.escape:
			r.state = stateKey
			r.appendKey(c)
		case whitespace(c):
			r.state = stateKeyOrWhitespace
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateWhitespace
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentWhitespace
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionWhitespace
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueWhitespace
		default:
			r.state = stateKey
			r.appendKey(c)
		}

	case stateKeyOrWhitespace:
		switch {
		case r.escape:
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		case whitespace(c):
			r.state = stateKeyOrWhitespace
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateWhitespace
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentWhitespace
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionWhitespace
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueWhitespace
		default:
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		}

	case stateValueWhitespace:
		switch {
		case r.escape:
			r.state = stateValue
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueWhitespace
		case newline(c):
			r.state = stateWhitespace
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentWhitespace
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionWhitespace
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateValue
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueWhitespace
		default:
			r.state = stateValue
			r.appendValue(c)
		}

	case stateValue:
		switch {
		case r.escape:
			r.state = stateValue
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueOrWhitespace
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateWhitespace
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentWhitespace
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionWhitespace
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateValue
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueWhitespace
		default:
			r.state = stateValue
			r.appendValue(c)
		}

	case stateValueOrWhitespace:
		switch {
		case r.escape:
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueOrWhitespace
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateWhitespace
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentWhitespace
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionWhitespace
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueWhitespace
		default:
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		}
	}
}
