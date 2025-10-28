package main

import (
	"fmt"
	"os"
	"strconv"
)

const (
	PREC_NONE       byte = iota + 1 // 1
	PREC_ASSIGNMENT                 // =
	PREC_OR                         // or
	PREC_AND                        // and
	PREC_EQUALITY                   // == !=
	PREC_COMPARISON                 // < > <= >=
	PREC_TERM                       // + -
	PREC_FACTOR                     // * /
	PREC_UNARY                      // ! -
	PREC_CALL                       // . ()
	PREC_PRIMARY
)

type ParseFn func(*Parser)

type ParseRule struct {
	prefix     ParseFn
	infix      ParseFn
	precedence byte
}

var rules map[byte]ParseRule

type Parser struct {
	scanner   Scanner
	current   Token
	previous  Token
	hadError  bool
	panicMode bool
	chunk     *Chunk
}

func (parser *Parser) makeConstant(value Value) byte {
	offset := AddConstant(parser.chunk, value)
	if offset > 255 {
		parser.errorAtPrevious("Too many constants in one chunk.")
	}
	return byte(offset)
}

func (parser *Parser) emitByte(b byte) {
	WriteChunk(parser.chunk, b, parser.previous.line)
}

func (parser *Parser) emitBytes(b1 byte, b2 byte) {
	parser.emitByte(b1)
	parser.emitByte(b2)
}

