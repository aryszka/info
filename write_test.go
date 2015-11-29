package keyval

import "testing"

func TestEscape(t *testing.T) {
	for i, ti := range []struct{ escaped, in, out string }{
		{"", "abc", "abc"},
		{"a", "abc", "\\abc"},
		{"b", "abc", "a\\bc"},
		{"c", "abc", "ab\\c"},
		{"ab", "abc", "\\a\\bc"},
		{"ac", "abc", "\\ab\\c"},
		{"bc", "abc", "a\\b\\c"},
		{"abc", "abc", "\\a\\b\\c"},
		{" \n", "some longer text with\nnew line ",
			"some\\ longer\\ text\\ with\\\nnew\\ line\\ "},
	} {
		out := string(escapeChars([]byte(ti.in), []byte(ti.escaped)))
		if out != ti.out {
			t.Error(i, ti.escaped, ti.in, ti.out, out)
		}
	}
}
