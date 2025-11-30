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
	closure    *LoxClosure
	ip         int
	slots_base int
}

type VM struct {
	frames       [FRAMES_MAX]CallFrame
	frameCount   int
	vstack       [VSTACK_MAX]Value
	vstackCount  int
	globals      map[string]Value
	openUpvalues *UpvalueObj
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
	data := frame.closure.function.chunk.bcodes[frame.ip]
	frame.ip++
	return data
}

func (frame *CallFrame) readShort() uint16 {
	result := uint16(frame.closure.function.chunk.bcodes[frame.ip])<<8 + uint16(frame.closure.function.chunk.bcodes[frame.ip+1])
	frame.ip += 2
	return result
}

func (frame *CallFrame) readConstant() Value {
	pos := frame.readByte()
	return frame.closure.function.chunk.constants[pos]
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
	vm.openUpvalues = nil
}

func (vm *VM) call(closure *LoxClosure, argCount int) bool {
	function := closure.function
	if argCount != function.arity {
		vm.RuntimeError("Expected %d arguments but got %d.", function.arity, argCount)
		return false
	}
	if vm.frameCount == FRAMES_MAX {
		vm.RuntimeError("Stack overflow.")
		return false
	}
	frame := &vm.frames[vm.frameCount]
	vm.frameCount++
	frame.closure = closure
	frame.ip = 0
	frame.slots_base = vm.vstackCount - argCount - 1
	return true
}

func (vm *VM) callValue(callee Value, argCount int) bool {
	if callee.IsClosure() {
		closure, _ := callee.GetClosure()
		return vm.call(closure, argCount)
	} else if callee.IsClass() {
		klass, _ := callee.GetClass()
		instance := NewInstance(klass)
		vm.vstack[vm.vstackCount-argCount-1] = InstanceVal(instance)
		return true
	} else if callee.IsBoundMethod() {
		boundMethod, _ := callee.GetBoundMethod()
		vm.vstack[vm.vstackCount-argCount-1] = boundMethod.receiver
		return vm.call(boundMethod.method, argCount)
	} else if callee.IsNative() {
		native, _ := callee.GetNative()
		result := native(argCount, &vm.vstack[vm.vstackCount-argCount])
		vm.vstackCount -= argCount + 1
		vm.pushVstack(result)
		return true
	}
	vm.RuntimeError("Can only call functions and classes.")
	return false
}

// local is vstack absolutely offset
func (vm *VM) CaptureUpvalue(value *Value, local int) *UpvalueObj {
	var previousUpvalue *UpvalueObj = nil
	var currentUpvalue *UpvalueObj = vm.openUpvalues
	for currentUpvalue != nil && currentUpvalue.location > local {
		previousUpvalue = currentUpvalue
		currentUpvalue = currentUpvalue.next
	}

	if currentUpvalue != nil && currentUpvalue.location == local {
		return currentUpvalue
	}
	createdUpvalue := NewUpvalueObj(value, local)
	createdUpvalue.next = currentUpvalue
	if previousUpvalue == nil {
		vm.openUpvalues = createdUpvalue
	} else {
		previousUpvalue.next = createdUpvalue
	}
	return createdUpvalue
}

func (vm *VM) closeUpvalues(last int) {
	for vm.openUpvalues != nil && vm.openUpvalues.location >= last {
		upvalue := vm.openUpvalues
		upvalue.closed = *upvalue.ref
		upvalue.ref = &upvalue.closed
		vm.openUpvalues = upvalue.next
	}
}

func (vm *VM) bindMethod(klass *LoxClass, name string) bool {
	val, ok := tableGet(klass.methods, name)
	if !ok {
		return false
	}
	method, _ := val.GetClosure()
	boundMethod := NewBoundMethod(vm.peekVstack(0), method)
	vm.popVstack() // pop instance value
	vm.pushVstack(BoundMethodVal(boundMethod))
	return true
}

