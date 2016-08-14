package keyval

import "strings"

const (
	EscapeChar       = '\\'
	KeySeparatorChar = '/'
	StartValueChar   = '='
	OpenSectionChar  = '['
	CloseSectionChar = ']'
	CommentChar      = '#'
	NewlineChar      = '\n'
	SpaceChar        = ' '
	TabChar          = '\t'
)

type Entry struct {
	Key     []string
	Val     string
	Comment string
}

var (
	escapeBound        = []byte{SpaceChar, TabChar}
	escapeBoundNl      = []byte{SpaceChar, TabChar, NewlineChar}
	escapeBoundComment = []byte{SpaceChar, TabChar, CommentChar}

	escapeKey = []byte{
		EscapeChar,
		KeySeparatorChar,
		StartValueChar,
		OpenSectionChar,
		CommentChar,
		NewlineChar}

	escapeVal = []byte{
		EscapeChar,
		StartValueChar,
		OpenSectionChar,
		CommentChar,
		NewlineChar}

	escapeSection = []byte{
		EscapeChar,
		CloseSectionChar,
		KeySeparatorChar}

	escapeComment = []byte{EscapeChar}
)

func JoinKey(key []string) string {
	return strings.Join(key, ".")
}

func SplitKey(key string) []string {
	return strings.Split(key, ".")
}

func KeyEq(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for i, k := range left {
		if k != right[i] {
			return false
		}
	}

	return true
}
