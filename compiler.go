package main

import (
	"fmt"
	"math"
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

const (
	FN_TYPE_SCRIPT int = iota + 0
	FN_TYPE_FUNCTION
	FN_TYPE_METHOD
)

const (
	MAX_NUM_OF_LOCAL_VARS int = math.MaxUint8
	MAX_NUM_OF_UP_VALS    int = math.MaxUint8
)

type ParseFn func(*Parser, bool)

type ParseRule struct {
	prefix     ParseFn
	infix      ParseFn
	precedence byte
}

type Parser struct {
	scanner      Scanner
	current      Token
	previous     Token
	rules        map[byte]ParseRule
	hadError     bool
	panicMode    bool
	compiler     *Compiler
	currentClass *ClassCompiler
}

type Local struct {
	name       Token
	depth      int
	isCaptured bool
}

type UpValue struct {
	index   byte
	isLocal bool
}

type Compiler struct {
	locals     [MAX_NUM_OF_LOCAL_VARS]Local // all locals that are in scope during each point in the compilation process
	scopeDepth int                          // the number of blocks surrounding the current bit of code we’re compiling
	localCount int
	upValues   [MAX_NUM_OF_UP_VALS]UpValue
	function   *LoxFunction
	fnType     int
	enclosing  *Compiler
}

type ClassCompiler struct {
	enclosing *ClassCompiler
}

func identifiersEqual(a *Token, b *Token) bool {
	return a.lexeme == b.lexeme
}

func (parser *Parser) getRule(token_type byte) ParseRule {
	return parser.rules[token_type]
}

func (parser *Parser) currentChunk() *Chunk {
	return &parser.compiler.function.chunk
}

func (parser *Parser) currentChunkSize() int {
	return len(parser.currentChunk().bcodes)
}

func (parser *Parser) makeConstant(value Value) byte {
	offset := AddConstant(parser.currentChunk(), value)
	if offset > 255 {
		parser.errorAtPrevious("Too many constants in one chunk.")
	}
	return byte(offset)
}

func (parser *Parser) emitByte(b byte) {
	WriteChunk(parser.currentChunk(), b, parser.previous.line)
}

func (parser *Parser) emitBytes(b1 byte, b2 byte) {
	parser.emitByte(b1)
	parser.emitByte(b2)
}

func (parser *Parser) emitConstant(value Value) {
	parser.emitBytes(OP_CONSTANT, parser.makeConstant(value))
}

func (parser *Parser) emitReturn() {
	parser.emitByte(OP_NIL)
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

func (parser *Parser) parsePrecedence(precedence byte) {
	parser.advance()
	prefix := parser.getRule(parser.previous.token_type).prefix
	if prefix == nil {
		parser.errorAtPrevious("Expect expression.")
		return
	}

	canAssign := precedence <= PREC_ASSIGNMENT
	prefix(parser, canAssign)

	for precedence <= parser.getRule(parser.current.token_type).precedence {
		parser.advance()
		infix := parser.getRule(parser.previous.token_type).infix
		infix(parser, canAssign)
	}

	if canAssign && parser.match(TOKEN_EQUAL) {
		parser.errorAtPrevious("Invalid assignment target.")
	}
}

func (parser *Parser) expression() {
	parser.parsePrecedence(PREC_ASSIGNMENT)
}

func (parser *Parser) number(canAssign bool) {
	value, err := strconv.ParseFloat(parser.previous.lexeme, 64)
	if err != nil {
		parser.errorAtPrevious("Failed convert string to number.")
		return
	}
	parser.emitConstant(FloatVal(value))
}

func (parser *Parser) grouping(canAssign bool) {
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after expression.")
}

func (parser *Parser) unary(canAssign bool) {
	operator_type := parser.previous.token_type
	parser.parsePrecedence(PREC_UNARY)
	switch operator_type {
	case TOKEN_MINUS:
		parser.emitByte(OP_NEGATE)
	case TOKEN_BANG:
		parser.emitByte(OP_NOT)
	}
}

func (parser *Parser) binary(canAssign bool) {
	operator_type := parser.previous.token_type
	rule := parser.getRule(operator_type)
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

func (parser *Parser) andRule(canAssign bool) {
	endJump := parser.emitJump(OP_JUMP_IF_FALSE)
	parser.emitByte(OP_POP)
	parser.parsePrecedence(PREC_AND)
	parser.patchJump(endJump)
}

func (parser *Parser) orRule(canAssign bool) {
	elseJump := parser.emitJump(OP_JUMP_IF_FALSE)
	endJump := parser.emitJump(OP_JUMP)
	parser.patchJump(elseJump)
	parser.emitByte(OP_POP)
	parser.parsePrecedence(PREC_OR)
	parser.patchJump(endJump)
}

func (parser *Parser) boolLiteral(canAssign bool) {
	switch parser.previous.token_type {
	case TOKEN_FALSE:
		parser.emitByte(OP_FALSE)
	case TOKEN_TRUE:
		parser.emitByte(OP_TRUE)
	case TOKEN_NIL:
		parser.emitByte(OP_NIL)
	}
}

func (parser *Parser) stringLiteral(canAssign bool) {
	parser.emitConstant(StringVal(parser.previous.lexeme[1 : len(parser.previous.lexeme)-1]))
}

func (parser *Parser) resolveLocal(compiler *Compiler, name *Token) (byte, bool) {
	for i := compiler.localCount - 1; i >= 0; i-- {
		local := &compiler.locals[i]
		if identifiersEqual(name, &local.name) {
			if local.depth == -1 {
				parser.errorAtPrevious("Can't read local variable in its own initializer.")
			}
			return byte(i), true
		}
	}
	return 0, false
}

func (parser *Parser) addUpvalue(compiler *Compiler, index byte, isLocal bool) (byte, bool) {
	for i := 0; i < compiler.function.upValueCount; i++ {
		if compiler.upValues[i].index == index && compiler.upValues[i].isLocal == isLocal {
			return byte(i), true
		}
	}
	upValueCount := compiler.function.upValueCount

	if upValueCount >= MAX_NUM_OF_UP_VALS {
		parser.errorAtPrevious("Too many closure variables in function.")
		return 0, false
	}
	compiler.upValues[upValueCount].isLocal = isLocal
	compiler.upValues[upValueCount].index = index
	compiler.function.upValueCount++
	return byte(upValueCount), true
}

func (parser *Parser) resolveUpvalue(compiler *Compiler, name *Token) (byte, bool) {
	var local byte
	var upvalue byte
	var ok bool
	if compiler.enclosing == nil {
		return 0, false
	}
	local, ok = parser.resolveLocal(compiler.enclosing, name)
	if ok {
		compiler.enclosing.locals[local].isCaptured = true
		return parser.addUpvalue(compiler, local, true)
	}
	upvalue, ok = parser.resolveUpvalue(compiler.enclosing, name)
	if ok {
		return parser.addUpvalue(compiler, upvalue, false)
	}
	return 0, false
}

func (parser *Parser) namedVariable(name *Token, canAssign bool) {
	var getOP, setOP byte
	var arg byte = 0
	var ok bool = false

	if arg, ok = parser.resolveLocal(parser.compiler, name); ok {
		getOP = OP_GET_LOCAL
		setOP = OP_SET_LOCAL
	} else if arg, ok = parser.resolveUpvalue(parser.compiler, name); ok {
		getOP = OP_GET_UPVALUE
		setOP = OP_SET_UPVALUE
	} else {
		arg = parser.identifierConstant(name)
		getOP = OP_GET_GLOBAL
		setOP = OP_SET_GLOBAL
	}

	if canAssign && parser.match(TOKEN_EQUAL) {
		parser.expression()
		parser.emitBytes(setOP, arg)
	} else {
		parser.emitBytes(getOP, arg)
	}
}

func (parser *Parser) variable(canAssign bool) {
	parser.namedVariable(&parser.previous, canAssign)
}

func (parser *Parser) argumentList() byte {
	var argCount byte = 0
	if !parser.check(TOKEN_RIGHT_PAREN) {
		for {
			parser.expression()
			if argCount == 255 {
				parser.errorAtPrevious("Can't have more than 255 arguments.")
			}
			argCount++
			if !parser.match(TOKEN_COMMA) {
				break
			}
		}
	}
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after argument list.")
	return argCount
}

func (parser *Parser) call(canAssign bool) {
	argCount := parser.argumentList()
	parser.emitBytes(OP_CALL, argCount)
}

func (parser *Parser) dot(canAssign bool) {
	parser.consume(TOKEN_IDENTIFIER, "Expect property name after '.'.")
	name := parser.identifierConstant(&parser.previous)
	if canAssign && parser.match(TOKEN_EQUAL) {
		parser.expression()
		parser.emitBytes(OP_SET_PROPERTY, name)
	} else {
		parser.emitBytes(OP_GET_PROPERTY, name)
	}
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

func (parser *Parser) block() {
	for !parser.check(TOKEN_RIGHT_BRACE) && !parser.check(TOKEN_EOF) {
		parser.declaration()
	}
	parser.consume(TOKEN_RIGHT_BRACE, "Expect '}' after block.")
}

func (parser *Parser) beginScope() {
	parser.compiler.scopeDepth++
}

func (parser *Parser) endScope() {
	parser.compiler.scopeDepth--
	for parser.compiler.localCount > 0 && parser.compiler.locals[parser.compiler.localCount-1].depth > parser.compiler.scopeDepth {
		if parser.compiler.locals[parser.compiler.localCount-1].isCaptured {
			parser.emitByte(OP_CLOSE_UPVALUE)
		} else {
			parser.emitByte(OP_POP)
		}
		parser.compiler.localCount--
	}
}

func (parser *Parser) emitJump(op byte) int {
	parser.emitByte(op)
	parser.emitByte(0xFF)
	parser.emitByte(0xFF)
	return parser.currentChunkSize() - 2
}

func (parser *Parser) patchJump(offset int) {
	jump := parser.currentChunkSize() - offset - 2
	if jump > math.MaxUint16 {
		parser.errorAtPrevious("Too much code to jump over.")
	}
	// Big-endian
	chunk := parser.currentChunk()
	chunk.bcodes[offset] = byte((jump >> 8) & 0xFF)
	chunk.bcodes[offset+1] = byte(jump & 0xFF)
}

/*
OP_JUMP_IF_FALSE
U16
OP_POP
statements(then)
OP_JUMP
U16
OP_POP
statements(else)
statements(after if statement)
*/
func (parser *Parser) ifStatement() {
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'if'.")
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")
	elseJump := parser.emitJump(OP_JUMP_IF_FALSE)
	parser.emitByte(OP_POP)
	parser.statement()
	endJump := parser.emitJump(OP_JUMP)
	parser.patchJump(elseJump)
	parser.emitByte(OP_POP)
	if parser.match(TOKEN_ELSE) {
		parser.statement()
	}
	parser.patchJump(endJump)
}

func (parser *Parser) emitLoop(loopStart int) {
	parser.emitByte(OP_LOOP)
	offset := parser.currentChunkSize() - loopStart + 2
	if offset > math.MaxUint16 {
		parser.errorAtPrevious("Loop body too large.")
	}
	parser.emitByte(byte(offset >> 8 & 0xFF))
	parser.emitByte(byte(offset & 0xFF))
}

func (parser *Parser) whileStatement() {
	loopStart := parser.currentChunkSize()
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'while'.")
	parser.expression()
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after condition.")

	exitJump := parser.emitJump(OP_JUMP_IF_FALSE)
	parser.emitByte(OP_POP)
	parser.statement()
	parser.emitLoop(loopStart)

	parser.patchJump(exitJump)
	parser.emitByte(OP_POP)
}

func (parser *Parser) forStatement() {
	parser.beginScope()

	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after 'for'.")

	if parser.match(TOKEN_SEMICOLON) {
		// no init
	} else if parser.match(TOKEN_VAR) {
		parser.varDeclaration()
	} else {
		parser.expressionStatement()
	}

	var loopStart int = parser.currentChunkSize()
	var exitJump int = -1

	if !parser.match(TOKEN_SEMICOLON) {
		parser.expression()
		parser.consume(TOKEN_SEMICOLON, "Expect ';' after for loop condition.")
		exitJump = parser.emitJump(OP_JUMP_IF_FALSE)
		parser.emitByte(OP_POP)
	}

	if !parser.match(TOKEN_RIGHT_PAREN) {
		bodyJump := parser.emitJump(OP_JUMP)
		loopIncrement := parser.currentChunkSize()
		parser.expression()
		parser.emitByte(OP_POP)
		parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after for loop clauses.")
		parser.emitLoop(loopStart)
		parser.patchJump(bodyJump)
		loopStart = loopIncrement
	}

	parser.statement()
	parser.emitLoop(loopStart)

	if exitJump != -1 {
		parser.patchJump(exitJump)
		parser.emitByte(OP_POP)
	}

	parser.endScope()
}

func (parser *Parser) returnStatement() {
	if parser.compiler.fnType == FN_TYPE_SCRIPT {
		parser.errorAtPrevious("Can't return from top-level code.")
	}
	if parser.match(TOKEN_SEMICOLON) {
		parser.emitReturn()
	} else {
		parser.expression()
		parser.consume(TOKEN_SEMICOLON, "Expect ';' after return value.")
		parser.emitByte(OP_RETURN)
	}
}

func (parser *Parser) statement() {
	if parser.match(TOKEN_PRINT) {
		parser.printStatement()
	} else if parser.match(TOKEN_LEFT_BRACE) {
		parser.beginScope()
		parser.block()
		parser.endScope()
	} else if parser.match(TOKEN_IF) {
		parser.ifStatement()
	} else if parser.match(TOKEN_WHILE) {
		parser.whileStatement()
	} else if parser.match(TOKEN_FOR) {
		parser.forStatement()
	} else if parser.match(TOKEN_RETURN) {
		parser.returnStatement()
	} else {
		parser.expressionStatement()
	}
}

func (parser *Parser) identifierConstant(token *Token) byte {
	return parser.makeConstant(StringVal(token.lexeme))
}

func (parser *Parser) addLocal(name *Token) {
	if parser.compiler.localCount >= MAX_NUM_OF_LOCAL_VARS {
		parser.errorAtPrevious("Too many local variables in function.")
		return
	}
	local := &parser.compiler.locals[parser.compiler.localCount]
	parser.compiler.localCount++
	local.name = *name
	local.depth = -1
	local.isCaptured = false
}

func (parser *Parser) declareVariable() {
	if parser.compiler.scopeDepth == 0 {
		return
	}

	// check local var duplicate declare
	for i := parser.compiler.localCount - 1; i >= 0; i++ {
		local := &parser.compiler.locals[i]
		if local.depth != -1 && local.depth < parser.compiler.scopeDepth {
			break
		}
		if identifiersEqual(&local.name, &parser.previous) {
			parser.errorAtPrevious("Already a variable with this name in this scope.")
		}
	}

	parser.addLocal(&parser.previous)
}

func (parser *Parser) parseVariable(error_msg string) byte {
	parser.consume(TOKEN_IDENTIFIER, error_msg)

	parser.declareVariable()
	if parser.compiler.scopeDepth > 0 {
		return 0
	}
	return parser.identifierConstant(&parser.previous)
}

func (parser *Parser) markInitialized() {
	if parser.compiler.scopeDepth == 0 {
		return
	}
	parser.compiler.locals[parser.compiler.localCount-1].depth = parser.compiler.scopeDepth
}

func (parser *Parser) defineVariable(global byte) {
	if parser.compiler.scopeDepth > 0 {
		parser.markInitialized()
		return
	}
	parser.emitBytes(OP_DEFINE_GLOBAL, global)
}

func (parser *Parser) varDeclaration() {
	var global byte = parser.parseVariable("Expect variable name.")
	if parser.match(TOKEN_EQUAL) {
		parser.expression()
	} else {
		parser.emitByte(OP_NIL) // push a default value to vstack
	}
	parser.consume(TOKEN_SEMICOLON, "Expect ';' after variable declaration.")
	parser.defineVariable(global)
}

func (parser *Parser) function(fnType int) {
	var compiler Compiler
	parser.initCompiler(&compiler, fnType)
	parser.beginScope()
	parser.consume(TOKEN_LEFT_PAREN, "Expect '(' after function name.")
	if !parser.check(TOKEN_RIGHT_PAREN) {
		for {
			parser.compiler.function.arity++
			if parser.compiler.function.arity > 255 {
				parser.errorAtCurrent("Can't have more than 255 paramenters.")
			}
			constant := parser.parseVariable("Expect parameter name.")
			parser.defineVariable(constant)
			if !parser.match(TOKEN_COMMA) {
				break
			}
		}
	}
	parser.consume(TOKEN_RIGHT_PAREN, "Expect ')' after paramenters.")
	parser.consume(TOKEN_LEFT_BRACE, "Expect '{' before function body.")
	parser.block()

	function := parser.endCompiler()
	parser.emitBytes(OP_CLOSURE, parser.makeConstant(FunctionVal(function))) // add function obj to bcode

	for i := 0; i < function.upValueCount; i++ {
		var islocal byte = 0
		if compiler.upValues[i].isLocal {
			islocal = 1
		}
		parser.emitByte(islocal)
		parser.emitByte(compiler.upValues[i].index)
	}
}

func (parser *Parser) functionDeclaration() {
	global := parser.parseVariable("Expect function name.")
	parser.markInitialized()
	parser.function(FN_TYPE_FUNCTION)
	parser.defineVariable(global)
}

func (parser *Parser) method() {
	parser.consume(TOKEN_IDENTIFIER, "Expect method name.")
	constant := parser.identifierConstant(&parser.previous)
	parser.function(FN_TYPE_METHOD)
	parser.emitBytes(OP_METHOD, constant)
}

func (parser *Parser) classDeclaration() {
	parser.consume(TOKEN_IDENTIFIER, "Expect class name.")
	classToken := &parser.previous
	nameConstant := parser.identifierConstant(&parser.previous)
	parser.declareVariable()
	parser.emitBytes(OP_CLASS, nameConstant)
	parser.defineVariable(nameConstant)
	classCompiler := ClassCompiler{}
	classCompiler.enclosing = parser.currentClass
	parser.currentClass = &classCompiler
	parser.namedVariable(classToken, false)
	parser.consume(TOKEN_LEFT_BRACE, "Expect '{' before class body.")
	for !parser.check(TOKEN_RIGHT_BRACE) && !parser.check(TOKEN_EOF) {
		parser.method()
	}
	parser.emitByte(OP_POP)
	parser.consume(TOKEN_RIGHT_BRACE, "Expect '}' after class body.")
	parser.currentClass = classCompiler.enclosing
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
	if parser.match(TOKEN_VAR) {
		parser.varDeclaration()
	} else if parser.match(TOKEN_FUN) {
		parser.functionDeclaration()
	} else if parser.match(TOKEN_CLASS) {
		parser.classDeclaration()
	} else {
		parser.statement()
	}
	if parser.panicMode {
		parser.synchronize()
	}
}

func (parser *Parser) thisExpr(canAssign bool) {
	if parser.currentClass == nil {
		parser.errorAtPrevious("Can't use 'this' outside of a class.")
		return
	}
	parser.variable(false)
}

func (parser *Parser) initParseRule() {
	parser.rules = map[byte]ParseRule{
		TOKEN_LEFT_PAREN:    {(*Parser).grouping, (*Parser).call, PREC_CALL},
		TOKEN_RIGHT_PAREN:   {nil, nil, PREC_NONE},
		TOKEN_LEFT_BRACE:    {nil, nil, PREC_NONE},
		TOKEN_RIGHT_BRACE:   {nil, nil, PREC_NONE},
		TOKEN_COMMA:         {nil, nil, PREC_NONE},
		TOKEN_DOT:           {nil, (*Parser).dot, PREC_CALL},
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
		TOKEN_IDENTIFIER:    {(*Parser).variable, nil, PREC_NONE},
		TOKEN_STRING:        {(*Parser).stringLiteral, nil, PREC_NONE},
		TOKEN_NUMBER:        {(*Parser).number, nil, PREC_NONE},
		TOKEN_AND:           {nil, (*Parser).andRule, PREC_AND},
		TOKEN_CLASS:         {nil, nil, PREC_NONE},
		TOKEN_ELSE:          {nil, nil, PREC_NONE},
		TOKEN_FALSE:         {(*Parser).boolLiteral, nil, PREC_NONE},
		TOKEN_FOR:           {nil, nil, PREC_NONE},
		TOKEN_FUN:           {nil, nil, PREC_NONE},
		TOKEN_IF:            {nil, nil, PREC_NONE},
		TOKEN_NIL:           {(*Parser).boolLiteral, nil, PREC_NONE},
		TOKEN_OR:            {nil, (*Parser).orRule, PREC_OR},
		TOKEN_PRINT:         {nil, nil, PREC_NONE},
		TOKEN_RETURN:        {nil, nil, PREC_NONE},
		TOKEN_SUPER:         {nil, nil, PREC_NONE},
		TOKEN_THIS:          {(*Parser).thisExpr, nil, PREC_NONE},
		TOKEN_TRUE:          {(*Parser).boolLiteral, nil, PREC_NONE},
		TOKEN_VAR:           {nil, nil, PREC_NONE},
		TOKEN_WHILE:         {nil, nil, PREC_NONE},
		TOKEN_ERROR:         {nil, nil, PREC_NONE},
		TOKEN_EOF:           {nil, nil, PREC_NONE},
	}
}

func (parser *Parser) initCompiler(compiler *Compiler, fnType int) {
	compiler.fnType = fnType
	compiler.function = NewFunction()
	compiler.scopeDepth = 0
	compiler.localCount = 0

	if fnType != FN_TYPE_SCRIPT {
		compiler.function.name = parser.previous.lexeme
	}

	local := &compiler.locals[compiler.localCount]
	compiler.localCount++
	local.depth = 0
	local.isCaptured = false
	if fnType == FN_TYPE_FUNCTION {
		local.name.lexeme = ""
	} else {
		local.name.lexeme = "this"
	}

	compiler.enclosing = parser.compiler
	parser.compiler = compiler
}

func (parser *Parser) endCompiler() *LoxFunction {
	parser.emitReturn()
	function := parser.compiler.function
	if !parser.hadError {
		DisassembleChunk(parser.currentChunk(), NormalizedFuncName(function.name))
	}
	parser.compiler = parser.compiler.enclosing
	return function
}

func Compile(source string) (bool, *LoxFunction) {
	var compiler Compiler
	parser := Parser{scanner: Scanner{1, 0, 0, source}, hadError: false, panicMode: false, currentClass: nil}
	parser.advance()

	parser.initParseRule()
	parser.initCompiler(&compiler, FN_TYPE_SCRIPT)
	for !parser.match(TOKEN_EOF) {
		parser.declaration()
	}
	function := parser.endCompiler()

	return !parser.hadError, function
}
