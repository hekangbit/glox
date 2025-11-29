package main

import (
	"fmt"
	"reflect"
)

type Variant struct {
	value interface{}
}

type Value Variant

type LoxFunction struct {
	arity        int
	chunk        Chunk
	name         string
	upValueCount int
}

type LoxClosure struct {
	function *LoxFunction
	upvalues []*UpvalueObj
}

type UpvalueObj struct {
	ref      *Value
	location int
	closed   Value
	next     *UpvalueObj
}

type LoxClass struct {
	name    string
	methods map[string]Value
}

type LoxInstance struct {
	klass  *LoxClass
	fields map[string]Value
}

type NativeFn func(int, *Value) Value

func NewFunction() *LoxFunction {
	return &LoxFunction{arity: 0, name: "", chunk: Chunk{}, upValueCount: 0}
}

func NewClosure(function *LoxFunction) *LoxClosure {
	closure := LoxClosure{function: function, upvalues: make([]*UpvalueObj, 0)}
	for i := 0; i < function.upValueCount; i++ {
		closure.upvalues = append(closure.upvalues, nil)
	}
	return &closure
}

func NewUpvalueObj(ref *Value, slot int) *UpvalueObj {
	upvalue := UpvalueObj{ref: ref, location: slot, next: nil, closed: Value{}}
	return &upvalue
}

func NewClass(name string) *LoxClass {
	return &LoxClass{name: name, methods: make(map[string]Value)}
}

func NewInstance(klass *LoxClass) *LoxInstance {
	instance := LoxInstance{klass: klass, fields: make(map[string]Value)}
	return &instance
}

func isSameType(v1 *Value, v2 *Value) bool {
	if v1.value == nil && v2.value == nil {
		return true
	}
	if v1.value == nil || v2.value == nil {
		return false
	}
	return reflect.TypeOf(v1.value) == reflect.TypeOf(v2.value)
}

func (v *Value) Set(value interface{}) {
	v.value = value
}

func (v *Value) Get() interface{} {
	return v.value
}

func NilVal() Value {
	return Value{value: nil}
}

func IntVal(v int) Value {
	return Value{value: v}
}

func FloatVal(v float64) Value {
	return Value{value: v}
}

func StringVal(v string) Value {
	return Value{value: v}
}

func BoolVal(v bool) Value {
	return Value{value: v}
}

func FunctionVal(function *LoxFunction) Value {
	return Value{value: function}
}

func NativeVal(function NativeFn) Value {
	return Value{value: function}
}

func ClosureVal(closure *LoxClosure) Value {
	return Value{value: closure}
}

func ClassVal(class *LoxClass) Value {
	return Value{value: class}
}

func InstanceVal(instance *LoxInstance) Value {
	return Value{value: instance}
}

func (v Value) IsNil() bool {
	return v.value == nil
}

func (v Value) IsInt() bool {
	_, ok := v.value.(int)
	return ok
}

func (v Value) IsFloat() bool {
	_, ok := v.value.(float64)
	return ok
}

func (v Value) IsString() bool {
	_, ok := v.value.(string)
	return ok
}

func (v Value) IsBool() bool {
	_, ok := v.value.(bool)
	return ok
}

func (v Value) IsFunction() bool {
	_, ok := v.value.(*LoxFunction)
	return ok
}

func (v Value) IsNative() bool {
	_, ok := v.value.(NativeFn)
	return ok
}

func (v Value) IsClosure() bool {
	_, ok := v.value.(*LoxClosure)
	return ok
}

func (v Value) IsClass() bool {
	_, ok := v.value.(*LoxClass)
	return ok
}

func (v Value) IsInstance() bool {
	_, ok := v.value.(*LoxInstance)
	return ok
}

func (v Value) GetInt() (int, bool) {
	result, ok := v.value.(int)
	if ok {
		return result, true
	}
	return 0, false
}

func (v Value) GetFloat() (float64, bool) {
	result, ok := v.value.(float64)
	if ok {
		return result, true
	}
	return 0.0, false
}

func (v Value) GetString() (string, bool) {
	result, ok := v.value.(string)
	if ok {
		return result, true
	}
	return "", false
}

func (v Value) GetBool() (bool, bool) {
	result, ok := v.value.(bool)
	if ok {
		return result, true
	}
	return false, false
}

func (v Value) GetFunction() (*LoxFunction, bool) {
	result, ok := v.value.(*LoxFunction)
	if ok {
		return result, true
	}
	return nil, false
}

func (v Value) GetNative() (NativeFn, bool) {
	result, ok := v.value.(NativeFn)
	if ok {
		return result, true
	}
	return nil, false
}

func (v Value) GetClosure() (*LoxClosure, bool) {
	result, ok := v.value.(*LoxClosure)
	if ok {
		return result, true
	}
	return nil, false
}

func (v Value) GetClass() (*LoxClass, bool) {
	result, ok := v.value.(*LoxClass)
	if ok {
		return result, true
	}
	return nil, false
}

func (v Value) GetInstance() (*LoxInstance, bool) {
	result, ok := v.value.(*LoxInstance)
	if ok {
		return result, true
	}
	return nil, false
}

func (v *Value) SetNil() {
	v.value = nil
}

func (v *Value) SetInt(value int) {
	v.value = value
}

func (v *Value) SetFloat(value float64) {
	v.value = value
}

func (v *Value) SetString(value string) {
	v.value = value
}

func (v *Value) SetBool(value bool) {
	v.value = value
}

func IsValueEqual(v1, v2 *Value) bool {
	if !isSameType(v1, v2) {
		return false
	}

	if v1.value == nil {
		return true
	}

	switch v1.value.(type) {
	case bool:
		a, _ := v1.GetBool()
		b, _ := v2.GetBool()
		return a == b
	case int:
		a, _ := v1.GetInt()
		b, _ := v2.GetInt()
		return a == b
	case float64:
		a, _ := v1.GetFloat()
		b, _ := v2.GetFloat()
		return a == b
	case string:
		a, _ := v1.GetString()
		b, _ := v2.GetString()
		return a == b
	}

	return false
}

func (v Value) String() string {
	switch v.value.(type) {
	case nil:
		return "nil"
	case bool:
		result, _ := v.GetBool()
		return fmt.Sprintf("%t", result)
	case int:
		result, _ := v.GetInt()
		return fmt.Sprintf("%d", result)
	case float64:
		result, _ := v.GetFloat()
		return fmt.Sprintf("%f", result)
	case string:
		result, _ := v.GetString()
		return result
	case *LoxFunction:
		function, _ := v.value.(*LoxFunction)
		return NormalizedFuncName(function.name)
	case *LoxClosure:
		closure, _ := v.value.(*LoxClosure)
		return NormalizedClosureName(closure)
	case NativeFn:
		return "<native fn>"
	case *LoxClass:
		klass, _ := v.value.(*LoxClass)
		return klass.name
	case *LoxInstance:
		instance, _ := v.value.(*LoxInstance)
		return instance.klass.name + " instance"
	default:
		return "unknown"
	}
}

func NormalizedFuncName(name string) string {
	if name == "" {
		return "<script>"
	}
	return "<fn " + name + ">"
}

func NormalizedClosureName(closure *LoxClosure) string {
	addr := fmt.Sprintf("%p", closure)
	if closure.function.name == "" {
		return "<" + addr + " script>"
	}
	return "<" + addr + " " + closure.function.name + ">"
}
