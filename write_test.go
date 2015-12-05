package keyval

import (
	"bytes"
	"errors"
	"testing"
)

type errWriter struct {
	writeCount int
	failLength bool
}

var (
	errExpectedFailingWrite   = errors.New("expected failing write")
	errUnexpectedFailingWrite = errors.New("unexpected failing write")
)

func (er *errWriter) Write(b []byte) (int, error) {
	er.writeCount++
	if er.failLength {
		return len(b) - 1, nil
	}

	if er.writeCount <= 1 {
		return 0, errExpectedFailingRead
	}

	return 0, errUnexpectedFailingRead
}

func TestEscapeWrite(t *testing.T) {
	for i, ti := range []struct{ escaped, in, out string }{
		{"", "abc", "abc"},
		{"a", "abc", "\\abc"},
		{"b", "abc", "a\\bc"},
		{"c", "abc", "ab\\c"},
		{"ab", "abc", "\\a\\bc"},
		{"ac", "abc", "\\ab\\c"},
		{"bc", "abc", "a\\b\\c"},
		{"abc", "abc", "\\a\\b\\c"},
	} {
		out := string(escapeWrite([]byte(ti.in), []byte(ti.escaped)))
		if out != ti.out {
			t.Error(i, ti.escaped, ti.in, ti.out, out)
		}
	}
}

func TestEscapeBoundaries(t *testing.T) {
	for i, ti := range []struct{ escapedLead, escapedTrail, in, out string }{
		{"", "", "abc", "abc"},
		{"a", "", "abc", "\\abc"},
		{"b", "", "abc", "abc"},
		{"c", "", "abc", "ab\\c"},
		{"ab", "", "abc", "\\abc"},
		{"ac", "", "abc", "\\ab\\c"},
		{"bc", "", "abc", "ab\\c"},
		{"abc", "", "abc", "\\ab\\c"},
		{"a", "c", "abc", "\\ab\\c"},
		{"", "c", "abc", "ab\\c"},
		{"c", "a", "abc", "abc"},
	} {
		lec := []byte(ti.escapedLead)
		tec := []byte(ti.escapedTrail)
		if len(tec) == 0 {
			tec = lec
		}

		out := string(escapeBoundaries([]byte(ti.in), lec, tec))
		if out != ti.out {
			t.Error(i, ti.escapedLead, ti.escapedTrail, ti.in, ti.out, out)
		}
	}
}

func TestEscapeOutput(t *testing.T) {
	for i, ti := range []struct{ escapedLead, escaped, escapedTrail, in, out string }{
		{"", "", "", "abcde", "abcde"},
		{"", "c", "", "abcde", "ab\\cde"},
		{"a", "c", "e", "abcde", "\\ab\\cd\\e"},
		{"a", "c", "", "abcde", "\\ab\\cde"},
		{"", "c", "e", "abcde", "ab\\cd\\e"},
		{"bcd", "bcd", "bcd", "abcde", "a\\b\\c\\de"},
		{"ace", "", "ace", "abcde", "\\abcd\\e"},
	} {
		out := string(escapeOutput([]byte(ti.in),
			[]byte(ti.escaped), []byte(ti.escapedLead), []byte(ti.escapedTrail)))
		if out != ti.out {
			t.Error(i, ti.escapedLead, ti.escaped, ti.escapedTrail, ti.in, ti.out, out)
		}
	}
}

func TestFailOnInvalidWriteLength(t *testing.T) {
	iw := &errWriter{failLength: true}
	w := NewEntryWriter(iw)
	err := w.WriteEntry(&Entry{Key: []string{"a key"}})
	if err != ErrWriteLength {
		t.Error("failed to fail")
	}
}

func TestReturnSameErrorOnRepeatedWriteCall(t *testing.T) {
	iw := &errWriter{}
	w := NewEntryWriter(iw)
	var err error

	err = w.WriteEntry(&Entry{Key: []string{"a key"}})
	if err != errExpectedFailingRead {
		t.Error("failed to fail")
	}

	err = w.WriteEntry(&Entry{Key: []string{"a key"}})
	if err != errExpectedFailingRead || iw.writeCount != 1 {
		t.Error("failed to store previous failure")
	}
}

