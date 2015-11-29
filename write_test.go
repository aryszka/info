package keyval

import "testing"

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
		{" \n", "some longer text with\nnew line ",
			"some\\ longer\\ text\\ with\\\nnew\\ line\\ "},
	} {
		out := string(escapeWrite([]byte(ti.in), []byte(ti.escaped)))
		if out != ti.out {
			t.Error(i, ti.escaped, ti.in, ti.out, out)
		}
	}
}
