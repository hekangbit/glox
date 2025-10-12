package main

import "fmt"

const (
	TOKEN_LEFT_PAREN int = iota + 1 // 1
	TOKEN_RIGHT_PAREN
	TOKEN_LEFT_BRACE
	TOKEN_RIGHT_BRACE
	TOKEN_SEMICOLON
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_MINUS
	TOKEN_PLUS
	TOKEN_SLASH
	TOKEN_STAR
	TOKEN_BANG_EQUAL
	TOKEN_BANG
	TOKEN_EQUAL_EQUAL
	TOKEN_EQUAL
	TOKEN_LESS_EQUAL
	TOKEN_LESS
	TOKEN_GREATER_EQUAL
	TOKEN_GREATER
	TOKEN_IDENTIFIER
	TOKEN_STRING
	TOKEN_NUMBER
	TOKEN_AND
	TOKEN_CLASS
	TOKEN_ELSE
	TOKEN_FALSE
	TOKEN_FOR
	TOKEN_FUN
	TOKEN_IF
	TOKEN_NIL
	TOKEN_OR
	TOKEN_PRINT
	TOKEN_RETURN
	TOKEN_SUPER
	TOKEN_THIS
	TOKEN_TRUE
	TOKEN_VAR
	TOKEN_WHILE
	TOKEN_EOF
	TOKEN_ERROR
)

type Token struct {
	token_type int
	lexeme     string
	line       int
}

type Scanner struct {
	line    int
	start   int
	current int
	source  string
}

func (scanner *Scanner) match(expected byte) bool {
	if scanner.isAtEnd() {
		return false
	}
	if scanner.source[scanner.current] != expected {
		return false
	}
	scanner.current++
	return true
}

func (scanner *Scanner) isAtEnd() bool {
	return scanner.current == len(scanner.source)
}

func (scanner *Scanner) advance() byte {
	result := scanner.source[scanner.current]
	scanner.current++
	return result
}

func (scanner *Scanner) peek() byte {
	return scanner.source[scanner.current]
}

func (scanner *Scanner) skipWhitespace() {
	for {
		c := scanner.peek()
		switch c {
		case ' ', '\t', '\r':
			scanner.advance()
		case '\n':
			scanner.line++
			scanner.advance()
		default:
			return
		}
	}
}

func (scanner *Scanner) EOFToken() Token {
	return Token{TOKEN_EOF, "EOF", scanner.line}
}

func (scanner *Scanner) ErrorToken(s string) Token {
	return Token{TOKEN_ERROR, s, scanner.line}
}

func (scanner *Scanner) MakeToken(token_type int) Token {
	var lexeme string = scanner.source[scanner.start:scanner.current]
	return Token{token_type, lexeme, scanner.line}
}

func (scanner *Scanner) ScanToken() Token {
	if scanner.isAtEnd() {
		return scanner.EOFToken()
	}
	scanner.skipWhitespace()
	scanner.start = scanner.current
	c := scanner.advance()
	switch c {
	case '(':
		return scanner.MakeToken(TOKEN_LEFT_PAREN)
	case ')':
		return scanner.MakeToken(TOKEN_RIGHT_PAREN)
	case '{':
		return scanner.MakeToken(TOKEN_LEFT_BRACE)
	case '}':
		return scanner.MakeToken(TOKEN_RIGHT_BRACE)
	case ';':
		return scanner.MakeToken(TOKEN_SEMICOLON)
	case ',':
		return scanner.MakeToken(TOKEN_COMMA)
	case '.':
		return scanner.MakeToken(TOKEN_DOT)
	case '-':
		return scanner.MakeToken(TOKEN_MINUS)
	case '+':
		return scanner.MakeToken(TOKEN_PLUS)
	case '/':
		return scanner.MakeToken(TOKEN_SLASH)
	case '*':
		return scanner.MakeToken(TOKEN_STAR)
	case '!':
		if scanner.match('=') {
			return scanner.MakeToken(TOKEN_BANG_EQUAL)
		}
		return scanner.MakeToken(TOKEN_BANG)
	case '=':
		if scanner.match('=') {
			return scanner.MakeToken((TOKEN_EQUAL_EQUAL))
		}
		return scanner.MakeToken((TOKEN_EQUAL))
	case '<':
		if scanner.match('=') {
			return scanner.MakeToken((TOKEN_LESS_EQUAL))
		}
		return scanner.MakeToken((TOKEN_LESS))
	case '>':
		if scanner.match('=') {
			return scanner.MakeToken((TOKEN_GREATER_EQUAL))
		}
		return scanner.MakeToken((TOKEN_GREATER))
	}

	return scanner.ErrorToken("Unexpected character.")
}

func DumpToken(token Token) {
	fmt.Printf("%6d: %2d <%s>\n", token.line, token.token_type, token.lexeme)
}

func Scan(source string) {
	scanner := Scanner{1, 0, 0, source}
	for {
		token := scanner.ScanToken()
		if token.token_type == TOKEN_EOF {
			break
		}
		DumpToken(token)
	}
}