func TestWrite(t *testing.T) {
	for i, ti := range []struct {
		entries []*Entry
		output  string
	}{{
		nil,
		"",
	}, {
		[]*Entry{nil},
		"",
	}, {
		[]*Entry{{}},
		"",
	}, {
		[]*Entry{{Key: []string{}}},
		"",
	}, {
		[]*Entry{{Key: []string{""}}},
		"\n",
	}, {
		[]*Entry{{Comment: "a comment"}},
		"# a comment\n",
	}, {
		[]*Entry{{Key: []string{"a key"}}},
		"a key\n",
	}, {
		[]*Entry{{Key: []string{"a section", "a key"}}},
		"[a section]\na key\n",
	}, {
		[]*Entry{{Val: "a value"}},
		"= a value\n",
	}, {
		[]*Entry{{Key: []string{"a key"}, Val: "a value"}},
		"a key = a value\n",
	}, {
		[]*Entry{{Comment: "a comment", Key: []string{"a section", "a key"}, Val: "a value"}},
		"# a comment\n[a section]\na key = a value\n",
	}, {

		// a '#' in a comment
		[]*Entry{{Comment: "a # comment"}},
		"# a # comment\n",
	}, {

		// a comment spanning multiple lines
		[]*Entry{{Comment: "a comment\nspanning\nmultiple lines"}},
		"# a comment\n# spanning\n# multiple lines\n",
	}, {

		// a key between comments
		[]*Entry{
			{Comment: "a comment", Key: []string{"a key"}},
			{Comment: "another\ncomment"}},
		"# a comment\na key\n\n# another\n# comment\n",
	}, {

		// section between comments
		[]*Entry{
			{Comment: "a comment", Key: []string{"a section", "a key"}},
			{Comment: "another\ncomment"}},
		"# a comment\n[a section]\na key\n\n# another\n# comment\n",
	}, {

		// section between comments
		[]*Entry{
			{Comment: "a comment", Val: "a value"},
			{Comment: "another\ncomment"}},
		"# a comment\n= a value\n\n# another\n# comment\n",
	}, {

		// multi-line comment with empty line
		[]*Entry{
			{Comment: "a multiline\n\ncomment"}},
		"# a multiline\n#\n# comment\n",
	}, {

		// discard comment
		[]*Entry{
			{Comment: "a comment", Key: []string{"a key"}},
			{Comment: ""}},
		"# a comment\na key\n\n##\n",
	}, {

		// discard comment, when no entry
		[]*Entry{
			{Comment: "a comment"},
			{Comment: ""}},
		"# a comment\n\n[] ##\n",
	}, {

		// discard comment, when no entry, in a section
		[]*Entry{
			{Key: []string{"a section", "a key"}},
			{Comment: "a comment"},
			{Comment: ""}},
		"[a section]\na key\n\n# a comment\n\n[a section] ##\n",
	}, {

		// comment of multiple entries
		[]*Entry{
			{Comment: "comment one", Key: []string{"key one"}, Val: "value one"},
			{Comment: "comment one", Key: []string{"key two"}, Val: "value two"},
			{Comment: "comment two", Key: []string{"key three"}, Val: "value three"},
			{Comment: "comment two", Key: []string{"key four"}, Val: "value four"}},
		"# comment one\nkey one = value one\nkey two = value two\n\n" +
			"# comment two\nkey three = value three\nkey four = value four\n",
	}, {

		// escape leading, tailing whitespace in comments
		[]*Entry{{Comment: " \t a comment \t "}},
		"# \\ \t a comment \t\\ \n",
	}, {

		// escape leading, tailing comment in comments
		[]*Entry{{Comment: "# a comment"}},
		"# \\# a comment\n",
	}, {

		// section in a comment
		[]*Entry{{Comment: "[section]"}},
		"# [section]\n",
	}, {

		// section in a section
		[]*Entry{{Key: []string{"[section", "key"}}},
		"[[section]\nkey\n",
	}, {

		// leading whitespace in a section
		[]*Entry{{Key: []string{" \t section \t ", "key"}}},
		"[\\ \t section \t\\ ]\nkey\n",
	}, {

		// leading newline in a section
		[]*Entry{{Key: []string{"\nsection\n", "key"}}},
		"[\\\nsection\\\n]\nkey\n",
	}, {

		// new line inside a section
		[]*Entry{{Key: []string{"section\nwith a new line", "key"}}},
		"[section\nwith a new line]\nkey\n",
	}, {

		// section with '.' in the name
		[]*Entry{{Key: []string{"section.with.dot", "key"}}},
		"[section\\.with\\.dot]\nkey\n",
	}, {

		// section of multiple entries
		[]*Entry{
			{Key: []string{"section one", "key one"}, Val: "value one"},
			{Key: []string{"section one", "key two"}, Val: "value two"},
			{Key: []string{"section two", "key three"}, Val: "value three"},
			{Key: []string{"section two", "key four"}, Val: "value four"}},
		"[section one]\n" +
			"key one = value one\n" +
			"key two = value two\n" +
			"\n" +
			"[section two]\n" +
			"key three = value three\n" +
			"key four = value four\n",
	}, {

		// discard a section
		[]*Entry{
			{Key: []string{"section one", "key one"}, Val: "value one"},
			{Key: []string{"section one", "key two"}, Val: "value two"},
			{Key: []string{"key three"}, Val: "value three"},
			{Key: []string{"key four"}, Val: "value four"}},
		"[section one]\n" +
			"key one = value one\n" +
			"key two = value two\n" +
			"\n" +
			"[]\n" +
			"key three = value three\n" +
			"key four = value four\n",
	}, {

		// no comment inside a section declaration
		[]*Entry{{Key: []string{"section #1", "key"}}},
		"[section #1]\n" +
			"key\n",
	}, {

		// leading and trailing whitespace in a key escaped
		[]*Entry{{Key: []string{" \t a key \t "}}},
		"\\ \t a key \t\\ \n",
	}, {

		// key spanning multiple lines
		[]*Entry{{Key: []string{"a key\nin multiple lines"}}},
		"a key\\\nin multiple lines\n",
	}, {

		// key with '.' in the name
		[]*Entry{{Key: []string{"key.with.dot"}}},
		"key\\.with\\.dot\n",
	}, {

		// leading and trailing whitespace in a value escaped
		[]*Entry{{Val: " \t a value \t "}},
		"= \\ \t a value \t\\ \n",
	}, {

		// value spanning multiple lines
		[]*Entry{{Val: "a value\nin multiple lines"}},
		"= a value\\\nin multiple lines\n",
	}, {

		// a value with '=' in it
		[]*Entry{{Val: "value=with=equals"}},
		"= value\\=with\\=equals\n",
	}} {
		buf := &bytes.Buffer{}
		w := NewEntryWriter(buf)
		for _, e := range ti.entries {
			if err := w.WriteEntry(e); err != nil {
				t.Error(err)
				return
			}
		}

		if buf.String() != ti.output {
			t.Error(i, "invalid output")
			t.Log(buf.String())
			t.Log(ti.output)
			return
		}
	}
}
