package main

const (
	OP_CONSTANT byte = iota + 1 // 1
	OP_RETURN
)

type BCode struct {
	bcode byte
	line  int
}

type Chunk struct {
	bcodes    []BCode
	constants []float64
}

func WriteChunk(chunk *Chunk, bcode byte, line int) {
	chunk.bcodes = append(chunk.bcodes, BCode{bcode, line})
}

func AddConstant(chunk *Chunk, c float64) int {
	chunk.constants = append(chunk.constants, c)
	return len(chunk.constants) - 1
}