func (parser *Parser) emitConstant(value Value) {
	parser.emitBytes(OP_CONSTANT, parser.makeConstant(value))
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

func getRule(token_type byte) ParseRule {
	return rules[token_type]
}

func (parser *Parser) expression() {
	parser.parsePrecedence(PREC_ASSIGNMENT)
}

func (parser *Parser) number() {
	value, err := strconv.ParseFloat(parser.previous.lexeme, 64)
	if err != nil {
		parser.errorAtPrevious("Failed convert string to number.")
		return
	}
	parser.emitConstant(NewFloat(value))
}

func (parser *Parser) grouping() {
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (parser *Parser) unary() {
	operator_type := parser.previous.token_type
	parser.parsePrecedence(PREC_UNARY)
	switch operator_type {
	case TOKEN_MINUS:
		parser.emitByte(OP_NEGATE)
	case TOKEN_BANG:
		parser.emitByte(OP_NOT)
	}
}

func (parser *Parser) binary() {
	operator_type := parser.previous.token_type
	rule := getRule(operator_type)
	parser.parsePrecedence(rule.precedence + 1)
	switch operator_type {
	case TOKEN_BANG_EQUAL:
		parser.emitBytes(OP_EQUAL, OP_NOT)
	case TOKEN_EQUAL_EQUAL:
		parser.emitByte(OP_EQUAL)
	case TOKEN_GREATER:
		parser.emitByte(OP_GREATER)
	case TOKEN_GREATER_EQUAL:
		parser.emitBytes(OP_LESS, OP_NOT)
	case TOKEN_LESS:
		parser.emitByte(OP_LESS)
	case TOKEN_LESS_EQUAL:
		parser.emitBytes(OP_GREATER, OP_NOT)
	case TOKEN_PLUS:
		parser.emitByte(OP_ADD)
	case TOKEN_MINUS:
		parser.emitByte(OP_SUBTRACT)
	case TOKEN_STAR:
		parser.emitByte(OP_MULTIPLY)
	case TOKEN_SLASH:
		parser.emitByte(OP_DIVIDE)
	}
}

func (parser *Parser) literal() {
	switch parser.previous.token_type {
	case TOKEN_FALSE:
		parser.emitByte(OP_FALSE)
	case TOKEN_TRUE:
		parser.emitByte(OP_TRUE)
	case TOKEN_NIL:
		parser.emitByte(OP_NIL)
	}
}

func (parser *Parser) stringRule() {
	parser.emitConstant(NewString(parser.previous.lexeme[1 : len(parser.previous.lexeme)-1]))
}

func (parser *Parser) parsePrecedence(precedence byte) {
	parser.advance()
	prefix := getRule(parser.previous.token_type).prefix
	if prefix == nil {
		parser.errorAtPrevious("Expect expression.")
		return
	}

	prefix(parser)

	for precedence <= getRule(parser.current.token_type).precedence {
		parser.advance()
		infix := getRule(parser.previous.token_type).infix
		infix(parser)
	}
}

func (parser *Parser) check(token_type byte) bool {
	return parser.current.token_type == token_type
}

func (parser *Parser) match(token_type byte) bool {
	if !parser.check(token_type) {
		return false
	}
	parser.advance()
	return true
}

func (parser *Parser) printStatement() {
	parser.expression()
	parser.consume(TOKEN_SEMICOLON, "Expect ';' after value.")
	parser.emitByte(OP_PRINT)
}

func (parser *Parser) expressionStatement() {
	parser.expression()
	parser.consume(TOKEN_SEMICOLON, "Expect '; after expression.")
	parser.emitByte(OP_POP)
}

func (parser *Parser) statement() {
	if parser.match(TOKEN_PRINT) {
		parser.printStatement()
	} else {
		parser.expressionStatement()
	}
}

func (parser *Parser) synchronize() {
	parser.panicMode = false
	for parser.current.token_type != TOKEN_EOF {
		if parser.previous.token_type == TOKEN_SEMICOLON {
			return
		}
		switch parser.current.token_type {
		case TOKEN_CLASS, TOKEN_FUN, TOKEN_VAR, TOKEN_FOR, TOKEN_WHILE, TOKEN_IF, TOKEN_PRINT, TOKEN_RETURN:
			return
		}
		parser.advance()
	}
}

func (parser *Parser) declaration() {
	parser.statement()
	if parser.panicMode {
		parser.synchronize()
	}
}

func CompilerInit() {
	rules = map[byte]ParseRule{
		TOKEN_LEFT_PAREN:    {(*Parser).grouping, nil, PREC_NONE},
		TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		TOKEN_COMMA:         {nil, nil, PREC_NONE},
		TOKEN_DOT:           {nil, nil, PREC_NONE},
		TOKEN_MINUS:         {(*Parser).unary, (*Parser).binary, PREC_TERM},
		TOKEN_PLUS:          {nil, (*Parser).binary, PREC_TERM},
		TOKEN_SEMICOLON:     {nil, nil, PREC_NONE},
		TOKEN_SLASH:         {nil, (*Parser).binary, PREC_FACTOR},
		TOKEN_STAR:          {nil, (*Parser).binary, PREC_FACTOR},
		TOKEN_BANG:          {(*Parser).unary, nil, PREC_NONE},
		TOKEN_BANG_EQUAL:    {nil, (*Parser).binary, PREC_EQUALITY},
		TOKEN_EQUAL:         {nil, nil, PREC_NONE},
		TOKEN_EQUAL_EQUAL:   {nil, (*Parser).binary, PREC_EQUALITY},
		TOKEN_GREATER:       {nil, (*Parser).binary, PREC_COMPARISON},
		TOKEN_GREATER_EQUAL: {nil, (*Parser).binary, PREC_COMPARISON},
		TOKEN_LESS:          {nil, (*Parser).binary, PREC_COMPARISON},
		TOKEN_LESS_EQUAL:    {nil, (*Parser).binary, PREC_COMPARISON},
		TOKEN_IDENTIFIER:    {nil, nil, PREC_NONE},
		TOKEN_STRING:        {(*Parser).stringRule, nil, PREC_NONE},
		TOKEN_NUMBER:        {(*Parser).number, nil, PREC_NONE},
		TOKEN_AND:           {nil, nil, PREC_NONE},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {(*Parser).literal, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {(*Parser).literal, nil, PREC_NONE},
		TOKEN_OR:            {nil, nil, PREC_NONE},
		TOKEN_PRINT:         {nil, nil, PREC_NONE},
		TOKEN_RETURN:        {nil, nil, PREC_NONE},
		TOKEN_SUPER:         {nil, nil, PREC_NONE},
		TOKEN_THIS:          {nil, nil, PREC_NONE},
		TOKEN_TRUE:          {(*Parser).literal, nil, PREC_NONE},
		TOKEN_VAR:           {nil, nil, PREC_NONE},
		TOKEN_WHILE:         {nil, nil, PREC_NONE},
		TOKEN_ERROR:         {nil, nil, PREC_NONE},
		TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

func Compile(source string) (bool, *Chunk) {
	var chunk Chunk
	parser := Parser{scanner: Scanner{1, 0, 0, source}, chunk: &chunk}
	parser.advance()

	for !parser.match(TOKEN_EOF) {
		parser.declaration()
	}

	if !parser.hadError {
		fmt.Println("Compile success!")
		DisassembleChunk(&chunk, "test chunk")
	}

	return !parser.hadError, &chunk
}
