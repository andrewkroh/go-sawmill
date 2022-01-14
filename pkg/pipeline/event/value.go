package event

import (
	"encoding/json"
	"fmt"
)

var NullValue = &Value{Type: NullType}

type ValueType uint8

const (
	NullType ValueType = iota
	ArrayType
	BoolType
	BytesType
	FloatType
	IntegerType
	ObjectType
	TimestampType
	UnsignedIntegerType
	maxValueType
)

var valueTypeNames = map[ValueType]string{
	NullType:            "null_type",
	ArrayType:           "array_type",
	BoolType:            "bool_type",
	BytesType:           "bytes_type",
	FloatType:           "float_type",
	IntegerType:         "integer_type",
	ObjectType:          "object_type",
	TimestampType:       "timestamp_type",
	UnsignedIntegerType: "unsigned_integer_type",
}

func (vt ValueType) String() string {
	if name, found := valueTypeNames[vt]; found {
		return name
	}
	return "unknown_type"
}

type Value struct {
	Integer         int64
	UnsignedInteger uint64
	Float           float64
	Timestamp       Time
	Bytes           string
	Array           []*Value
	Object          map[string]*Value
	Bool            bool
	Type            ValueType
}

func (v *Value) value() interface{} {
	switch v.Type {
	case ArrayType:
		return v.Array
	case BoolType:
		return v.Bool
	case BytesType:
		return v.Bytes
	case FloatType:
		return v.Float
	case IntegerType:
		return v.Integer
	case ObjectType:
		return v.Object
	case TimestampType:
		return v.Timestamp
	case UnsignedIntegerType:
		return v.UnsignedInteger
	case NullType:
		return nil
	default:
		return nil
	}
}

func (v Value) String() string {
	return fmt.Sprintf("%s:%v", v.Type.String(), v.value())
}

func (v Value) MarshalJSON() ([]byte, error) {
	// TODO: Replace this with an optimized version.
	return json.Marshal(v.value())
}

func Bool(v bool) *Value {
	return &Value{Type: BoolType, Bool: v}
}

func Integer(v int64) *Value {
	return &Value{Type: IntegerType, Integer: v}
}

func UnsignedInteger(v uint64) *Value {
	return &Value{Type: UnsignedIntegerType, UnsignedInteger: v}
}

func Float(v float64) *Value {
	return &Value{Type: FloatType, Float: v}
}

func String(v string) *Value {
	return &Value{Type: BytesType, Bytes: v}
}

func Timestamp(v Time) *Value {
	return &Value{Type: TimestampType, Timestamp: v}
}

func Array(v ...*Value) *Value {
	return &Value{Type: ArrayType, Array: v}
}

func Object(v map[string]*Value) *Value {
	return &Value{Type: ObjectType, Object: v}
}
