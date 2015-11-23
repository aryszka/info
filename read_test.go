package keyval

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type infiniteBuffer struct {
	reader io.Reader
}

func (b *infiniteBuffer) Read(data []byte) (int, error) {
	l, err := b.reader.Read(data)
	if err == io.EOF {
		err = nil
	}

	return l, err
}

func TestNothingToRead(t *testing.T) {
	r := NewReader(nil)
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read")
	}
}

func TestEmptyReader(t *testing.T) {
	r := NewReader(&infiniteBuffer{bytes.NewBuffer(nil)})
	e, err := r.ReadEntry()
	if e != nil || err != nil {
		t.Error("failed not to read")
	}
}

func TestEmptyReaderEof(t *testing.T) {
	r := NewReader(bytes.NewBuffer(nil))
	e, err := r.ReadEntry()
	if e != nil || err != io.EOF {
		t.Error("failed to read eof")
	}
}

func TestParse(t *testing.T) {
	for i, d := range []struct {
		doc      string
		infinite bool
		entries  []*Entry
		err      error
	}{{
		// empty, eof
		"",
		false,
		nil,
		io.EOF,
	}, {
		// empty, infinite
		"",
		true,
		nil,
		nil,
	}, {
		// single comment, eof
		"# a comment",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// single comment, infinite
		"# a comment",
		true,
		nil,
		nil,
	}, {
		// single section, eof
		"[section]",
		false,
		[]*Entry{{Key: []string{"section"}}},
		io.EOF,
	}, {
		// single section, infinite
		"[section]",
		true,
		nil,
		nil,
	}, {
		// single key, eof
		"a key",
		false,
		[]*Entry{{Key: []string{"a key"}}},
		io.EOF,
	}, {
		// single key, infinite
		"a key",
		true,
		nil,
		nil,
	}, {
		// single key, infinite, with newline
		"a key\n",
		true,
		[]*Entry{{Key: []string{"a key"}}},
		nil,
	}, {
		// single value, eof
		"= a value",
		false,
		[]*Entry{{Val: "a value"}},
		io.EOF,
	}, {
		// single value, infinite
		"= a value",
		true,
		nil,
		nil,
	}, {
		// single value, infinite, with newline
		"= a value\n",
		true,
		[]*Entry{{Val: "a value"}},
		nil,
	}, {
		// key and value, eof
		"a key = a value",
		false,
		[]*Entry{{Key: []string{"a key"}, Val: "a value"}},
		io.EOF,
	}, {
		// key and value, infinite
		"a key = a value",
		true,
		nil,
		nil,
	}, {
		// key and value, infinite, with new line
		"a key = a value\n",
		true,
		[]*Entry{{Key: []string{"a key"}, Val: "a value"}},
		nil,
	}, {
		// comment, section, key, value, eof
		`# a comment
        [a section]
        a key = a value`,
		false,
		[]*Entry{{
			Comment: "a comment",
			Key:     []string{"a section", "a key"},
			Val:     "a value",
		}},
		io.EOF,
	}, {
		// comment, section, key, value, infinite
		`# a comment
        [a section]
        a key = a value`,
		true,
		nil,
		nil,
	}, {
		// comment, section, key, value, infinite, with newline
		`# a comment
        [a section]
        a key = a value
        `,
		true,
		[]*Entry{{
			Comment: "a comment",
			Key:     []string{"a section", "a key"},
			Val:     "a value",
		}},
		nil,
	}, {
		// Anything is a comment between a '#' or a ';' and a new line.
		"# a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// Anything is a comment between a '#' or a ';' and a new line.
		"; a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// A comment doesn't need to start in a new line.
		"a key = a value # a comment\n",
		false,
		[]*Entry{
			{Key: []string{"a key"}, Val: "a value"},
			{Comment: "a comment"},
		},
		io.EOF,
	}, {
		// A comment is trimmed from leading and ending whitespace.
		"# \t a comment \t ",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// One or more '#' or ';' are ignored at the beginning of a comment, even if separated by whitespaces.
		"## a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// One or more '#' or ';' are ignored at the beginning of a comment, even if separated by whitespaces.
		"# \t# a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// One or more '#' or ';' are ignored at the beginning of a comment, even if separated by whitespaces.
		";; a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// One or more '#' or ';' are ignored at the beginning of a comment, even if separated by whitespaces.
		"; \t# a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// A comment can span multiple lines if it is not broken by a section, a key or a value.
		"# A comment can span multiple\n" +
			"# lines if it is not broken by\n" +
			"# a section, a key or a value.",
		false,
		[]*Entry{{Comment: "A comment can span multiple\n" +
			"lines if it is not broken by\n" +
			"a section, a key or a value."}},
		io.EOF,
	}, {
		// A comment can span multiple lines if it is not broken by a section, a key or a value.
		"# A comment can span multiple\n" +
			"[section]" +
			"# lines if it is not broken by\n" +
			"# a section, a key or a value.",
		false,
		[]*Entry{
			{Comment: "lines if it is not broken by\na section, a key or a value.", Key: []string{"section"}},
		},
		io.EOF,
	}, {
		// A comment can span multiple lines if it is not broken by a section, a key or a value.
		"# A comment can span multiple\n" +
			"a key\n" +
			"# lines if it is not broken by\n" +
			"# a section, a key or a value.",
		false,
		[]*Entry{
			{Comment: "A comment can span multiple", Key: []string{"a key"}},
			{Comment: "lines if it is not broken by\na section, a key or a value."},
		},
		io.EOF,
	}, {
		// A comment can span multiple lines if it is not broken by a section, a key or a value.
		"# A comment can span multiple\n" +
			"= a value\n" +
			"# lines if it is not broken by\n" +
			"# a section, a key or a value.",
		false,
		[]*Entry{
			{Comment: "A comment can span multiple", Val: "a value"},
			{Comment: "lines if it is not broken by\na section, a key or a value."},
		},
		io.EOF,
	}, {
		// In multiline comments, lines not starting with a '#' or a ';' are ignored.
		"# In multiline comments, lines not\n" +
			" \t \n" +
			"; starting with a '#' or a ';' are ignored.",
		false,
		[]*Entry{{Comment: "In multiline comments, lines not\n" +
			"starting with a '#' or a ';' are ignored.",
		}},
		io.EOF,
	}, {
		// An empty comment line starting with '#' or a ';', whitespace not counted, is part of the comment, if it is
		// between two non-empty comment lines.
		"# An empty comment line starting with '#' or a ';', whitespace not counted,\n" +
			" \t # \t \n" +
			"# between two non-empty comment lines.",
		false,
		[]*Entry{{Comment: "An empty comment line starting with '#' or a ';', whitespace not counted,\n" +
			"\n" +
			"between two non-empty comment lines.",
		}},
		io.EOF,
	}, {
		// An empty comment line starting with '#' or a ';', whitespace not counted, is part of the comment, if it is
		// between two non-empty comment lines.
		"# An empty comment line starting with '#' or a ';', whitespace not counted,\n" +
			" \t ; \t \n" +
			"# between two non-empty comment lines.",
		false,
		[]*Entry{{Comment: "An empty comment line starting with '#' or a ';', whitespace not counted,\n" +
			"\n" +
			"between two non-empty comment lines.",
		}},
		io.EOF,
	}, {
		// A comment closed by EOF gives an entry without a key and a value.
		"# a comment",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {
		// A comment belongs to all following entries until the next comment.
		`# a comment
         key1 = value1
         key2 = value2
         # another comment
         key3 = value3`,
		false,
		[]*Entry{
			{Comment: "a comment", Key: []string{"key1"}, Val: "value1"},
			{Comment: "a comment", Key: []string{"key2"}, Val: "value2"},
			{Comment: "another comment", Key: []string{"key3"}, Val: "value3"},
		},
		io.EOF,
	}, {
		// Only '\' and '\n' can be escaped in a comment.
		`# Only '\\' and '\\n' can be escaped\` + "\n" +
			`in a comment.`,
		false,
		[]*Entry{{Comment: `Only '\' and '\n' can be escaped` + "\n" + `in a comment.`}},
		io.EOF,
	}, {
		// Only '\' and '\n' can be escaped in a comment.
		"# a comment \\[",
		false,
		[]*Entry{{Comment: "a comment ["}},
		io.EOF,
	}} {
		var innerReader io.Reader = bytes.NewBuffer([]byte(d.doc))
		if d.infinite {
			innerReader = &infiniteBuffer{innerReader}
		}

		reader := NewReader(innerReader)

		var (
			allEntries []*Entry
			err        error
			nillRead   bool
		)

		for {
			var entry *Entry
			entry, err = reader.ReadEntry()
			if entry != nil {
				allEntries = append(allEntries, entry)
			}

			if err != nil {
				break
			}

			if entry == nil {
				if nillRead {
					break
				}

				nillRead = true
				continue
			}
		}

		if err != d.err {
			t.Error(i, "unexpected error", err, d.err)
			return
		}

		if len(allEntries) != len(d.entries) {
			t.Error(i, "invalid number of entries", len(allEntries), len(d.entries))
			return
		}

		for j, entry := range allEntries {
			checkEntry := d.entries[j]

			if entry.Comment != checkEntry.Comment {
				t.Error(i, j, "invalid entry comment")
				t.Log(entry.Comment)
				t.Log(checkEntry.Comment)
				return
			}

			if strings.Join(entry.Key, ".") != strings.Join(checkEntry.Key, ".") {
				t.Error(i, j, "invalid entry key")
				t.Log(strings.Join(entry.Key, "."))
				t.Log(strings.Join(checkEntry.Key, "."))
				return
			}

			if entry.Val != checkEntry.Val {
				t.Error(i, j, "invalid entry value")
				t.Log(entry.Val)
				t.Log(checkEntry.Val)
				return
			}
		}
	}
}
