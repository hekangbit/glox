package main

import (
	"fmt"
	"os"
)

type VM struct {
	chunk   *Chunk
	ip      int
	vstack  []Value
	globals map[string]Value
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

func tableSet(table map[string]Value, name string, value Value) bool {
	_, ok := table[name]
	table[name] = value
	return !ok
}

func tableGet(table map[string]Value, name string) (Value, bool) {
	value, ok := table[name]
	return value, ok
}

func tableDelete(table map[string]Value, name string) {
	delete(table, name)
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

func (vm *VM) getVStack(slot byte) Value {
	return vm.vstack[slot]
}

func (vm *VM) setVStack(slot byte, value Value) {
	vm.vstack[slot] = value
}

func (vm *VM) sizeVstack() int {
	return len(vm.vstack)
}

func (vm *VM) resetStack() {
	vm.vstack = nil
}

func (vm *VM) readShort() uint16 {
	result := uint16(vm.chunk.bcodes[vm.ip]<<8 + vm.chunk.bcodes[vm.ip+1])
	vm.ip += 2
	return result
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
		DebugVM(vm)

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
			right := vm.popVstack()
			left := vm.popVstack()
			vm.pushVstack(NewBool(ValueEqual(left, right)))
		case OP_GREATER:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewBool(left > right))
			} else {
				vm.RuntimeError("Operand must be number for > op.")
				return false
			}
		case OP_LESS:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewBool(left < right))
			} else {
				vm.RuntimeError("Operand must be number for > op.")
				return false
			}
		case OP_ADD:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewFloat(left + right))
			} else if vm.peekVstack(0).IsString() && vm.peekVstack(1).IsString() {
				right, _ := vm.popVstack().GetString()
				left, _ := vm.popVstack().GetString()
				vm.pushVstack(NewString(left + right))
			} else {
				vm.RuntimeError("Operand must be number or string for add op.")
				return false
			}
		case OP_SUBTRACT:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewFloat(left - right))
			} else {
				vm.RuntimeError("Operand must be number for sub op.")
				return false
			}
		case OP_MULTIPLY:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(NewFloat(left * right))
			} else {
				vm.RuntimeError("Operand must be number for multiply op.")
				return false
			}
		case OP_DIVIDE:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
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
		case OP_DEFINE_GLOBAL:
			pos := vm.chunk.bcodes[vm.ip]
			vm.ip++
			name, _ := vm.chunk.constants[pos].GetString()
			tableSet(vm.globals, name, vm.peekVstack(0))
			vm.popVstack()
		case OP_GET_GLOBAL:
			pos := vm.chunk.bcodes[vm.ip]
			vm.ip++
			name, _ := vm.chunk.constants[pos].GetString()
			value, ok := tableGet(vm.globals, name)
			if !ok {
				vm.RuntimeError("Undefined variable '%s'.", name)
				return false
			}
			vm.pushVstack(value)
		case OP_SET_GLOBAL:
			pos := vm.chunk.bcodes[vm.ip]
			vm.ip++
			name, _ := vm.chunk.constants[pos].GetString()
			isNewKey := tableSet(vm.globals, name, vm.peekVstack(0))
			if isNewKey {
				tableDelete(vm.globals, name)
				vm.RuntimeError("Undefined variable '%s'.", name)
				return false
			}
		case OP_GET_LOCAL:
			slot := vm.chunk.bcodes[vm.ip]
			vm.ip++
			vm.pushVstack(vm.getVStack(slot))
		case OP_SET_LOCAL:
			slot := vm.chunk.bcodes[vm.ip]
			vm.ip++
			vm.setVStack(slot, vm.peekVstack(0))
		case OP_JUMP:
			offset := vm.readShort()
			vm.ip += int(offset)
		case OP_JUMP_IF_FALSE:
			offset := vm.readShort()
			if isfalsey(vm.peekVstack(0)) {
				vm.ip += int(offset)
			}
		}
	}
	return true
}

func Interprete(chunk *Chunk) {
	vm := VM{chunk: chunk, ip: 0, vstack: make([]Value, 0), globals: make(map[string]Value)}
	fmt.Printf("-- GLOX VM --\n")
	ok := runVM(&vm)
	if !ok {
		fmt.Printf("GLOX VM runtime error\n")
	}
}
