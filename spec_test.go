package keyval

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
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
		io.ErrNoProgress,
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
		io.ErrNoProgress,
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
		io.ErrNoProgress,
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
		io.ErrNoProgress,
	}, {

		// single key, infinite, with newline
		"a key\n",
		true,
		[]*Entry{{Key: []string{"a key"}}},
		io.ErrNoProgress,
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
		io.ErrNoProgress,
	}, {

		// single value, infinite, with newline
		"= a value\n",
		true,
		[]*Entry{{Val: "a value"}},
		io.ErrNoProgress,
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
		io.ErrNoProgress,
	}, {

		// key and value, infinite, with new line
		"a key = a value\n",
		true,
		[]*Entry{{Key: []string{"a key"}, Val: "a value"}},
		io.ErrNoProgress,
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
		io.ErrNoProgress,
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
		io.ErrNoProgress,
	}, {

		// Anything is a comment between a '#' and a new line.
		"# a comment\n",
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

		// One or more '#' are ignored at the beginning of a comment, even if separated by whitespaces.
		"## a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {

		// One or more '#' are ignored at the beginning of a comment, even if separated by whitespaces.
		"# \t# a comment\n",
		false,
		[]*Entry{{Comment: "a comment"}},
		io.EOF,
	}, {

		// A '#' in a comment is part of the comment.
		"# a # comment",
		false,
		[]*Entry{{Comment: "a # comment"}},
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

		// In multiline comments, lines not starting with a '#' are ignored.
		"# In multiline comments, lines not\n" +
			" \t \n" +
			"# starting with a '#' are ignored.",
		false,
		[]*Entry{{Comment: "In multiline comments, lines not\n" +
			"starting with a '#' are ignored.",
		}},
		io.EOF,
	}, {

		// An empty comment line starting with '#', whitespace not counted, is part of the comment, if it is
		// between two non-empty comment lines.
		"# An empty comment line starting with '#', whitespace not counted,\n" +
			" \t # \t \n" +
			"# between two non-empty comment lines.",
		false,
		[]*Entry{{Comment: "An empty comment line starting with '#', whitespace not counted,\n" +
			"\n" +
			"between two non-empty comment lines.",
		}},
		io.EOF,
	}, {

		// A standalone, empty comment discards the current comment for the following entries.
		`# a comment
         key one = value one
         ##
         key two = value two`,
		false,
		[]*Entry{
			{Comment: "a comment", Key: []string{"key one"}, Val: "value one"},
			{Key: []string{"key two"}, Val: "value two"}},
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
		"# escaped \\\\ and escaped \\\n in a comment",
		false,
		[]*Entry{{Comment: "escaped \\ and escaped \n in a comment"}},
		io.EOF,
	}, {

		// Escaped whitespace at the comment boundaries.
		"# \\ \tescaped whitespace at the comment boundaries \\\t",
		false,
		[]*Entry{{Comment: " \tescaped whitespace at the comment boundaries \t"}},
		io.EOF,
	}, {

		// Anything declares a section between '[' and a ']', when '[' is not inside a comment or a section declaration.
		"[a section]",
		false,
		[]*Entry{{Key: []string{"a section"}}},
		io.EOF,
	}, {

		// Anything declares a section between '[' and a ']', when '[' is not inside a comment or a section declaration.
		"# [a section]",
		false,
		[]*Entry{{Comment: "[a section]"}},
		io.EOF,
	}, {

		// A '[' inside a section declaration is part of the declaration.
		"[a [section]",
		false,
		[]*Entry{{Key: []string{"a [section"}}},
		io.EOF,
	}, {

		// A section declaration doesn't need to start in a new line.
		"a key = a value [a section]",
		false,
		[]*Entry{{Key: []string{"a key"}, Val: "a value"}, {Key: []string{"a section"}}},
		io.EOF,
	}, {

		// A section declaration is trimmed from leading and trailing whitespace.
		"[ \n \t a section \n \t ]",
		false,
		[]*Entry{{Key: []string{"a section"}}},
		io.EOF,
	}, {

		// A section declaration can span multiple lines.
		"[ a section\nin multiple lines ]",
		false,
		[]*Entry{{Key: []string{"a section\nin multiple lines"}}},
		io.EOF,
	}, {

		// Sections spanning multiple lines are trimmed only at the start of the first line and the end of last line.
		"[ a section \n in multiple lines ]",
		false,
		[]*Entry{{Key: []string{"a section \n in multiple lines"}}},
		io.EOF,
	}, {

		// Section declarations, just like keys, can be separated by '.', defining hierarchical structure of sections.
		"[ a section / hierarchy/ multiple/levels ]",
		false,
		[]*Entry{{Key: []string{"a section", "hierarchy", "multiple", "levels"}}},
		io.EOF,
	}, {

		// Whitespace around '.' separators is trimmed.
		"[ section \n \t / \n \t hierarchy]",
		false,
		[]*Entry{{Key: []string{"section", "hierarchy"}}},
		io.EOF,
	}, {

		// All keys and values following a section declaration belong to the declared section, until the next section
		// declaration. The section is applied to the keys as a prefix.
		`[section one]
         key one = value one
         key two = value two
         [section two]
         key one = value one
         key two = value two`,
		false,
		[]*Entry{
			{Key: []string{"section one", "key one"}, Val: "value one"},
			{Key: []string{"section one", "key two"}, Val: "value two"},
			{Key: []string{"section two", "key one"}, Val: "value one"},
			{Key: []string{"section two", "key two"}, Val: "value two"}},
		io.EOF,
	}, {

		// An empty section declaration, '[]', discards the current section.
		`[section one]
         key one = value one
         key two = value two
         []
         key one = value one
         key two = value two`,
		false,
		[]*Entry{
			{Key: []string{"section one", "key one"}, Val: "value one"},
			{Key: []string{"section one", "key two"}, Val: "value two"},
			{Key: []string{"key one"}, Val: "value one"},
			{Key: []string{"key two"}, Val: "value two"}},
		io.EOF,
	}, {

		// A section without keys and values, gives an entry with the section as the key and an empty value.
		"[section one][section two]",
		false,
		[]*Entry{{Key: []string{"section one"}}, {Key: []string{"section two"}}},
		io.EOF,
	}, {

		// An incomplete section declaration before EOF gives an error distinct from EOF.
		"[section",
		false,
		nil,
		EOFIncomplete,
	}, {

		// A section without keys and values, gives an entry with the section as the key and an empty value.
		// An incomplete section declaration before EOF gives an error distinct from EOF.
		"[section one][section two",
		false,
		[]*Entry{{Key: []string{"section one"}}},
		EOFIncomplete,
	}, {

		// Escaped '\', '.' and ']' in a section declaration.
		"[section with escaped \\\\, \\/ and \\]]",
		false,
		[]*Entry{{Key: []string{"section with escaped \\, / and ]"}}},
		io.EOF,
	}, {

		// Escaped whitespace at the section boundaries.
		"[\\ \tescaped whitespace \\\t]",
		false,
		[]*Entry{{Key: []string{" \tescaped whitespace \t"}}},
		io.EOF,
	}, {

		// There are no comments inside a section declaration.
		"[section #1]",
		false,
		[]*Entry{{Key: []string{"section #1"}}},
		io.EOF,
	}, {

		// Anything is a key, that is before a new line or EOF, and is not a value, a section or a comment.
		"a key\n",
		true,
		[]*Entry{{Key: []string{"a key"}}},
		io.ErrNoProgress,
	}, {

		// Anything is a key, that is before a new line or EOF, and is not a value, a section or a comment.
		"a key",
		false,
		[]*Entry{{Key: []string{"a key"}}},
		io.EOF,
	}, {

		// Anything is a key, that is before a new line or EOF, and is not a value, a section or a comment.
		"a key = a value",
		false,
		[]*Entry{{Key: []string{"a key"}, Val: "a value"}},
		io.EOF,
	}, {

		// Anything is a key, that is before a new line or EOF, and is not a value, a section or a comment.
		"a key [a section]",
		true,
		[]*Entry{{Key: []string{"a key"}}},
		io.ErrNoProgress,
	}, {

		// Anything is a key, that is before a new line or EOF, and is not a value, a section or a comment.
		"a key # a comment",
		true,
		[]*Entry{{Key: []string{"a key"}}},
		io.ErrNoProgress,
	}, {

		// A key doesn't need to start in a new line.
		"[a section] a key",
		false,
		[]*Entry{{Key: []string{"a section", "a key"}}},
		io.EOF,
	}, {

		// A key is trimmed from leading and trailing whitespace.
		" \t a key \t ",
		false,
		[]*Entry{{Key: []string{"a key"}}},
		io.EOF,
	}, {

		// Keys spanning multiple lines are trimmed at only the start of the first line and the end of last line.
		" \t a key \t \\\n \t in multiple lines \t ",
		false,
		[]*Entry{{Key: []string{"a key \t \n \t in multiple lines"}}},
		io.EOF,
	}, {

		// Keys can be separated by '.', defining hierarchical structure of keys.
		"a/structured/key",
		false,
		[]*Entry{{Key: []string{"a", "structured", "key"}}},
		io.EOF,
	}, {

		// Whitespace around '.' separators is trimmed.
		" \t a \t / \t structured \t / \t key",
		false,
		[]*Entry{{Key: []string{"a", "structured", "key"}}},
		io.EOF,
	}, {

		// Only '\', '/', '=', '[', '#' and '\n' can be escaped in a key.
		"key with escaped \\\\, \\/, \\=, \\[, \\# and \\\nin it",
		false,
		[]*Entry{{Key: []string{"key with escaped \\, /, =, [, # and \nin it"}}},
		io.EOF,
	}, {

		// Escaped whitespace at the key boundaries.
		"\\ \tkey with escaped whitespace \\\t",
		false,
		[]*Entry{{Key: []string{" \tkey with escaped whitespace \t"}}},
		io.EOF,
	}, {

		// Anything is a value, that is after a '=' and before a new line, EOF, another value, section declaration or a
		// comment.
		"= a value\n",
		true,
		[]*Entry{{Val: "a value"}},
		io.ErrNoProgress,
	}, {

		// Anything is a value, that is after a '=' and before a new line, EOF, another value, section declaration or a
		// comment.
		"= a value",
		false,
		[]*Entry{{Val: "a value"}},
		io.EOF,
	}, {

		// Anything is a value, that is after a '=' and before a new line, EOF, another value, section declaration or a
		// comment.
		"= a value = another value",
		false,
		[]*Entry{{Val: "a value"}, {Val: "another value"}},
		io.EOF,
	}, {

		// Anything is a value, that is after a '=' and before a new line, EOF, another value, section declaration or a
		// comment.
		"= a value [section]",
		false,
		[]*Entry{{Val: "a value"}, {Key: []string{"section"}}},
		io.EOF,
	}, {

		// Anything is a value, that is after a '=' and before a new line, EOF, another value, section declaration or a
		// comment.
		"= a value # a comment",
		false,
		[]*Entry{{Val: "a value"}, {Comment: "a comment"}},
		io.EOF,
	}, {

		// A value doesn't need to start in a new line.
		"[section] = value",
		false,
		[]*Entry{{Key: []string{"section"}, Val: "value"}},
		io.EOF,
	}, {

		// A value is trimmed from leading and trailing whitespace.
		"= \t a value \t ",
		false,
		[]*Entry{{Val: "a value"}},
		io.EOF,
	}, {

		// Values spanning multiple lines are trimmed only at the start of the first and the end of last line.
		"= \t a value \t \\\n \t in multiple lines",
		false,
		[]*Entry{{Val: "a value \t \n \t in multiple lines"}},
		io.EOF,
	}, {

		// Only '\', '=', '[', '#' and '\n' can be escaped in a value.
		"= value with escaped \\\\, \\=, \\[, \\# and \\\nin it",
		false,
		[]*Entry{{Val: "value with escaped \\, =, [, # and \nin it"}},
		io.EOF,
	}, {

		// Escaped whitespace at the value boundaries.
		"= \\ \tescaped whitespace \\\t",
		false,
		[]*Entry{{Val: " \tescaped whitespace \t"}},
		io.EOF,
	}} {
		var innerReader io.Reader = bytes.NewBuffer([]byte(d.doc))
		if d.infinite {
			innerReader = &infiniteBuffer{innerReader}
		}

		reader := NewEntryReader(innerReader)

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

			if strings.Join(entry.Key, "/") != strings.Join(checkEntry.Key, "/") {
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
