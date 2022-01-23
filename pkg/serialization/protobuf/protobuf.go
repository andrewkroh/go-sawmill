package protobuf

import (
	"fmt"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/util"
)

// FromEvent returns a protocol buffer message containing the log
// event.
func FromEvent(evt *event.Event) *MessageWrapper {
	return &MessageWrapper{
		Message: &MessageWrapper_Log{
			Log: &Log{
				Object: fromEventValueObject(evt.Get(".")),
			},
		},
	}
}

func fromEventValue(v *event.Value) *Value {
	if v == nil {
		return nil
	}

	switch v.Type {
	case event.ArrayType:
		items := make([]*Value, 0, len(v.Array))
		for _, x := range v.Array {
			items = append(items, fromEventValue(x))
		}
		return &Value{
			Kind: &Value_Array{
				Array: &ValueArray{
					Items: items,
				},
			},
		}
	case event.BoolType:
		return &Value{
			Kind: &Value_Boolean{
				Boolean: v.Bool,
			},
		}
	case event.FloatType:
		return &Value{
			Kind: &Value_Float{
				Float: v.Float,
			},
		}
	case event.IntegerType:
		return &Value{
			Kind: &Value_Integer{
				Integer: v.Integer,
			},
		}
	case event.ObjectType:
		return &Value{
			Kind: &Value_Object{
				Object: fromEventValueObject(v),
			},
		}
	case event.StringType:
		return &Value{
			Kind: &Value_String_{
				String_: []byte(v.String),
			},
		}
	case event.TimestampType:
		return &Value{
			Kind: &Value_Timestamp{
				Timestamp: v.Timestamp.UnixNanos,
			},
		}
	case event.UnsignedIntegerType:
		return &Value{
			Kind: &Value_UnsignedInteger{
				UnsignedInteger: v.UnsignedInteger,
			},
		}
	case event.NullType:
		return &Value{
			Kind: &Value_Null{
				Null: ValueNull_NULL_VALUE,
			},
		}
	default:
		panic(fmt.Errorf("unhandled value type <%v>", v.Type))
	}
}

func fromEventValueObject(v *event.Value) *ValueObject {
	if v == nil {
		return nil
	}

	if v.Type != event.ObjectType {
		// Developer error.
		panic("fromEventValueObject can only be used with type=ObjectType")
	}

	fields := make(map[string]*Value, len(v.Object))
	for k, v := range v.Object {
		fields[k] = fromEventValue(v)
	}
	return &ValueObject{
		Fields: fields,
	}
}

// ToLogEvent returns a log event from a protocol buffer message.
func ToLogEvent(l *Log) *event.Event {
	if l == nil || l.Object == nil {
		return nil
	}

	evt := event.New()
	for k, v := range l.Object.Fields {
		evt.Put(util.EscapeKey(k), toEventValue(v))
	}
	return evt
}

func toEventValue(v *Value) *event.Value {
	switch t := v.GetKind().(type) {
	case *Value_Null:
		return event.NullValue
	case *Value_Array:
		items := make([]*event.Value, 0, len(t.Array.Items))
		for _, x := range t.Array.Items {
			items = append(items, toEventValue(x))
		}
		return event.Array(items...)
	case *Value_Boolean:
		return event.Bool(t.Boolean)
	case *Value_Float:
		return event.Float(t.Float)
	case *Value_Integer:
		return event.Integer(t.Integer)
	case *Value_Object:
		fields := make(map[string]*event.Value, len(t.Object.Fields))
		for k, v := range t.Object.Fields {
			fields[util.EscapeKey(k)] = toEventValue(v)
		}
		return event.Object(fields)
	case *Value_String_:
		return event.String(string(t.String_))
	case *Value_Timestamp:
		return event.Timestamp(t.Timestamp)
	case *Value_UnsignedInteger:
		return event.UnsignedInteger(t.UnsignedInteger)
	default:
		panic(fmt.Errorf("unhandled value type <%T>", v.GetKind()))
	}
}
