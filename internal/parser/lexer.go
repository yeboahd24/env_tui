package parser

import (
	"strings"
	"unicode"
)

type TokenType int

const (
	KEY TokenType = iota
	VALUE
	COMMENT
	NEWLINE
	EXPORT
	EOF
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

type Lexer struct {
	input   string
	pos     int
	line    int
	col     int
	current rune
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.pos >= len(l.input) {
		l.current = 0
	} else {
		l.current = rune(l.input[l.pos])
	}
	l.pos++
	l.col++
}

func (l *Lexer) peekChar() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return rune(l.input[l.pos])
}

func (l *Lexer) skipWhitespace() {
	for l.current == ' ' || l.current == '\t' || l.current == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readString(delimiter rune) string {
	start := l.pos
	l.readChar() // skip opening quote
	
	for l.current != delimiter && l.current != 0 {
		if l.current == '\\' {
			l.readChar() // skip escape char
		}
		l.readChar()
	}
	
	if l.current == delimiter {
		l.readChar() // skip closing quote
	}
	
	return l.input[start:l.pos-1]
}

func (l *Lexer) readIdentifier() string {
	start := l.pos - 1
	for unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_' {
		l.readChar()
	}
	return l.input[start:l.pos-1]
}

func (l *Lexer) ReadValue() string {
	// We're positioned right after the = sign and any whitespace
	// l.current points to the first character of the value
	start := l.pos - 1 // Start from current character
	
	// Handle quoted values
	if l.current == '"' {
		l.readChar() // skip opening quote
		valueStart := l.pos - 1
		for l.current != '"' && l.current != 0 && l.current != '\n' {
			if l.current == '\\' {
				l.readChar() // skip escape char
			}
			l.readChar()
		}
		value := l.input[valueStart:l.pos-1]
		if l.current == '"' {
			l.readChar() // skip closing quote
		}
		return value
	}
	
	if l.current == '\'' {
		l.readChar() // skip opening quote
		valueStart := l.pos - 1
		for l.current != '\'' && l.current != 0 && l.current != '\n' {
			if l.current == '\\' {
				l.readChar() // skip escape char
			}
			l.readChar()
		}
		value := l.input[valueStart:l.pos-1]
		if l.current == '\'' {
			l.readChar() // skip closing quote
		}
		return value
	}
	
	// Read unquoted value until newline or comment
	if l.current == '\n' || l.current == '#' || l.current == 0 {
		return "" // Empty value
	}
	
	for l.current != '\n' && l.current != '#' && l.current != 0 {
		l.readChar()
	}
	
	// Get the value from start to current position
	value := l.input[start:l.pos-1]
	return strings.TrimSpace(value)
}

func (l *Lexer) readComment() string {
	start := l.pos - 1
	for l.current != '\n' && l.current != 0 {
		l.readChar()
	}
	return l.input[start:l.pos-1]
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	
	token := Token{Line: l.line, Col: l.col}
	
	switch l.current {
	case 0:
		token.Type = EOF
	case '\n':
		token.Type = NEWLINE
		token.Value = "\n"
		l.readChar()
		l.line++
		l.col = 0
	case '#':
		token.Type = COMMENT
		token.Value = l.readComment()
	default:
		if unicode.IsLetter(l.current) || l.current == '_' {
			identifier := l.readIdentifier()
			
			if identifier == "export" {
				token.Type = EXPORT
				token.Value = identifier
				return token
			}
			
			// Check if this is a key (followed by =)
			l.skipWhitespace()
			if l.current == '=' {
				token.Type = KEY
				token.Value = identifier
				l.readChar() // consume =
				l.skipWhitespace()
			} else {
				// Not a key, treat as value
				token.Type = VALUE
				token.Value = identifier
			}
		} else {
			// Skip unknown characters
			l.readChar()
			return l.NextToken()
		}
	}
	
	return token
}