func (vm *VM) RuntimeError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	fmt.Fprintf(os.Stderr, "\n")

	for i := vm.frameCount - 1; i >= 0; i-- {
		frame := &vm.frames[i]
		function := frame.closure.function
		fmt.Fprintf(os.Stderr, "[line %d] in ", function.chunk.lines[frame.ip])
		if function.name == "" {
			fmt.Fprintf(os.Stderr, "script\n")
		} else {
			fmt.Fprintf(os.Stderr, "%s()\n", function.name)
		}
	}

	vm.resetStack()
}

func (vm *VM) runVM() bool {
	frame := &vm.frames[vm.frameCount-1]

	for {
		if frame.ip >= len(frame.closure.function.chunk.bcodes) {
			break
		}
		DebugVM(vm)

		instruction := frame.readByte()

		switch instruction {
		case OP_CONSTANT:
			vm.pushVstack(frame.readConstant())
		case OP_NIL:
			vm.pushVstack(NilVal())
		case OP_FALSE:
			vm.pushVstack(BoolVal(false))
		case OP_TRUE:
			vm.pushVstack(BoolVal(true))
		case OP_NOT:
			vm.pushVstack(BoolVal(isfalsey(vm.peekVstack(0))))
		case OP_NEGATE:
			if !(vm.peekVstack(0).IsFloat()) {
				vm.RuntimeError("Operand must be number for negate op.")
				return false
			}
			value := vm.popVstack()
			tmp, _ := value.GetFloat()
			vm.pushVstack(FloatVal(-tmp))
		case OP_EQUAL:
			right := vm.popVstack()
			left := vm.popVstack()
			vm.pushVstack(BoolVal(IsValueEqual(&left, &right)))
		case OP_GREATER:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(BoolVal(left > right))
			} else {
				vm.RuntimeError("Operand must be number for > op.")
				return false
			}
		case OP_LESS:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(BoolVal(left < right))
			} else {
				vm.RuntimeError("Operand must be number for > op.")
				return false
			}
		case OP_ADD:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(FloatVal(left + right))
			} else if vm.peekVstack(0).IsString() && vm.peekVstack(1).IsString() {
				right, _ := vm.popVstack().GetString()
				left, _ := vm.popVstack().GetString()
				vm.pushVstack(StringVal(left + right))
			} else {
				vm.RuntimeError("Operand must be number or string for add op.")
				return false
			}
		case OP_SUBTRACT:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(FloatVal(left - right))
			} else {
				vm.RuntimeError("Operand must be number for sub op.")
				return false
			}
		case OP_MULTIPLY:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(FloatVal(left * right))
			} else {
				vm.RuntimeError("Operand must be number for multiply op.")
				return false
			}
		case OP_DIVIDE:
			if vm.peekVstack(0).IsFloat() && vm.peekVstack(1).IsFloat() {
				right, _ := vm.popVstack().GetFloat()
				left, _ := vm.popVstack().GetFloat()
				vm.pushVstack(FloatVal(left / right))
			} else {
				vm.RuntimeError("Operand must be number for divide op.")
				return false
			}
		case OP_RETURN:
			result := vm.popVstack()
			vm.closeUpvalues(frame.slots_base)
			vm.frameCount--
			if vm.frameCount == 0 {
				vm.popVstack()
				return true
			}
			vm.vstackCount = frame.slots_base
			vm.pushVstack(result)
			frame = &vm.frames[vm.frameCount-1]
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
				vm.RuntimeError("Undefined variable '%s' when GET_GLOBAL.", name)
				return false
			}
			vm.pushVstack(value)
		case OP_SET_GLOBAL:
			name, _ := frame.readConstant().GetString()
			isNewKey := tableSet(vm.globals, name, vm.peekVstack(0))
			if isNewKey {
				tableDelete(vm.globals, name)
				vm.RuntimeError("Undefined variable '%s' when SET_GLOBAL.", name)
				return false
			}
		case OP_GET_LOCAL:
			slot := frame.readByte()
			vm.pushVstack(vm.vstack[frame.slots_base+int(slot)])
		case OP_SET_LOCAL:
			slot := frame.readByte()
			vm.vstack[frame.slots_base+int(slot)] = vm.peekVstack(0)
		case OP_GET_UPVALUE:
			slot := frame.readByte()
			vm.pushVstack(*frame.closure.upvalues[slot].ref)
		case OP_SET_UPVALUE:
			slot := frame.readByte()
			*frame.closure.upvalues[slot].ref = vm.peekVstack(0)
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
		case OP_CALL:
			argCount := frame.readByte()
			if !vm.callValue(vm.peekVstack(int(argCount)), int(argCount)) {
				return false
			}
			frame = &vm.frames[vm.frameCount-1]
		case OP_CLOSURE:
			val := frame.readConstant()
			function, ok := val.GetFunction()
			if !ok {
				vm.RuntimeError("Expect LoxFunction obj for OP_CLOSURE.")
				return false
			}
			closure := NewClosure(function)
			vm.pushVstack(ClosureVal(closure))

			for i := 0; i < len(closure.upvalues); i++ {
				isLocal := frame.readByte()
				index := frame.readByte()
				if isLocal != 0 {
					closure.upvalues[i] = vm.CaptureUpvalue(&vm.vstack[frame.slots_base+int(index)], frame.slots_base+int(index))
				} else {
					closure.upvalues[i] = frame.closure.upvalues[index]
				}
			}
		case OP_CLOSE_UPVALUE:
			vm.closeUpvalues(vm.vstackCount - 1)
			vm.popVstack()
		case OP_CLASS:
			name, _ := frame.readConstant().GetString()
			vm.pushVstack(ClassVal(NewClass(name)))
		case OP_GET_PROPERTY:
			if !vm.peekVstack(0).IsInstance() {
				vm.RuntimeError("Only instances have fields when get.")
				return false
			}
			instance, _ := vm.peekVstack(0).GetInstance()
			name, _ := frame.readConstant().GetString()
			val, ok := tableGet(instance.fields, name)
			if ok {
				vm.popVstack()
				vm.pushVstack(val)
				break
			}
			if vm.bindMethod(instance.klass, name) {
				break
			}
			vm.RuntimeError("Undefined property '%s'.", name)
			return false
		case OP_SET_PROPERTY:
			if !vm.peekVstack(1).IsInstance() {
				vm.RuntimeError("Only instances have fields when set.")
				return false
			}
			instance, _ := vm.peekVstack(1).GetInstance()
			fieldName, _ := frame.readConstant().GetString()
			tableSet(instance.fields, fieldName, vm.peekVstack(0))
			value := vm.popVstack()
			vm.popVstack()
			vm.pushVstack(value)
		case OP_METHOD:
			klass, _ := vm.peekVstack(1).GetClass()
			methodName, _ := frame.readConstant().GetString()
			klass.methods[methodName] = vm.peekVstack(0)
			vm.popVstack()
		}
	}
	return true
}

func (vm *VM) DefineNative(name string, function NativeFn) {
	vm.pushVstack(StringVal(name))
	vm.pushVstack(NativeVal(function))
	tmp, _ := vm.peekVstack(1).GetString()
	tableSet(vm.globals, tmp, vm.peekVstack(0))
	vm.popVstack()
	vm.popVstack()
}

func Interprete(function *LoxFunction) {
	fmt.Printf("-- GLOX VM --\n")
	vm := VM{}
	vm.resetStack()
	vm.DefineNative("clock", ClockNative)

	clousre := NewClosure(function)
	vm.pushVstack(ClosureVal(clousre))
	vm.call(clousre, 0)

	ok := vm.runVM()
	if !ok {
		fmt.Printf("GLOX VM runtime error\n")
	}
}
