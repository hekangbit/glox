package main

const (
	OP_CONSTANT byte = iota + 1 // 1
	OP_NEGATE
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
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
