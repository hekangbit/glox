package main

import (
	"fmt"
	"os"
)

type Parser struct {
	scanner   Scanner
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
	chunk     *Chunk
}

func (parser *Parser) emitByte(b byte) {
	WriteChunk(parser.chunk, b, parser.previous.line)
}

func (parser *Parser) emitBytes(b1 byte, b2 byte) {
	parser.emitByte(b1)
	parser.emitByte(b2)
}

func (parser *Parser) emitReturn() {
	parser.emitByte(OP_RETURN)
}

func (parser *Parser) errorAt(token *Token, message string) {
	if parser.panicMode {
		return
	}
	parser.panicMode = true
	fmt.Fprintf(os.Stderr, "[line %d] Error", token.line)
	switch token.token_type {
	case TOKEN_EOF:
		fmt.Fprintf(os.Stderr, " at end")
	case TOKEN_ERROR:
		// Nothing.
	default:
		fmt.Fprintf(os.Stderr, " at '%s'", token.lexeme)
	}

	fmt.Fprintf(os.Stderr, ": %s\n", message)
	parser.hadError = true
}

func (parser *Parser) errorAtCurrent(message string) {
	parser.errorAt(&parser.current, message)
}

func (parser *Parser) errorAtPrevious(message string) {
	parser.errorAt(&parser.previous, message)
}

func (parser *Parser) advance() {
	parser.previous = parser.current
	for {
		parser.current = parser.scanner.ScanToken()
		if parser.current.token_type != TOKEN_ERROR {
			break
		}
		parser.errorAtCurrent(parser.current.lexeme)
	}
}

func (parser *Parser) consume(token_type byte, message string) {
	if parser.current.token_type == token_type {
		parser.advance()
		return
	}
	parser.errorAtCurrent(message)
}

func (parser *Parser) expression() {
	// TODO
}

func Compile(source string) (bool, *Chunk) {
	var chunk Chunk
	parser := Parser{scanner: Scanner{1, 0, 0, source}, chunk: &chunk}
	parser.advance()
	parser.expression()
	parser.consume(TOKEN_EOF, "Expect end of expression.")

	// temp
	AddConstant(&chunk, 12.345)
	parser.emitByte(OP_CONSTANT)
	parser.emitByte(0)

	parser.emitReturn()

	return !parser.hadError, &chunk
}
