package jsonpath

const (
	exprError = iota
	exprEOF
	exprParenLeft
	exprParenRight
	exprNumber
	exprPath
	exprBool

	exprOperators
	exprOpEq
	exprOpNeq
	exprOpLt
	exprOpLe
	exprOpGt
	exprOpGe
	exprOpAnd
	exprOpOr
	exprOpPlus
	exprOpMinus
	exprOpSlash
	exprOpStar
	exprOpHat
	exprOpPercent
	exprOpExclam
)

var exprTokenNames = map[int]string{
	exprError: "ERROR",
	exprEOF:   "EOF",

	exprParenLeft:  "(",
	exprParenRight: ")",
	exprNumber:     "NUMBER",
	exprPath:       "PATH",
	exprBool:       "BOOL",
	exprOpEq:       "==",
	exprOpNeq:      "!=",
	exprOpLt:       "<",
	exprOpLe:       "<=",
	exprOpGt:       ">",
	exprOpGe:       ">=",
	exprOpAnd:      "&&",
	exprOpOr:       "||",
	exprOpPlus:     "+",
	exprOpMinus:    "-",
	exprOpSlash:    "/",
	exprOpStar:     "*",
	exprOpHat:      "^",
	exprOpPercent:  "%",
	exprOpExclam:   "!",
}

var EXPRESSION = lexExprText

func lexExprText(l lexer, state *intStack) stateFn {
	ignoreSpaceRun(l)
	cur := l.peek()
	var next stateFn
	switch cur {
	case '(':
		l.take()
		state.push(exprParenLeft)
		l.emit(exprParenLeft)
		next = lexExprText
	case ')':
		if top, ok := state.peek(); ok && top != exprParenLeft {
			next = l.errorf("Received %#U but has no matching (", cur)
			break
		}
		state.pop()
		l.take()
		l.emit(exprParenRight)

		if state.len() == 0 { // assumes root expression is always encased in ( and )
			next = lexExprEnd
		} else {
			next = lexExprText
		}
	case '=':
		l.take()
		cur = l.take()
		if cur != '=' {
			return l.errorf("Expected double = instead of %#U", cur)
		}
		l.emit(exprOpEq)
		next = lexExprText
	case '<':
		l.take()
		cur = l.peek()
		if cur == '=' {
			l.take()
			l.emit(exprOpLe)
		} else {
			l.emit(exprOpLt)
		}
		next = lexExprText
	case '>':
		l.take()
		cur = l.peek()
		if cur == '=' {
			l.take()
			l.emit(exprOpGe)
		} else {
			l.emit(exprOpGt)
		}
		next = lexExprText
	case '@', '$':
		l.take()
		takePath(l)
		l.emit(exprPath)
		next = lexExprText
	case '+':
		l.take()
		l.emit(exprOpPlus)
		next = lexExprText
	case '-':
		l.take()
		l.emit(exprOpMinus)
		next = lexExprText
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		takeNumeric(l)
		l.emit(exprNumber)
		next = lexExprText
	case eof:
		l.emit(exprEOF)
		// next = nil
	}
	return next
}

func takeNumeric(l lexer) {
	takeDigits(l)
	if l.peek() == '.' {
		l.take()
		takeDigits(l)
	}
	if l.peek() == 'e' || l.peek() == 'E' {
		l.take()
		if l.peek() == '+' || l.peek() == '-' {
			takeDigits(l)
		}
	}
}

func takePath(l lexer) {
	inQuotes := false
	var prev int = 0
	// capture until end of path - ugly
takeLoop:
	for {
		cur := l.peek()
		switch cur {
		case '"':
			if prev != '\\' {
				inQuotes = !inQuotes
			}
			l.take()
		case ' ':
			if !inQuotes {
				break takeLoop
			}
			l.take()
		case eof:
			break takeLoop
		default:
			l.take()
		}

		prev = cur
	}
}

func lexExprEnd(l lexer, state *intStack) stateFn {
	cur := l.take()
	if cur != eof {
		return l.errorf("Expected EOF but received %#U", cur)
	}
	l.emit(exprEOF)
	return nil
}
