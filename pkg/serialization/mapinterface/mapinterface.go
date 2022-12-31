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

package mapinterface

import (
	"fmt"

	"github.com/andrewkroh/go-sawmill/pkg/event"
	"github.com/andrewkroh/go-sawmill/pkg/eventutil"
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
	fields, err := eventutil.ReflectValue(m)
	if err != nil {
		return nil, err
	}

	evt := event.New()
	for k, v := range fields.Object {
		evt.Put(eventutil.EscapeKey(k), v)
	}

	return evt, nil
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
