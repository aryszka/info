package keyval

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestReadSpec(t *testing.T) {
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

		// Only '\', '\n', ' ' and '\t' can be escaped in a comment.
		`# Only '\\', '\\n', ' ' and '\\t' can be escaped\` + "\n" +
			`in a comment.`,
		false,
		[]*Entry{{Comment: `Only '\', '\n', ' ' and '\t' can be escaped` + "\n" + `in a comment.`}},
		io.EOF,
	}, {

		// Only '\', '\n', ' ' and '\t' can be escaped in a comment.
		`# Only '\\', '\\n', ' ' and '\\t' \ ` + "\n" +
			`# can be escaped in a comment.`,
		false,
		[]*Entry{{Comment: `Only '\', '\n', ' ' and '\t'  ` + "\n" +
			`can be escaped in a comment.`}},
		io.EOF,
	}, {

		// Only '\', '\n', ' ' and '\t' can be escaped in a comment.
		"# a comment \\[",
		false,
		[]*Entry{{Comment: "a comment ["}},
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
		"[ a section . hierarchy. multiple.levels ]",
		false,
		[]*Entry{{Key: []string{"a section", "hierarchy", "multiple", "levels"}}},
		io.EOF,
	}, {

		// Whitespace around '.' separators is trimmed.
		"[ section \n \t . \n \t hierarchy]",
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

		// Only '\', ']', '.', '\n', ' ' and '\t' can be escaped inside a section declaration.
		"[section \\\\ ]",
		false,
		[]*Entry{{Key: []string{"section \\"}}},
		io.EOF,
	}, {

		// Only '\', ']', '.', '\n', ' ' and '\t' can be escaped inside a section declaration.
		"[section \\] ]",
		false,
		[]*Entry{{Key: []string{"section ]"}}},
		io.EOF,
	}, {

		// Only '\', ']', '.', '\n', ' ' and '\t' can be escaped inside a section declaration.
		"[a \\. section]",
		false,
		[]*Entry{{Key: []string{"a . section"}}},
		io.EOF,
	}, {

		// Only '\', ']', '.', '\n', ' ' and '\t' can be escaped inside a section declaration.
		"[section \\\n ]",
		false,
		[]*Entry{{Key: []string{"section \n"}}},
		io.EOF,
	}, {

		// Only '\', ']', '.', '\n', ' ' and '\t' can be escaped inside a section declaration.
		"[section \\  ]",
		false,
		[]*Entry{{Key: []string{"section  "}}},
		io.EOF,
	}, {

		// Only '\', ']', '.', '\n', ' ' and '\t' can be escaped inside a section declaration.
		"[section \\\t ]",
		false,
		[]*Entry{{Key: []string{"section \t"}}},
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
		nil,
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
		nil,
	}, {

		// Anything is a key, that is before a new line or EOF, and is not a value, a section or a comment.
		"a key # a comment",
		true,
		[]*Entry{{Key: []string{"a key"}}},
		nil,
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
		"a.structured.key",
		false,
		[]*Entry{{Key: []string{"a", "structured", "key"}}},
		io.EOF,
	}, {

		// Whitespace around '.' separators is trimmed.
		" \t a \t . \t structured \t . \t key",
		false,
		[]*Entry{{Key: []string{"a", "structured", "key"}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"\\\\",
		false,
		[]*Entry{{Key: []string{"\\"}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"single\\.key",
		false,
		[]*Entry{{Key: []string{"single.key"}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"\\: not a value",
		false,
		[]*Entry{{Key: []string{": not a value"}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"\\[not a section]",
		false,
		[]*Entry{{Key: []string{"[not a section]"}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"\\# not a comment",
		false,
		[]*Entry{{Key: []string{"# not a comment"}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"\\\n not a new entry",
		false,
		[]*Entry{{Key: []string{"\n not a new entry"}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"\\ ",
		false,
		[]*Entry{{Key: []string{" "}}},
		io.EOF,
	}, {

		// Only '\', '.', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a key.
		"\\\t",
		false,
		[]*Entry{{Key: []string{"\t"}}},
		io.EOF,
	}, {

		// Anything is a value, that is after a '=' and before a new line, EOF, another value, section declaration or a
		// comment.
		"= a value\n",
		true,
		[]*Entry{{Val: "a value"}},
		nil,
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

		// Only '\', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a value.
		"= \\\\",
		false,
		[]*Entry{{Val: "\\"}},
		io.EOF,
	}, {

		// Only '\', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a value.
		"= still the same \\= value",
		false,
		[]*Entry{{Val: "still the same = value"}},
		io.EOF,
	}, {

		// Only '\', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a value.
		"= a value \\[ not a section ]",
		false,
		[]*Entry{{Val: "a value [ not a section ]"}},
		io.EOF,
	}, {

		// Only '\', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a value.
		"= a value \\# not a comment",
		false,
		[]*Entry{{Val: "a value # not a comment"}},
		io.EOF,
	}, {

		// Only '\', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a value.
		"= a value \\\nin two lines",
		false,
		[]*Entry{{Val: "a value \nin two lines"}},
		io.EOF,
	}, {

		// Only '\', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a value.
		"= \\ ",
		false,
		[]*Entry{{Val: " "}},
		io.EOF,
	}, {

		// Only '\', '=', '[', '#', '\n', ' ' and '\t' can be escaped in a value.
		"= \\\t",
		false,
		[]*Entry{{Val: "\t"}},
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

func TestWriteSpec(t *testing.T) {
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
	}} {
		buf := &bytes.Buffer{}
		w := NewWriter(buf)
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
