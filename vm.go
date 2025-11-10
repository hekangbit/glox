package main

import (
	"fmt"
	"math"
	"os"
)

const (
	FRAMES_MAX int = iota + 64
	VSTACK_MAX int = FRAMES_MAX * math.MaxUint8
)

type CallFrame struct {
	function   *LoxFunction
	ip         int
	slots_base int
}

type VM struct {
	frames      [FRAMES_MAX]CallFrame
	frameCount  int
	vstack      [VSTACK_MAX]Value
	vstackCount int
	globals     map[string]Value
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

func (frame *CallFrame) readByte() byte {
	data := frame.function.chunk.bcodes[frame.ip]
	frame.ip++
	return data
}

func (frame *CallFrame) readShort() uint16 {
	result := uint16(frame.function.chunk.bcodes[frame.ip])<<8 + uint16(frame.function.chunk.bcodes[frame.ip+1])
	frame.ip += 2
	return result
}

func (frame *CallFrame) readConstant() Value {
	pos := frame.readByte()
	return frame.function.chunk.constants[pos]
}

func (vm *VM) pushVstack(value Value) {
	vm.vstack[vm.vstackCount] = value
	vm.vstackCount++
}

func (vm *VM) popVstack() Value {
	value := vm.vstack[vm.vstackCount-1]
	vm.vstackCount--
	return value
}

func (vm *VM) peekVstack(offset int) Value {
	return vm.vstack[vm.vstackCount-1-offset]
}

func (vm *VM) resetStack() {
	vm.vstackCount = 0
	vm.vstack = [VSTACK_MAX]Value{}
	vm.frameCount = 0
	vm.frames = [FRAMES_MAX]CallFrame{}
	vm.globals = make(map[string]Value)
}

func (vm *VM) RuntimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")
	frame := vm.frames[vm.frameCount-1]
	line := frame.function.chunk.lines[frame.ip]
	fmt.Fprintf(os.Stderr, "[line %d] in script\n", line)
	vm.resetStack()
}

func (vm *VM) runVM() bool {
	frame := &vm.frames[vm.frameCount-1]

	for {
		if frame.ip >= len(frame.function.chunk.bcodes) {
			break
		}
		DebugVM(vm)

		instruction := frame.readByte()

		switch instruction {
		case OP_CONSTANT:
			vm.pushVstack(frame.readConstant())
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
			vm.pushVstack(NewBool(IsValueEqual(&left, &right)))
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
			name, _ := frame.readConstant().GetString()
			tableSet(vm.globals, name, vm.peekVstack(0))
			vm.popVstack()
		case OP_GET_GLOBAL:
			name, _ := frame.readConstant().GetString()
			value, ok := tableGet(vm.globals, name)
			if !ok {
				vm.RuntimeError("Undefined variable '%s'.", name)
				return false
			}
			vm.pushVstack(value)
		case OP_SET_GLOBAL:
			name, _ := frame.readConstant().GetString()
			isNewKey := tableSet(vm.globals, name, vm.peekVstack(0))
			if isNewKey {
				tableDelete(vm.globals, name)
				vm.RuntimeError("Undefined variable '%s'.", name)
				return false
			}
		case OP_GET_LOCAL:
			slot := frame.readByte()
			vm.pushVstack(vm.vstack[frame.slots_base+int(slot)])
		case OP_SET_LOCAL:
			slot := frame.readByte()
			vm.vstack[frame.slots_base+int(slot)] = vm.peekVstack(0)
		case OP_JUMP:
			offset := frame.readShort()
			frame.ip += int(offset)
		case OP_JUMP_IF_FALSE:
			offset := frame.readShort()
			if isfalsey(vm.peekVstack(0)) {
				frame.ip += int(offset)
			}
		case OP_LOOP:
			offset := frame.readShort()
			frame.ip -= int(offset)
		}
	}
	return true
}

func Interprete(function *LoxFunction) {
	fmt.Printf("-- GLOX VM --\n")
	vm := VM{}
	vm.resetStack()

	vm.pushVstack(FunctionValue(function))

	frame := &vm.frames[vm.frameCount]
	frame.function = function
	frame.ip = 0
	frame.slots_base = 0
	vm.frameCount++

	ok := vm.runVM()
	if !ok {
		fmt.Printf("GLOX VM runtime error\n")
	}
}
