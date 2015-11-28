package keyval

type readState int

const (
	stateInitial readState = iota
	stateCommentInitial
	stateComment
	stateCommentOrElse
	stateContinueCommentOrElse
	stateSectionInitial
	stateSection
	stateSectionOrElse
	stateKey
	stateKeyOrElse
	stateValueInitial
	stateValue
	stateValueOrElse
)

// choose the right scroll speed and enjoy the animation
func (r *Reader) appendChar(c byte) {
	switch r.state {
	case stateInitial:
		switch {
		case r.escape:
			r.state = stateKey
			r.appendKey(c)
		case whitespace(c):
			r.state = stateInitial
		case newline(c):
			r.state = stateInitial
		case startComment(c):
			r.state = stateCommentInitial
			r.clearWhitespace()
			r.clearComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.clearSection()
		case closeSection(c):
			r.state = stateKey
			r.appendKey(c)
		case keySeparator(c):
			r.state = stateKey
			r.completeKey()
		case startValue(c):
			r.state = stateValueInitial
		default:
			r.state = stateKey
			r.appendKey(c)
		}

	case stateCommentInitial:
		switch {
		case r.escape:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentInitial
		case newline(c):
			r.state = stateContinueCommentOrElse
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateCommentInitial
		case openSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case closeSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case keySeparator(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case startValue(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		default:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		}

	case stateComment:
		switch {
		case r.escape:
			r.state = stateComment
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateComment
			r.appendComment(c)
		case openSection(c):
			r.state = stateComment
			r.appendComment(c)
		case closeSection(c):
			r.state = stateComment
			r.appendComment(c)
		case keySeparator(c):
			r.state = stateComment
			r.appendComment(c)
		case startValue(c):
			r.state = stateComment
			r.appendComment(c)
		default:
			r.state = stateComment
			r.appendComment(c)
		}

	case stateCommentOrElse:
		switch {
		case r.escape:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case whitespace(c):
			r.state = stateCommentOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case openSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case closeSection(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case keySeparator(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case startValue(c):
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		default:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		}

	case stateContinueCommentOrElse:
		switch {
		case r.escape:
			r.state = stateKey
			r.appendKey(c)
		case whitespace(c):
			r.state = stateContinueCommentOrElse
		case newline(c):
			r.state = stateContinueCommentOrElse
		case startComment(c):
			r.state = stateCommentInitial
		case openSection(c):
			r.state = stateSectionInitial
			r.clearSection()
		case closeSection(c):
			r.state = stateKey
			r.appendKey(c)
		case keySeparator(c):
			r.state = stateKey
			r.completeKey()
		case startValue(c):
			r.state = stateValueInitial
		default:
			r.state = stateKey
			r.appendKey(c)
		}

	case stateSectionInitial:
		switch {
		case r.escape:
			r.state = stateSection
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionInitial
		case newline(c):
			r.state = stateSectionInitial
		case startComment(c):
			r.state = stateSection
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.appendSection(c)
		case closeSection(c):
			r.state = stateInitial
			r.completeSection()
		case keySeparator(c):
			r.state = stateSectionInitial
			r.completeSection()
		case startValue(c):
			r.state = stateSection
			r.appendSection(c)
		default:
			r.state = stateSection
			r.appendSection(c)
		}

	case stateSection:
		switch {
		case r.escape:
			r.state = stateSection
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateSectionOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateSection
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.appendSection(c)
		case closeSection(c):
			r.state = stateInitial
			r.completeSection()
		case keySeparator(c):
			r.state = stateSectionInitial
			r.completeSection()
		case startValue(c):
			r.state = stateSection
			r.appendSection(c)
		default:
			r.state = stateSection
			r.appendSection(c)
		}

	case stateSectionOrElse:
		switch {
		case r.escape:
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case whitespace(c):
			r.state = stateSectionOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateSectionOrElse
			r.appendWhitespace(c)
		case startComment(c):
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case openSection(c):
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case closeSection(c):
			r.state = stateInitial
			r.completeSection()
		case keySeparator(c):
			r.state = stateSectionInitial
			r.completeSection()
		case startValue(c):
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		default:
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		}

	case stateKey:
		switch {
		case r.escape:
			r.state = stateKey
			r.appendKey(c)
		case whitespace(c):
			r.state = stateKeyOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeKey()
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeKey()
			r.completeEntry()
			r.clearWhitespace()
			r.clearComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeKey()
			r.completeEntry()
			r.clearSection()
		case closeSection(c):
			r.state = stateKey
			r.appendKey(c)
		case keySeparator(c):
			r.state = stateKeyOrElse
			r.completeKey()
		case startValue(c):
			r.state = stateValueInitial
			r.completeKey()
		default:
			r.state = stateKey
			r.appendKey(c)
		}

	case stateKeyOrElse:
		switch {
		case r.escape:
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		case whitespace(c):
			r.state = stateKeyOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeKey()
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeKey()
			r.completeEntry()
			r.clearWhitespace()
			r.clearComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeKey()
			r.completeEntry()
			r.clearSection()
		case closeSection(c):
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		case keySeparator(c):
			r.state = stateKeyOrElse
			r.completeKey()
		case startValue(c):
			r.state = stateValueInitial
			r.completeKey()
		default:
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		}

	case stateValueInitial:
		switch {
		case r.escape:
			r.state = stateValue
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueInitial
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.clearComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.clearSection()
		case closeSection(c):
			r.state = stateValue
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueInitial
			r.completeEntry()
		default:
			r.state = stateValue
			r.appendValue(c)
		}

	case stateValue:
		switch {
		case r.escape:
			r.state = stateValue
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.clearComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.clearSection()
		case closeSection(c):
			r.state = stateValue
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueInitial
			r.completeEntry()
		default:
			r.state = stateValue
			r.appendValue(c)
		}

	case stateValueOrElse:
		switch {
		case r.escape:
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		case whitespace(c):
			r.state = stateValueOrElse
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateInitial
			r.completeEntry()
		case startComment(c):
			r.state = stateCommentInitial
			r.completeEntry()
			r.clearWhitespace()
			r.clearComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.completeEntry()
			r.clearSection()
		case closeSection(c):
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		case startValue(c):
			r.state = stateValueInitial
			r.completeEntry()
		default:
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		}
	}
}
