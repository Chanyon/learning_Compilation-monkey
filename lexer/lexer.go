package lexer

import (
	"monkey/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition = l.readPosition + 1
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace() // skip whitespace ' '
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = token.NewToken(token.ASSIGN, l.ch)
		}
	case '-':
		tok = token.NewToken(token.MINUS, l.ch)
	case '/':
		tok = token.NewToken(token.SLASH, l.ch)
	case '*':
		tok = token.NewToken(token.ASTERISK, l.ch)
	case '+':
		tok = token.NewToken(token.PLUS, l.ch)
	case ';':
		tok = token.NewToken(token.SEMICOLON, l.ch)
	case '(':
		tok = token.NewToken(token.LPAREN, l.ch)
	case ')':
		tok = token.NewToken(token.RPAREN, l.ch)
	case '{':
		tok = token.NewToken(token.LBRACE, l.ch)
	case '}':
		tok = token.NewToken(token.RBRACE, l.ch)
	case ',':
		tok = token.NewToken(token.COMMA, l.ch)
	case '.':
		tok = token.NewToken(token.DOT, l.ch)
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.LTQ, Literal: literal}
		} else {
			tok = token.NewToken(token.LT, l.ch)
		}
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.GTQ, Literal: literal}
		} else {
			tok = token.NewToken(token.GT, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.OR, Literal: literal}
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = token.NewToken(token.BANG, l.ch)
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = string(l.readString())
	case '[':
		tok = token.NewToken(token.LBRACKET, l.ch)
	case ']':
		tok = token.NewToken(token.RBRACKET, l.ch)
	case ':':
		tok = token.NewToken(token.COLON, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			number := l.readNumber()
			// if  {

			// }
			tok.Type = token.INT
			tok.Literal = number
			return tok
		} else {
			tok = token.NewToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	// let a = 'a';
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	// if (l.ch == '.') && isDigit(l.peekChar()) {
	// 	l.readChar()
	// 	for isDigit(l.ch) {
	// 		l.readChar()
	// 	}
	// }
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\\' && (l.peekChar() == '"' || l.peekChar() == '\'') {
			l.readChar()
			continue
		} else if l.ch == '\t' || l.ch == '\n' || l.ch == '\r' || l.ch == ' ' {
			continue
		} else if l.ch == 0 || l.ch == '"' {
			break
		} else {
			continue
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// func peekNext(l *Lexer) byte {
// 	if l.readPosition >= len(l.input) {
// 		return 0
// 	}
// 	return l.input[l.readPosition]
// }

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) checkChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch) || ch == '-' || ch == ',' ||
		ch == '_' || l.ch == '\\' || ch == '/' || ch == '.' ||
		ch == '+' || ch == '`' || ch == '?' || ch == '~' || ch == ';'
}

func (l *Lexer) peekLetter() byte {
	pos := l.readPosition + 1
	if pos >= len(l.input) {
		return 0
	}

	return l.input[pos]
}
