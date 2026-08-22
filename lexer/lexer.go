package lexer

import (
	"Tokype/evalut"
	"Tokype/token"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
}

func NewLexer(input string) *Lexer {
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
	l.readPosition++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	if l.ch == '~' && l.peekChar() == '~' {
		l.skipComment()
		return l.NextToken()
	}

	var tok token.Token

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: string(ch) + string(l.ch)}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: string(l.ch)}
		}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: string(l.ch)}
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: string(l.ch)}
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: string(l.ch)}
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: string(l.ch)}
	case '%':
		tok = token.Token{Type: token.MODULO, Literal: string(l.ch)}
	case ':':
		if l.peekChar() == ':' {
			l.readChar()
			tok = token.Token{Type: token.MAPEQ, Literal: "::"}
		} else {
			tok = token.Token{Type: token.COLON, Literal: string(l.ch)}
		}
	case '.':
		if l.peekChar() == '[' {
			l.readChar()
			tok = token.Token{Type: token.DOTLBRACKET, Literal: ".["}
		} else if l.peekChar() == '.' {
			l.readChar()
			tok = token.Token{Type: token.DOTDOT, Literal: ".."}
		} else {
			tok = token.Token{Type: token.DOT, Literal: string(l.ch)}
		}
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: string(l.ch)}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: string(l.ch)}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.ch)}
	case '[':
		tok = token.Token{Type: token.OPENLIST, Literal: string(l.ch)}
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: string(l.ch)}
	case '{':
		tok = token.Token{Type: token.OPENMAP, Literal: string(l.ch)}
	case '}':
		tok = token.Token{Type: token.CLOSEMAP, Literal: string(l.ch)}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		return tok
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: "<="}
		} else {
			tok = token.Token{Type: token.LT, Literal: string(l.ch)}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: ">="}
		} else {
			tok = token.Token{Type: token.GT, Literal: string(l.ch)}
		}
	case '?':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "?="}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch)}
		}
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			literal, isFloat := l.readNumber()
			if isFloat {
				tok.Type = token.FLOAT
			} else {
				tok.Type = token.INT
			}
			tok.Literal = literal
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch)}
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	result := l.input[position:l.position]
	l.readChar()
	return result
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return evalut.Intern(l.input[position:l.position])
}

func (l *Lexer) readNumber() (string, bool) {
	isFloat := false
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()

		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[position:l.position], isFloat
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
}
