package mapinterface

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/util"
)

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "mapinterface: unsupported type: " + e.Type.String()
}

var TagName = "event"

var (
	errorInterface = reflect.TypeOf((*error)(nil)).Elem()
	timeInterface  = reflect.TypeOf((*time.Time)(nil)).Elem()
)

// ToEvent converts a map[string]interface{} to an Event.
//
// Pointer values encode as the value pointed to.
// A nil pointer encodes as the null JSON value.
//
// Interface values encode as the value contained in the interface.
// A nil interface value encodes as the null JSON value.
//
// Channel, complex, and function values cannot be represented in an Event.
// Attempting to encode such a value causes ToEvent to return
// an UnsupportedTypeError.
//
// The data should not contain any cycles. Cycles are not detected and will
// result in a panic caused by stack overflow.
func ToEvent(m map[string]interface{}) (*event.Event, error) {
	fields, err := interfaceToValue(m)
	if err != nil {
		return nil, err
	}

	evt := event.New()
	for k, v := range fields.Object {
		evt.Put(util.EscapeKey(k), v)
	}

	return evt, nil
}

func interfaceToValue(ifc interface{}) (*event.Value, error) {
	switch t := ifc.(type) {
	case bool:
		return event.Bool(t), nil
	case float32:
		return event.Float(float64(t)), nil
	case float64:
		return event.Float(t), nil
	case int:
		return event.Integer(int64(t)), nil
	case int8:
		return event.Integer(int64(t)), nil
	case int16:
		return event.Integer(int64(t)), nil
	case int32:
		return event.Integer(int64(t)), nil
	case int64:
		return event.Integer(t), nil
	case uint:
		return event.UnsignedInteger(uint64(t)), nil
	case uint8:
		return event.UnsignedInteger(uint64(t)), nil
	case uint16:
		return event.UnsignedInteger(uint64(t)), nil
	case uint32:
		return event.UnsignedInteger(uint64(t)), nil
	case uint64:
		return event.UnsignedInteger(t), nil
	case uintptr:
		return event.UnsignedInteger(uint64(t)), nil
	case string:
		return event.String(t), nil
	case time.Time:
		return event.Timestamp(t.UnixNano()), nil
	case map[string]interface{}:
		obj := make(map[string]*event.Value, len(t))
		for k, v := range t {
			v, err := interfaceToValue(v)
			if err != nil {
				return nil, fmt.Errorf("failed on key %q: %w", k, err)
			}

			obj[util.EscapeKey(k)] = v
		}
		return event.Object(obj), nil
	case nil:
		return event.NullValue, nil
	default:
		return reflectToValue(reflect.ValueOf(t))
	}
}

func reflectToValue(rv reflect.Value) (*event.Value, error) {
	typ := rv.Type()

	switch {
	case typ.Implements(errorInterface):
		err := rv.Interface().(error)
		return event.String(err.Error()), nil
	}

	switch typ.Kind() {
	case reflect.Ptr:
		if rv.IsNil() {
			return event.NullValue, nil
		}
		return reflectToValue(rv.Elem())
	case reflect.Bool:
		return event.Bool(rv.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return event.Integer(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return event.UnsignedInteger(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return event.Float(rv.Float()), nil
	case reflect.String:
		return event.String(rv.String()), nil
	case reflect.Array, reflect.Slice:
		return reflectSliceToArray(rv)
	case reflect.Map:
		return reflectMapToObject(rv)
	case reflect.Struct:
		switch {
		case rv.CanConvert(timeInterface):
			return interfaceToValue(rv.Convert(timeInterface).Interface())
		}
		return reflectStructToObject(rv)
	default:
		return nil, &UnsupportedTypeError{typ}
	}
}

func reflectSliceToArray(rv reflect.Value) (*event.Value, error) {
	n := rv.Len()
	values := make([]*event.Value, 0, n)
	for i := 0; i < n; i++ {
		sliceValue, err := reflectToValue(rv.Index(i))
		if err != nil {
			return nil, err
		}

		values = append(values, sliceValue)
	}

	return event.Array(values...), nil
}

func reflectMapToObject(rv reflect.Value) (*event.Value, error) {
	obj := make(map[string]*event.Value, rv.Len())

	m := rv.MapRange()
	for m.Next() {
		key := fmt.Sprintf("%v", m.Key())
		val, err := interfaceToValue(m.Value().Interface())
		if err != nil {
			return nil, err
		}
		obj[key] = val
	}

	return event.Object(obj), nil
}

func reflectStructToObject(rv reflect.Value) (*event.Value, error) {
	fields := structFields(rv)
	obj := make(map[string]*event.Value, len(fields))

	for _, field := range fields {
		name := strings.ToLower(field.Name)

		tagName, tagOpts := parseTag(field.Tag.Get(TagName))
		if tagName != "" {
			name = tagName
		}

		val := rv.FieldByName(field.Name)

		// if the value is a zero value and the field is marked as omitempty do
		// not include
		if tagOpts.Has("omitempty") {
			zero := reflect.Zero(val.Type()).Interface()
			current := val.Interface()

			if reflect.DeepEqual(current, zero) {
				continue
			}
		}

		finalVal, err := reflectToValue(val)
		if err != nil {
			return nil, err
		}

		if tagOpts.Has("string") {
			s, ok := val.Interface().(fmt.Stringer)
			if ok {
				obj[name] = event.String(s.String())
			}
			continue
		}

		obj[name] = finalVal
	}

	return event.Object(obj), nil
}

func structFields(rv reflect.Value) []reflect.StructField {
	t := rv.Type()

	var f []reflect.StructField

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// we can't access the value of unexported fields
		if field.PkgPath != "" {
			continue
		}

		// don't check if it's omitted
		if tag := field.Tag.Get(TagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}

	return f
}

func FromEvent(evt *event.Event) map[string]interface{} {
	return fromEventValueObject(evt.Get("."))
}

func fromEventValueObject(v *event.Value) map[string]interface{} {
	if v == nil {
		return nil
	}

	if v.Type != event.ObjectType {
		// Developer error.
		panic("fromEventValueObject can only be used with type=ObjectType")
	}

	fields := make(map[string]interface{}, len(v.Object))
	for k, v := range v.Object {
		fields[k] = fromEventValue(v)
	}
	return fields
}

func fromEventValue(v *event.Value) interface{} {
	if v == nil {
		return nil
	}

	switch v.Type {
	case event.ArrayType:
		items := make([]interface{}, 0, len(v.Array))
		for _, x := range v.Array {
			items = append(items, fromEventValue(x))
		}
		return items
	case event.BoolType:
		return v.Bool
	case event.FloatType:
		return v.Float
	case event.IntegerType:
		return v.Integer
	case event.ObjectType:
		return fromEventValueObject(v)
	case event.StringType:
		return v.String
	case event.TimestampType:
		return v.Timestamp.GoTime()
	case event.UnsignedIntegerType:
		return v.UnsignedInteger
	case event.NullType:
		return nil
	default:
		panic(fmt.Errorf("unhandled value type <%v>", v.Type))
	}
}
