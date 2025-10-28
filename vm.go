package main

import (
	"fmt"
	"os"
)

type VM struct {
	chunk  *Chunk
	ip     int
	vstack []Value
}

func isfalsey(value Value) bool {
	if value.IsNil() {
		return true
	}
	if value.IsBool() {
		v, _ := value.GetBool()
		return !v
	}
	return false
}

func (vm *VM) pushVstack(value Value) {
	vm.vstack = append(vm.vstack, value)
}

func (vm *VM) popVstack() Value {
	value := vm.vstack[len(vm.vstack)-1]
	vm.vstack = vm.vstack[:len(vm.vstack)-1]
	return value
}

func (vm *VM) peekVstack(offset int) Value {
	return vm.vstack[len(vm.vstack)-1-offset]
}

func (vm *VM) sizeVstack() int {
	return len(vm.vstack)
}

func (vm *VM) resetStack() {
	vm.vstack = nil
}

func (vm *VM) RuntimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")
	line := vm.chunk.lines[vm.ip-1]
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
	vm.resetStack()
}

func runVM(vm *VM) bool {
	for {
		if vm.ip >= len(vm.chunk.bcodes) {
			break
		}
		fmt.Print("          ")
		for _, v := range vm.vstack {
			fmt.Print("[ ")
			fmt.Printf("%s", v.String())
			fmt.Print(" ]")
		}
		fmt.Print("\n")
		DisassembleInstruction(vm.chunk, vm.ip)

		instruction := vm.chunk.bcodes[vm.ip]
		vm.ip++
		switch instruction {
		case OP_CONSTANT:
			pos := vm.chunk.bcodes[vm.ip]
			vm.ip++
			value := vm.chunk.constants[pos]
			vm.pushVstack(value)
		case OP_NIL:
			vm.pushVstack(NewNil())
		case OP_FALSE:
			vm.pushVstack(NewBool(false))
		case OP_TRUE:
			vm.pushVstack(NewBool(true))
		case OP_NOT:
			vm.pushVstack(NewBool(isfalsey(vm.peekVstack(0))))
		case OP_NEGATE:
			if !(vm.peekVstack(0).IsFloat()) {
				vm.RuntimeError("Operand must be number for negate op.")
				return false
			}
			value := vm.popVstack()
			tmp, _ := value.GetFloat()
			vm.pushVstack(NewFloat(-tmp))
		case OP_EQUAL:
			a := vm.popVstack()
			b := vm.popVstack()
			vm.pushVstack(NewBool(ValueEqual(a, b)))
		case OP_GREATER:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				left, _ := vm.popVstack().GetFloat()
				right, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewBool(left > right))
			} else {
				vm.RuntimeError("Operand must be number for > op.")
				return false
			}
		case OP_LESS:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				left, _ := vm.popVstack().GetFloat()
				right, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewBool(left < right))
			} else {
				vm.RuntimeError("Operand must be number for > op.")
				return false
			}
		case OP_ADD:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				left, _ := vm.popVstack().GetFloat()
				right, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewFloat(left + right))
			} else if vm.peekVstack(0).IsString() && vm.peekVstack(1).IsString() {
				left, _ := vm.popVstack().GetString()
				right, _ := vm.popVstack().GetString()
				vm.pushVstack(NewString(left + right))
			} else {
				vm.RuntimeError("Operand must be number or string for add op.")
				return false
			}
		case OP_SUBTRACT:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				left, _ := vm.popVstack().GetFloat()
				right, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewFloat(left - right))
			} else {
				vm.RuntimeError("Operand must be number for sub op.")
				return false
			}
		case OP_MULTIPLY:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				left, _ := vm.popVstack().GetFloat()
				right, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewFloat(left * right))
			} else {
				vm.RuntimeError("Operand must be number for multiply op.")
				return false
			}
		case OP_DIVIDE:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				left, _ := vm.popVstack().GetFloat()
				right, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewFloat(left / right))
			} else {
				vm.RuntimeError("Operand must be number for divide op.")
				return false
			}
		case OP_RETURN:
			return true
		case OP_PRINT:
			fmt.Printf("%s\n", vm.popVstack().String())
		case OP_POP:
			vm.popVstack()
		}
	}
	return true
}

func Interprete(chunk *Chunk) {
	vm := VM{chunk: chunk, ip: 0, vstack: make([]Value, 0)}
	fmt.Printf("-- VM Runtime start\n")
	ok := runVM(&vm)
	fmt.Printf("-- VM Runtime result: %v\n", ok)
}
