package info

const (
	EscapeChar        = '\\'
	KeySeparatorChar  = '.'
	StartValueChar    = ':'
	StartValueCharAlt = '='
	OpenSectionChar   = '['
	CloseSectionChar  = ']'
	CommentChar       = ';'
	CommentCharAlt    = '#'
	NewlineChar       = '\n'
	ReturnChar        = '\r'
	SpaceChar         = ' '
	TabChar           = '\t'
)

type Entry struct {
	Key     []string
	Val     string
	Comment string
}

var (
	escapeKey = []byte{
		EscapeChar,
		KeySeparatorChar,
		StartValueChar,
		StartValueCharAlt,
		OpenSectionChar,
		CommentChar,
		CommentCharAlt,
		NewlineChar,
		ReturnChar,
		SpaceChar,
		TabChar}

	escapeVal = []byte{
		EscapeChar,
		StartValueChar,
		StartValueCharAlt,
		OpenSectionChar,
		CommentChar,
		CommentCharAlt,
		NewlineChar,
		ReturnChar,
		SpaceChar,
		TabChar}

	escapeSection = []byte{
		EscapeChar,
		CloseSectionChar,
		KeySeparatorChar,
		NewlineChar,
		ReturnChar,
		SpaceChar,
		TabChar}

	escapeComment = []byte{
		EscapeChar,
		NewlineChar,
		SpaceChar,
		TabChar}
)
