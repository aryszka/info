package info

import (
	"errors"
	"io"
	"strings"
)

type Writer struct {
	BufferSize        int
	UseAltCommentChar bool
	writer            io.Writer
	started           bool
	buffer            []byte
	comment           string
	section           []string
}

var WriteLengthError = errors.New("write failed: byte count does not match")

func NewWriter(w io.Writer) *Writer {
	return &Writer{BufferSize: 0, writer: w}
}

func (w *Writer) writeAll(b []byte) error {
	if l, err := w.writer.Write(b); err != nil {
		return err
	} else if l != len(b) {
		return WriteLengthError
	}

	return nil
}

func (w *Writer) writeBuffer() error {
	for len(w.buffer) >= w.BufferSize {
		if err := w.writeAll(w.buffer[0:w.BufferSize]); err != nil {
			return err
		}

		w.buffer = w.buffer[w.BufferSize:]
	}

	return nil
}

func (w *Writer) write(b ...byte) error {
	if w.BufferSize <= 0 {
		return w.writeAll(b)
	}

	w.buffer = append(w.buffer, b...)
	return nil
}

func (w *Writer) writeLine() error {
	return w.write(NewlineChar)
}

func escapeChars(b, ec []byte) []byte {
	var ep []int
	for i, c := range b {
		for _, e := range ec {
			if c == e {
				ep = append(ep, i)
				break
			}
		}
	}

	var eb []byte
	for i, p := range ep {
		eb = append(eb, append(b[len(eb)-i:p], EscapeChar, b[p])...)
	}

	return eb
}

func (w *Writer) needWriteComment(comment string) bool {
	return w.comment != comment
}

func (w *Writer) commentChar() byte {
	if w.UseAltCommentChar {
		return CommentCharAlt
	}

	return CommentChar
}

func (w *Writer) writeComment() error {
	cc := w.commentChar()
	if w.comment == "" {
		return w.write(cc, cc, NewlineChar)
	}

	lines := strings.Split(w.comment, string([]byte{NewlineChar}))
	for _, l := range lines {
		if err := w.write(append([]byte{cc, ' '}, append(escapeChars([]byte(l), escapeComment), NewlineChar)...)...); err != nil {
			return err
		}
	}

	return nil
}

func (w *Writer) splitSection(key []string) ([]string, []string) {
	if len(key) == 0 {
		return nil, nil
	}

	last := len(key) - 1
	return key[:last], key[last:]
}

func (w *Writer) needWriteSection(section []string) bool {
	if len(section) != len(w.section) {
		return true
	}

	for i, s := range section {
		if s != w.section[i] {
			return true
		}
	}

	return false
}

func (w *Writer) writeKeyEscaped(key []string, esc []byte) error {
	first := true
	for _, s := range w.section {
		if !first {
			if err := w.write(KeySeparatorChar); err != nil {
				return err
			}
		}

		if err := w.write(escapeChars([]byte(s), esc)...); err != nil {
			return err
		}

		first = false
	}

	return nil
}

func (w *Writer) writeSection() error {
	if err := w.write(OpenSectionChar); err != nil {
		return err
	}

	if err := w.writeKeyEscaped(w.section, escapeSection); err != nil {
		return err
	}

	return w.write(CloseSectionChar)
}

func hasKey(key []string) bool {
	return false
}

func (w *Writer) writeKey(key []string) error {
	return w.writeKeyEscaped(key, escapeKey)
}

func (w *Writer) writeVal(val string) error {
	if err := w.write(StartValueChar, SpaceChar); err != nil {
		return err
	}

	return w.write(escapeChars([]byte(val), escapeVal)...)
}

func (w *Writer) WriteEntry(e *Entry) error {
	var err error
	withError := func(f ...func() error) {
		for err == nil && len(f) > 0 {
			err = f[0]()
			f = f[1:]
		}
	}

	commentWritten := false
	if w.needWriteComment(e.Comment) {
		w.comment = e.Comment

		if w.started {
			withError(w.writeLine)
		}

		withError(w.writeComment)
		commentWritten = true
	}

	section, key := w.splitSection(e.Key)
	if w.needWriteSection(section) {
		w.section = section

		if w.started && !commentWritten {
			withError(w.writeLine)
		}

		withError(w.writeSection, w.writeLine)
	}

	hkey := hasKey(key)
	if hkey {
		withError(func() error { return w.writeKey(key) })
	}

	if len(e.Val) > 0 {
		withError(func() error { return w.writeVal(e.Val) })
	}

	if hkey || len(e.Val) > 0 {
		withError(w.writeLine)
	}

	return err
}

func (w *Writer) Flush() error {
	if err := w.writeBuffer(); err != nil {
		return err
	}

	return w.writeAll(w.buffer)
}
