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
	TOKEN_BREAK
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

var keyword map[string]int

func ScannerInit() {
	keyword = make(map[string]int)
	keyword["and"] = TOKEN_AND
	keyword["or"] = TOKEN_OR
	keyword["true"] = TOKEN_TRUE
	keyword["false"] = TOKEN_FALSE
	keyword["if"] = TOKEN_IF
	keyword["else"] = TOKEN_ELSE
	keyword["for"] = TOKEN_FOR
	keyword["while"] = TOKEN_WHILE
	keyword["var"] = TOKEN_VAR
	keyword["this"] = TOKEN_THIS
	keyword["super"] = TOKEN_SUPER
	keyword["return"] = TOKEN_RETURN
	keyword["print"] = TOKEN_PRINT
	keyword["fun"] = TOKEN_FUN
	keyword["class"] = TOKEN_CLASS
	keyword["break"] = TOKEN_BREAK
	keyword["nil"] = TOKEN_NIL
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') && c == '_'
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
	return scanner.current >= len(scanner.source)
}

func (scanner *Scanner) advance() byte {
	result := scanner.source[scanner.current]
	scanner.current++
	return result
}

func (scanner *Scanner) peek() byte {
	if scanner.current >= len(scanner.source) {
		return 0
	}
	return scanner.source[scanner.current]
}

func (scanner *Scanner) peekNext() byte {
	if scanner.current >= len(scanner.source)-1 {
		return 0
	}
	return scanner.source[scanner.current+1]
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
		case '/':
			if scanner.peekNext() != '/' {
				return
			}
			for {
				if scanner.peek() != '\n' && !scanner.isAtEnd() {
					scanner.advance()
				}
			}
		default:
			return
		}
	}
}

func (scanner *Scanner) stringLiteral() Token {
	for scanner.peek() != '"' && !scanner.isAtEnd() {
		if scanner.peek() == '\n' {
			scanner.line++
		}
		scanner.advance()
	}

	if scanner.isAtEnd() {
		return scanner.ErrorToken("Unterminated string.")
	}

	scanner.advance()

	// str := scanner.source[scanner.start+1 : scanner.current-1]
	// return Token{TOKEN_STRING, str, scanner.line}
	return scanner.MakeToken(TOKEN_STRING)
}

func (scanner *Scanner) numberLiteral() Token {
	for isDigit(scanner.peek()) {
		scanner.advance()
	}
	if scanner.peek() == '.' && isDigit(scanner.peekNext()) {
		scanner.advance()
		for isDigit(scanner.peek()) {
			scanner.advance()
		}
	}
	return scanner.MakeToken(TOKEN_NUMBER)
}

func (scanner *Scanner) identiferToken() Token {
	for isAlpha(scanner.peek()) || isDigit(scanner.peek()) {
		scanner.advance()
	}

	identifer := scanner.source[scanner.start:scanner.current]

	if token_type, ok := keyword[identifer]; ok {
		return scanner.MakeToken(token_type)
	}

	return scanner.MakeToken(TOKEN_IDENTIFIER)
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
	scanner.skipWhitespace()

	scanner.start = scanner.current
	if scanner.isAtEnd() {
		return scanner.EOFToken()
	}

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
	case '"':
		return scanner.stringLiteral()
	}

	if isDigit(c) {
		return scanner.numberLiteral()
	}

	if isAlpha(c) {
		return scanner.identiferToken()
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
