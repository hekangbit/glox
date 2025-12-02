package main

const (
	OP_CONSTANT byte = iota + 1 // 1
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_NOT
	OP_NEGATE
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_PRINT
	OP_POP
	OP_DEFINE_GLOBAL
	OP_GET_GLOBAL
	OP_SET_GLOBAL
	OP_GET_LOCAL
	OP_SET_LOCAL
	OP_JUMP
	OP_JUMP_IF_FALSE
	OP_LOOP
	OP_CALL
	OP_CLOSURE
	OP_GET_UPVALUE
	OP_SET_UPVALUE
	OP_CLOSE_UPVALUE
	OP_CLASS
	OP_SET_PROPERTY
	OP_GET_PROPERTY
	OP_METHOD
	OP_INVOKE
	OP_RETURN
)

type Chunk struct {
	bcodes    []byte
	lines     []int
	constants []Value
}

func WriteChunk(chunk *Chunk, bcode byte, line int) {
	chunk.bcodes = append(chunk.bcodes, bcode)
	chunk.lines = append(chunk.lines, line)
}

func AddConstant(chunk *Chunk, c Value) int {
	chunk.constants = append(chunk.constants, c)
	return len(chunk.constants) - 1
}
