package event

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueTypeString(t *testing.T) {
	for i := 0; i < int(maxValueType); i++ {
		assert.NotEqual(t, "unknown_type", ValueType(i).String())
	}

	assert.Equal(t, "unknown_type", ValueType(maxValueType).String())
}

func TestValueString(t *testing.T) {
	var testCases = []struct {
		Value  Value
		String string
	}{
		{
			Value{
				Type: NullType,
			},
			"null_type:<nil>",
		},
		{
			Value{
				Type:    IntegerType,
				Integer: 1,
			},
			"integer_type:1",
		},
		{
			Value{
				Type:            UnsignedIntegerType,
				UnsignedInteger: 1,
			},
			"unsigned_integer_type:1",
		},
		{
			Value{
				Type:  FloatType,
				Float: 1.0,
			},
			"float_type:1",
		},
		{
			Value{
				Type:  BytesType,
				Bytes: "foo",
			},
			"bytes_type:foo",
		},
		{
			Value{
				Type: BoolType,
				Bool: true,
			},
			"bool_type:true",
		},
		{
			Value{
				Type:      TimestampType,
				Timestamp: Time{testTimeUnix},
			},
			"timestamp_type:" + strconv.FormatInt(testTimeUnix, 10),
		},
		{
			Value{
				Type:  ArrayType,
				Array: Array(NullValue).Array,
			},
			"array_type:[null_type:<nil>]",
		},
		{
			Value{
				Type: ObjectType,
				Object: map[string]*Value{
					"foo": String("bar"),
				},
			},
			"object_type:{foo:bytes_type:bar}",
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.String, tc.Value.String())
	}
}

func TestValueJSONMarshal(t *testing.T) {
	v := Value{
		Type: ObjectType,
		Object: map[string]*Value{
			"hello": String("world"),
		},
	}

	data, err := json.Marshal(v)
	require.NoError(t, err)
	t.Log(string(data))

	assert.Equal(t, `{"hello":"world"}`, string(data))
}
