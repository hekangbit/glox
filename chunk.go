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
