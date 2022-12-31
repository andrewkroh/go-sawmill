// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package eventutil

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/andrewkroh/go-sawmill/pkg/event"
)

const tagName = "event"

var (
	errorInterface = reflect.TypeOf((*error)(nil)).Elem()
	timeInterface  = reflect.TypeOf((*time.Time)(nil)).Elem()
)

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "eventutil: unsupported type: " + e.Type.String()
}

func ReflectValue(v interface{}) (*event.Value, error) {
	return interfaceToValue(v)
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

			obj[EscapeKey(k)] = v
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
		sliceValue, err := interfaceToValue(rv.Index(i).Interface())
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

		tagName, tagOpts := parseTag(field.Tag.Get(tagName))
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
		if tag := field.Tag.Get(tagName); tag == "-" {
			continue
		}

		f = append(f, field)
	}

	return f
}
