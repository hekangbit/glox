package main

import (
	"fmt"
)

type VariantType int

const (
	TypeInt VariantType = iota
	TypeNil
	TypeFloat
	TypeString
	TypeBool
)

type Variant struct {
	typ  VariantType
	iVal int
	fVal float64
	sVal string
	bVal bool
}

type Value Variant

func ValueEqual(v1, v2 Value) bool {
	if v1.typ != v2.typ {
		return false
	}
	switch v1.typ {
	case TypeNil:
		return true
	case TypeBool:
		return isfalsey(v1) == isfalsey(v2)
	case TypeInt:
		a, _ := v1.GetInt()
		b, _ := v2.GetInt()
		return a == b
	case TypeFloat:
		a, _ := v1.GetFloat()
		b, _ := v2.GetFloat()
		return a == b
	case TypeString:
		a, _ := v1.GetString()
		b, _ := v2.GetString()
		return a == b
	}
	return false
}

func NewNil() Value {
	return Value{typ: TypeNil, iVal: 0}
}

func NewInt(v int) Value {
	return Value{typ: TypeInt, iVal: v}
}

func NewFloat(v float64) Value {
	return Value{typ: TypeFloat, fVal: v}
}

func NewString(v string) Value {
	return Value{typ: TypeString, sVal: v}
}

func NewBool(v bool) Value {
	return Value{typ: TypeBool, bVal: v}
}

func (v Value) Type() VariantType {
	return v.typ
}

func (v Value) IsNil() bool {
	return v.Type() == TypeNil
}

func (v Value) IsInt() bool {
	return v.Type() == TypeInt
}

func (v Value) IsFloat() bool {
	return v.Type() == TypeFloat
}

func (v Value) IsString() bool {
	return v.Type() == TypeString
}

func (v Value) IsBool() bool {
	return v.Type() == TypeBool
}

func (v Value) GetInt() (int, bool) {
	if v.typ == TypeInt {
		return v.iVal, true
	}
	return 0, false
}

func (v Value) GetFloat() (float64, bool) {
	if v.typ == TypeFloat {
		return v.fVal, true
	}
	return 0, false
}

func (v Value) GetString() (string, bool) {
	if v.typ == TypeString {
		return v.sVal, true
	}
	return "", false
}

func (v Value) GetBool() (bool, bool) {
	if v.typ == TypeBool {
		return v.bVal, true
	}
	return false, false
}

func (v *Value) SetNil() {
	v.typ = TypeNil
	v.iVal = 0
}

func (v *Value) SetInt(i int) {
	v.typ = TypeInt
	v.iVal = i
}

func (v *Value) SetFloat(f float64) {
	v.typ = TypeFloat
	v.fVal = f
}

func (v *Value) SetString(s string) {
	v.typ = TypeString
	v.sVal = s
}

func (v *Value) SetBool(b bool) {
	v.typ = TypeBool
	v.bVal = b
}

func (v Value) String() string {
	switch v.typ {
	case TypeNil:
		return "nil"
	case TypeInt:
		return fmt.Sprintf("%d", v.iVal)
	case TypeFloat:
		return fmt.Sprintf("%f", v.fVal)
	case TypeString:
		return v.sVal
	case TypeBool:
		return fmt.Sprintf("%t", v.bVal)
	default:
		return "unknown"
	}
}
