package event

import (
	"bytes"
	"encoding/json"
	"fmt"
)

var NullValue = &Value{Type: NullType}

type ValueType uint8

const (
	NullType ValueType = iota
	ArrayType
	BoolType
	FloatType
	IntegerType
	ObjectType
	StringType
	TimestampType
	UnsignedIntegerType
	maxValueType
)

var valueTypeNames = map[ValueType]string{
	NullType:            "null_type",
	ArrayType:           "array_type",
	BoolType:            "bool_type",
	FloatType:           "float_type",
	IntegerType:         "integer_type",
	ObjectType:          "object_type",
	StringType:          "string_type",
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
	String          string
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
	case FloatType:
		return v.Float
	case IntegerType:
		return v.Integer
	case ObjectType:
		return v.Object
	case StringType:
		return v.String
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

func (v *Value) Format(f fmt.State, verb rune) {
	// Not implementing String() because it collides with the String field.
	f.Write([]byte(v.Type.String()))
	f.Write([]byte(":"))
	fmt.Fprintf(f, "%v", v.value())
}

func (v *Value) MarshalJSON() ([]byte, error) {
	// TODO: Replace this with an optimized version. There are a finite
	// set of types so we can optimize this to avoid reflection.
	return json.Marshal(v.value())
}

func (v *Value) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	switch tok {
	case json.Delim('['):
		if err = json.Unmarshal(data, &v.Array); err != nil {
			return err
		}
		v.Type = ArrayType
	case json.Delim('{'):
		if err = json.Unmarshal(data, &v.Object); err != nil {
			return err
		}
		v.Type = ObjectType
	default:
		var value interface{}
		if err = json.Unmarshal(data, &value); err != nil {
			return err
		}

		switch x := value.(type) {
		case string:
			v.Type = StringType
			v.String = x
		case float64:
			v.Type = FloatType
			v.Float = x
		case int64:
			v.Type = IntegerType
			v.Integer = x
		case bool:
			v.Type = BoolType
			v.Bool = x
		case nil:
			v.Type = NullType
		default:
			return fmt.Errorf("unhandled type %T for value %q", x, string(data))
		}
	}

	return nil
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
	return &Value{Type: StringType, String: v}
}

func Timestamp(unixNanos int64) *Value {
	return &Value{Type: TimestampType, Timestamp: Time{unixNanos}}
}

func Array(v ...*Value) *Value {
	return &Value{Type: ArrayType, Array: v}
}

func Object(v map[string]*Value) *Value {
	return &Value{Type: ObjectType, Object: v}
}
