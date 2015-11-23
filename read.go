package keyval

import (
	"errors"
	"io"
	"strings"
)

type readState int

const (
	stateInitial readState = iota
	stateCommentInitial
	stateComment
	stateCommentOrElse
	stateContinueCommentOrElse
	stateSectionInitial
	stateSection
	stateSectionOrElse
	stateKey
	stateKeyOrElse
	stateValueInitial
	stateValue
	stateValueOrElse
)

const DefaultBufferSize = 1 << 18

type readEntry struct {
	comment []byte
	section []byte
	key     []byte
	val     []byte
}

type Entry struct {
	Key     []string
	Val     string
	Comment string
}

type Reader struct {
	BufferSize     int
	reader         io.Reader
	buffer         []byte
	state          readState
	entries        []*readEntry
	current        *readEntry
	escape         bool
	escapeNext     bool
	comment        []byte
	section        []byte
	whitespace     []byte
	commentApplied bool
	sectionApplied bool
	eof            bool
}

var EOFRemainder = errors.New("EOF: incomplete remainder data")

func escape(c byte) bool       { return c == '\\' }
func startComment(c byte) bool { return c == ';' || c == '#' }
func openSection(c byte) bool  { return c == '[' }
func closeSection(c byte) bool { return c == ']' }
func startValue(c byte) bool   { return c == '=' }
func whitespace(c byte) bool   { return c == '\t' || c == ' ' }
func newline(c byte) bool      { return c == '\n' || c == '\r' }

func NewReader(r io.Reader) *Reader {
	return &Reader{
		BufferSize: DefaultBufferSize,
		reader:     r,
		state:      stateInitial}
}

func (r *Reader) startComment()           { r.comment = []byte{} }
func (r *Reader) appendComment(c byte)    { r.comment = append(r.comment, c) }
func (r *Reader) startSection()           { r.section = []byte{} }
func (r *Reader) appendSection(c byte)    { r.section = append(r.section, c) }
func (r *Reader) appendWhitespace(c byte) { r.whitespace = append(r.whitespace, c) }
func (r *Reader) clearWhitespace()        { r.whitespace = []byte{} }
func (r *Reader) appendKey(c byte)        { r.current.key = append(r.current.key, c) }
func (r *Reader) appendValue(c byte)      { r.current.val = append(r.current.val, c) }
func (r *Reader) keyWhitespace()          { r.current.key = r.applyWhitespace(r.current.key) }
func (r *Reader) valueWhitespace()        { r.current.val = r.applyWhitespace(r.current.val) }

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
	r.current = &readEntry{
		section: r.section,
		comment: r.comment}

	r.commentApplied = true
	r.sectionApplied = true
}

func (r *Reader) completeEntry() {
	r.entries = append(r.entries, r.current)
	r.current = nil
}

func (r *Reader) applyWhitespace(b []byte) []byte {
	ws := r.whitespace
	r.clearWhitespace()
	return append(b, ws...)
}

func (r *Reader) commentWhitespace() {
	c := r.applyWhitespace(r.comment)
	if len(r.comment) > 0 {
		r.comment = c
	}
}

func (r *Reader) sectionWhitespace() {
	s := r.applyWhitespace(r.section)
	if len(r.section) > 0 {
		r.section = s
	}
}

func mergeKey(section, key []byte) []string {
	var entryKey []byte

	if len(section) > 0 {
		entryKey = section
	}

	if len(key) > 0 {
		if len(entryKey) > 0 {
			entryKey = append(entryKey, '.')
		}

		entryKey = append(entryKey, key...)
	}

	return strings.Split(string(entryKey), ".")
}

func (r *Reader) fetchEntry() *Entry {
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

func (r *Reader) hasRemainderSection() bool {
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

func (r *Reader) eofResult() (*Entry, error) {
	err := io.EOF
	hrs := r.hasRemainderSection()
	if hrs || r.escapeNext {
		err = EOFRemainder
	}

	var last *Entry
	if r.current != nil {
		r.completeEntry()
		last = r.fetchEntry()
	} else if (!r.commentApplied && len(r.comment) > 0) ||
		(!r.sectionApplied && !hrs && len(r.section) > 0) {

		r.newEntry()
		r.completeEntry()
		last = r.fetchEntry()
		r.commentApplied = true
		r.sectionApplied = true
	}

	return last, err
}

func (r *Reader) ReadEntry() (*Entry, error) {
	if r.reader == nil {
		return nil, nil
	}

	next := r.fetchEntry()
	if next != nil {
		return next, nil
	}

	if r.eof {
		return r.eofResult()
	}

	if len(r.buffer) != r.BufferSize {
		r.buffer = make([]byte, r.BufferSize)
	}

	l, err := r.reader.Read(r.buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	r.eof = err == io.EOF
	if r.eof && l == 0 {
		return r.eofResult()
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
		return r.eofResult()
	}

	return nil, nil
}

func (r *Reader) appendChar(c byte) {
	switch r.state {
	case stateInitial:
		switch {
		case r.escape:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case whitespace(c):
			r.state = stateInitial
		case newline(c):
			r.state = stateInitial
		case startComment(c):
			r.state = stateCommentInitial
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueInitial
			r.newEntry()
		default:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		}

	case stateCommentInitial:
		switch {
		case r.escape:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentInitial
		case newline(c):
			r.state = stateContinueCommentOrElse
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
			r.state = stateCommentOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrElse
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

	case stateCommentOrElse:
		switch {
		case r.escape:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrElse
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

	case stateContinueCommentOrElse:
		switch {
		case r.escape:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case whitespace(c):
			r.state = stateContinueCommentOrElse
		case newline(c):
			r.state = stateContinueCommentOrElse
		case startComment(c):
			r.state = stateCommentInitial
		case openSection(c):
			r.state = stateSectionInitial
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueInitial
			r.newEntry()
		default:
			r.state = stateKey
			r.newEntry()
			r.appendKey(c)
		}

	case stateSectionInitial:
		switch {
		case r.escape:
			r.state = stateSection
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionInitial
		case newline(c):
			r.state = stateSectionInitial
		case startComment(c):
			r.state = stateSection
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.appendSection(c)
		case closeSection(c):
			r.state = stateInitial
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
			r.state = stateSectionOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateSectionOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateSection
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.appendSection(c)
		case closeSection(c):
			r.state = stateInitial
		case startValue(c):
			r.state = stateSection
			r.appendSection(c)
		default:
			r.state = stateSection
			r.appendSection(c)
		}

	case stateSectionOrElse:
		switch {
		case r.escape:
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateSectionOrElse
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
			r.state = stateInitial
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
			r.state = stateKeyOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueInitial
		default:
			r.state = stateKey
			r.appendKey(c)
		}

	case stateKeyOrElse:
		switch {
		case r.escape:
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		case whitespace(c):
			r.state = stateKeyOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		case startValue(c):
			r.state = stateValueInitial
		default:
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		}

	case stateValueInitial:
		switch {
		case r.escape:
			r.state = stateValue
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueInitial
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateValue
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueInitial
			r.completeEntry()
			r.newEntry()
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
			r.state = stateValueOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateValue
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueInitial
			r.completeEntry()
			r.newEntry()
		default:
			r.state = stateValue
			r.appendValue(c)
		}

	case stateValueOrElse:
		switch {
		case r.escape:
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.startComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.startSection()
		case closeSection(c):
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueInitial
			r.completeEntry()
			r.newEntry()
		default:
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		}
	}
}
