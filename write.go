package keyval

import (
	"errors"
	"io"
	"strings"
)

type EntryWriter struct {
	BufferSize        int
	writer            io.Writer
	started           bool
	buffer            []byte
	comment           string
	commentIncomplete bool
	section           []string
	err               error
}

var ErrWriteLength = errors.New("write failed: byte count does not match")

func NewEntryWriter(w io.Writer) *EntryWriter {
	return &EntryWriter{BufferSize: 0, writer: w}
}

func escapeWrite(b, ec []byte) []byte {
	eb := make([]byte, 0, len(b))
	en := 0
	for i, c := range b {
		for _, e := range ec {
			if c == e {
				eb = append(eb, b[len(eb)-en:i]...)
				eb = append(eb, EscapeChar, c)
				en++
			}
		}
	}

	eb = append(eb, b[len(eb)-en:]...)
	return eb
}

func escapeBoundaries(b, lec, tec []byte) []byte {
	switch {
	case len(b) == 0:
		return b
	case len(b) == 1:
		return escapeWrite(b, lec)
	default:
		return append(escapeWrite(b[:1], lec),
			append(b[1:len(b)-1],
				escapeWrite(b[len(b)-1:], tec)...)...)
	}
}

func escapeOutput(b, wec, lec, tec []byte) []byte {
	return escapeBoundaries(escapeWrite(b, wec), lec, tec)
}

func (w *EntryWriter) writeAll(b []byte) error {
	if l, err := w.writer.Write(b); err != nil {
		return err
	} else if l != len(b) {
		return ErrWriteLength
	}

	return nil
}

func (w *EntryWriter) writeBuffer() error {
	for len(w.buffer) > 0 && len(w.buffer) >= w.BufferSize {
		if err := w.writeAll(w.buffer[0:w.BufferSize]); err != nil {
			return err
		}

		w.buffer = w.buffer[w.BufferSize:]
	}

	return nil
}

func (w *EntryWriter) write(b ...byte) error {
	if w.BufferSize <= 0 {
		return w.writeAll(b)
	}

	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *EntryWriter) writeLine() error {
	return w.write(NewlineChar)
}

func (w *EntryWriter) needWriteComment(comment string) bool {
	return w.comment != comment
}

func (w *EntryWriter) writeComment() error {
	var err error

	withError := func(f ...func() error) {
		for err == nil && len(f) > 0 {
			err = f[0]()
			f = f[1:]
		}
	}

	writeWithError := func(b ...byte) {
		withError(func() error { return w.write(b...) })
	}

	if w.commentIncomplete {
		withError(w.writeSection)
		writeWithError(SpaceChar)
	}

	if w.comment == "" {
		writeWithError(CommentChar)
	}

	lines := strings.Split(w.comment, string([]byte{NewlineChar}))
	for _, l := range lines {
		if len(l) == 0 {
			writeWithError(CommentChar, NewlineChar)
		} else {
			writeWithError(append([]byte{CommentChar, SpaceChar},
				append(escapeOutput([]byte(l), escapeComment, escapeBoundComment, escapeBound),
					NewlineChar)...)...)
		}
	}

	return err
}

func (w *EntryWriter) splitSection(key []string) ([]string, []string) {
	if len(key) == 0 {
		return nil, nil
	}

	last := len(key) - 1
	return key[:last], key[last:]
}

func (w *EntryWriter) needWriteSection(section, key []string, val string) bool {
	sectionChanged := false
	if len(section) != len(w.section) {
		sectionChanged = true
	}

	if !sectionChanged {
		for i, s := range section {
			if s != w.section[i] {
				sectionChanged = true
				break
			}
		}
	}

	return sectionChanged && (hasKey(key) || len(val) > 0)
}

func (w *EntryWriter) writeKeyEscaped(key []string, wesc, besc []byte) error {
	first := true
	for _, s := range key {
		if !first {
			if err := w.write(KeySeparatorChar); err != nil {
				return err
			}
		}

		if err := w.write(escapeOutput([]byte(s), wesc, besc, besc)...); err != nil {
			return err
		}

		first = false
	}

	return nil
}

func (w *EntryWriter) writeSection() error {
	if err := w.write(OpenSectionChar); err != nil {
		return err
	}

	if err := w.writeKeyEscaped(w.section, escapeSection, escapeBoundNl); err != nil {
		return err
	}

	return w.write(CloseSectionChar)
}

func hasKey(key []string) bool {
	return len(key) > 0
}

func (w *EntryWriter) writeKey(key []string) error {
	return w.writeKeyEscaped(key, escapeKey, escapeBound)
}

func (w *EntryWriter) writeVal(val string, leadingSpace bool) error {
	if leadingSpace {
		if err := w.write(SpaceChar); err != nil {
			return err
		}
	}

	if err := w.write(StartValueChar, SpaceChar); err != nil {
		return err
	}

	return w.write(escapeOutput([]byte(val), escapeVal, escapeBound, escapeBound)...)
}

func (w *EntryWriter) WriteEntry(e *Entry) error {
	withError := func(f ...func() error) {
		for w.err == nil && len(f) > 0 {
			w.err = f[0]()
			f = f[1:]
		}
	}

	var (
		commentWritten bool
		sectionWritten bool
		keyWritten     bool
		valWritten     bool
	)

	if w.writer == nil || e == nil || w.err != nil {
		return w.err
	}

	if w.needWriteComment(e.Comment) {
		w.comment = e.Comment

		if w.started {
			withError(w.writeLine)
		}

		withError(w.writeComment)
		commentWritten = true
		w.commentIncomplete = true
	}

	section, key := w.splitSection(e.Key)
	if w.needWriteSection(section, key, e.Val) {
		w.section = section

		if w.started && !commentWritten {
			withError(w.writeLine)
		}

		withError(w.writeSection, w.writeLine)
		sectionWritten = true
	}

	hkey := hasKey(key)
	if hkey {
		withError(func() error { return w.writeKey(key) })
		keyWritten = true
	}

	if len(e.Val) > 0 {
		withError(func() error { return w.writeVal(e.Val, keyWritten) })
		valWritten = true
	}

	if keyWritten || valWritten {
		withError(w.writeLine)
	}

	w.started = w.started || commentWritten || sectionWritten || keyWritten || valWritten
	w.commentIncomplete = w.commentIncomplete && !sectionWritten && !keyWritten && !valWritten

	withError(w.writeBuffer)
	return w.err
}

func (w *EntryWriter) Flush() error {
	if w.writer == nil || w.err != nil {
		return w.err
	}

	if w.err = w.writeBuffer(); w.err != nil {
		return w.err
	}

	w.err = w.writeAll(w.buffer)
	return w.err
}
