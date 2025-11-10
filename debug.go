package main

import "fmt"

var DebugFlag bool = false

func SimpleInstruction(name string, offset int) int {
	fmt.Printf("%s\n", name)
	return offset + 1
}

func ConstInstruction(name string, chunk *Chunk, offset int) int {
	constant_index := chunk.bcodes[offset+1]
	fmt.Printf("%-16s %4d '", name, constant_index)
	fmt.Printf("%v'\n", chunk.constants[constant_index])
	return offset + 2
}

func ByteInstruction(name string, chunk *Chunk, offset int) int {
	slot_index := chunk.bcodes[offset+1]
	fmt.Printf("%-16s %4d\n", name, slot_index)
	return offset + 2
}

func JumpInstruction(name string, sign int, chunk *Chunk, offset int) int {
	var jump uint16 = uint16(chunk.bcodes[offset+1]<<8 + chunk.bcodes[offset+2])
	fmt.Printf("%-16s %4d -> %d\n", name, offset, offset+3+sign*int(jump))
	return offset + 3
}

func DisassembleInstruction(chunk *Chunk, offset int) int {
	fmt.Printf("%04d ", offset)

	instruction := chunk.bcodes[offset]
	switch instruction {
	case OP_CONSTANT:
		return ConstInstruction("OP_CONSTANT", chunk, offset)
	case OP_NIL:
		return SimpleInstruction("OP_NIL", offset)
	case OP_FALSE:
		return SimpleInstruction("OP_FALSE", offset)
	case OP_TRUE:
		return SimpleInstruction("OP_TRUE", offset)
	case OP_NOT:
		return SimpleInstruction("OP_NOT", offset)
	case OP_NEGATE:
		return SimpleInstruction("OP_NEGATE", offset)
	case OP_EQUAL:
		return SimpleInstruction("OP_EQUAL", offset)
	case OP_GREATER:
		return SimpleInstruction("OP_GREATER", offset)
	case OP_LESS:
		return SimpleInstruction("OP_LESS", offset)
	case OP_ADD:
		return SimpleInstruction("OP_ADD", offset)
	case OP_SUBTRACT:
		return SimpleInstruction("OP_SUBTRACT", offset)
	case OP_MULTIPLY:
		return SimpleInstruction("OP_MULTIPLY", offset)
	case OP_DIVIDE:
		return SimpleInstruction("OP_DIVIDE", offset)
	case OP_RETURN:
		return SimpleInstruction("OP_RETURN", offset)
	case OP_PRINT:
		return SimpleInstruction("OP_PRINT", offset)
	case OP_POP:
		return SimpleInstruction("OP_POP", offset)
	case OP_DEFINE_GLOBAL:
		return ConstInstruction("OP_DEFINE_GLOBAL", chunk, offset)
	case OP_GET_GLOBAL:
		return ConstInstruction("OP_GET_GLOBAL", chunk, offset)
	case OP_SET_GLOBAL:
		return ConstInstruction("OP_SET_GLOBAL", chunk, offset)
	case OP_GET_LOCAL:
		return ByteInstruction("OP_GET_LOCAL", chunk, offset)
	case OP_SET_LOCAL:
		return ByteInstruction("OP_SET_LOCAL", chunk, offset)
	case OP_JUMP:
		return JumpInstruction("OP_JUMP", 1, chunk, offset)
	case OP_JUMP_IF_FALSE:
		return JumpInstruction("OP_JUMP_IF_FALSE", 1, chunk, offset)
	case OP_LOOP:
		return JumpInstruction("OP_LOOP", -1, chunk, offset)
	default:
		fmt.Printf("Unknown opcode %v\n", instruction)
		return offset + 1
	}
}

func DisassembleChunk(chunk *Chunk, name string) {
	if !DebugFlag {
		return
	}
	fmt.Printf("== %s ==\n", name)
	for offset := 0; offset < len(chunk.bcodes); {
		offset = DisassembleInstruction(chunk, offset)
	}
}

func DebugVM(vm *VM) {
	if !DebugFlag {
		return
	}
	fmt.Print("          ")
	for _, v := range vm.vstack {
		fmt.Print("[ ")
		fmt.Printf("%s", v.String())
		fmt.Print(" ]")
	}
	fmt.Print("\n")
	frame := &vm.frames[vm.frameCount-1]
	DisassembleInstruction(&frame.function.chunk, frame.ip)
}
