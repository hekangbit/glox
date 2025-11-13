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
	arity int
	chunk Chunk
	name  string
}

type NativeFn func(int, *Value) Value

func AllocFunction() *LoxFunction {
	return &LoxFunction{arity: 0, name: "", chunk: Chunk{}}
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

func NewNil() Value {
	return Value{value: nil}
}

func NewInt(v int) Value {
	return Value{value: v}
}

func NewFloat(v float64) Value {
	return Value{value: v}
}

func NewString(v string) Value {
	return Value{value: v}
}

func NewBool(v bool) Value {
	return Value{value: v}
}

func NewFunction(function *LoxFunction) Value {
	return Value{value: function}
}

func NewNative(function NativeFn) Value {
	return Value{value: function}
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
	case NativeFn:
		return "<native fn>"
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
