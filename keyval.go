package keyval

const (
	EscapeChar       = '\\'
	KeySeparatorChar = '.'
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
