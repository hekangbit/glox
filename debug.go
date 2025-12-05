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
	fmt.Printf("%-16s %4d\n", name, chunk.bcodes[offset+1])
	return offset + 2
}

func JumpInstruction(name string, sign int, chunk *Chunk, offset int) int {
	var jump uint16 = uint16(chunk.bcodes[offset+1])<<8 + uint16(chunk.bcodes[offset+2])
	fmt.Printf("%-16s %4d -> %d\n", name, offset, offset+3+sign*int(jump))
	return offset + 3
}

func InvokeInstruction(name string, chunk *Chunk, offset int) int {
	constant_index := chunk.bcodes[offset+1]
	argCount := chunk.bcodes[offset+2]
	fmt.Printf("%-16s (%d args) %4d '", name, argCount, constant_index)
	fmt.Printf("%v'\n", chunk.constants[constant_index])
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
	case OP_GET_UPVALUE:
		return ByteInstruction("OP_GET_UPVALUE", chunk, offset)
	case OP_SET_UPVALUE:
		return ByteInstruction("OP_SET_UPVALUE", chunk, offset)
	case OP_JUMP:
		return JumpInstruction("OP_JUMP", 1, chunk, offset)
	case OP_JUMP_IF_FALSE:
		return JumpInstruction("OP_JUMP_IF_FALSE", 1, chunk, offset)
	case OP_LOOP:
		return JumpInstruction("OP_LOOP", -1, chunk, offset)
	case OP_CALL:
		return ByteInstruction("OP_CALL", chunk, offset)
	case OP_CLOSURE:
		offset++
		constant := chunk.bcodes[offset]
		offset++
		fmt.Printf("%-16s %4d ", "OP_CLOSURE", constant)
		fmt.Printf("%s", chunk.constants[constant].String())
		fmt.Printf("\n")
		function, _ := chunk.constants[constant].GetFunction()
		for i := 0; i < function.upValueCount; i++ {
			isLocal := chunk.bcodes[offset]
			index := chunk.bcodes[offset+1]
			var msg string = "upvalue"
			if isLocal != 0 {
				msg = "local"
			}
			fmt.Printf("%04d      |                     %s %d\n", offset, msg, index)
			offset += 2
		}
		return offset
	case OP_CLOSE_UPVALUE:
		return SimpleInstruction("OP_CLOSE_UPVALUE", offset)
	case OP_CLASS:
		return ConstInstruction("OP_CLASS", chunk, offset)
	case OP_GET_PROPERTY:
		return ConstInstruction("OP_GET_PREPERTY", chunk, offset)
	case OP_SET_PROPERTY:
		return ConstInstruction("OP_SET_PROPERTY", chunk, offset)
	case OP_METHOD:
		return ConstInstruction("OP_METHOD", chunk, offset)
	case OP_INVOKE:
		return InvokeInstruction("OP_INVOKE", chunk, offset)
	case OP_INHERIT:
		return SimpleInstruction("OP_INHERIT", offset)
	case OP_GET_SUPER:
		return ConstInstruction("OP_GET_SUPER", chunk, offset)
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
	frame := &vm.frames[vm.frameCount-1]
	fmt.Print("          ")
	for i := 0; i < vm.vstackCount; i++ {
		if i == frame.slots_base {
			fmt.Print("^")
		}
		fmt.Print("[ ")
		fmt.Printf("%s", vm.vstack[i].String())
		fmt.Print(" ]")
	}
	fmt.Print("\n")
	DisassembleInstruction(&frame.closure.function.chunk, frame.ip)
}
