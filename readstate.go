package info

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

func (r *Reader) appendChar(c byte) {
	if r.escape {
		switch r.state {
		case stateInitial:
			r.state = stateKey
			r.appendKey(c)
		case stateCommentInitial:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case stateComment:
			r.appendComment(c)
		case stateCommentOrElse:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		case stateContinueCommentOrElse:
			r.state = stateKey
			r.appendKey(c)
		case stateSectionInitial, stateSection:
			r.state = stateSection
			r.appendSection(c)
		case stateSectionOrElse:
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		case stateKey:
			r.appendKey(c)
		case stateKeyOrElse:
			r.state = stateKey
			r.keyWhitespace()
			r.appendKey(c)
		case stateValueInitial, stateValue:
			r.state = stateValue
			r.appendValue(c)
		case stateValueOrElse:
			r.state = stateValue
			r.valueWhitespace()
			r.appendValue(c)
		}

		return
	}

	switch r.state {
	case stateInitial:
		switch {
		case whitespace(c), newline(c):
		case startComment(c):
			r.state = stateCommentInitial
			r.clearWhitespace()
			r.clearComment()
		case openSection(c):
			r.state = stateSectionInitial
			r.clearSection()
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
		case whitespace(c), startComment(c):
		case newline(c):
			r.state = stateContinueCommentOrElse
			r.appendWhitespace(c)
		default:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		}

	case stateComment:
		switch {
		case whitespace(c):
			r.state = stateCommentOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		default:
			r.appendComment(c)
		}

	case stateCommentOrElse:
		switch {
		case whitespace(c):
			r.appendWhitespace(c)
		case newline(c):
			r.state = stateContinueCommentOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		default:
			r.state = stateComment
			r.commentWhitespace()
			r.appendComment(c)
		}

	case stateContinueCommentOrElse:
		switch {
		case whitespace(c), newline(c):
		case startComment(c):
			r.state = stateCommentInitial
		case openSection(c):
			r.state = stateSectionInitial
			r.clearSection()
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
		case whitespace(c), newline(c):
		case closeSection(c):
			r.state = stateInitial
			r.completeSection()
		case keySeparator(c):
			r.completeSection()
		default:
			r.state = stateSection
			r.appendSection(c)
		}

	case stateSection:
		switch {
		case whitespace(c), newline(c):
			r.state = stateSectionOrElse
			r.clearWhitespace()
			r.appendWhitespace(c)
		case closeSection(c):
			r.state = stateInitial
			r.completeSection()
		case keySeparator(c):
			r.state = stateSectionInitial
			r.completeSection()
		default:
			r.appendSection(c)
		}

	case stateSectionOrElse:
		switch {
		case whitespace(c), newline(c):
			r.appendWhitespace(c)
		case closeSection(c):
			r.state = stateInitial
			r.completeSection()
		case keySeparator(c):
			r.state = stateSectionInitial
			r.completeSection()
		default:
			r.state = stateSection
			r.sectionWhitespace()
			r.appendSection(c)
		}

	case stateKey:
		switch {
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
		case keySeparator(c):
			r.state = stateKeyOrElse
			r.completeKey()
		case startValue(c):
			r.state = stateValueInitial
			r.completeKey()
		default:
			r.appendKey(c)
		}

	case stateKeyOrElse:
		switch {
		case whitespace(c):
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
		case keySeparator(c):
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
		case whitespace(c):
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
		case startValue(c):
			r.completeEntry()
		default:
			r.state = stateValue
			r.appendValue(c)
		}

	case stateValue:
		switch {
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
		case startValue(c):
			r.state = stateValueInitial
			r.completeEntry()
		default:
			r.appendValue(c)
		}

	case stateValueOrElse:
		switch {
		case whitespace(c):
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
