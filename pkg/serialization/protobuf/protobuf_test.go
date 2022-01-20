package protobuf

import (
	"encoding/json"
	"github.com/gogo/protobuf/types"
	"github.com/golang/protobuf/proto"
	"log"
	"testing"
	"time"
)

var m = map[string]interface{} {
	"hello": "world",
	"event": map[string]interface{}{
		"created": time.Now().UTC().Format(time.RFC3339Nano),
		"ingested": time.Now().UTC().Format(time.RFC3339Nano),
		"count": 3,
	},
}

var logEvent = &Log{
	Fields: map[string]*Value{
		"hello": &Value{
			Kind: &Value_RawBytes{RawBytes: []byte("world")},
		},
		"event": &Value{
			Kind: &Value_Map{
				Map: &ValueMap{
					Fields: map[string]*Value{
						"created": &Value{
							Kind: &Value_Timestamp{Timestamp: timestampNow},
						},
						"ingested": &Value{
							Kind: &Value_Timestamp{Timestamp: timestampNow},
						},
						"count": &Value{
							Kind: &Value_Integer{Integer: 3},
						},
					},
				},
			},
		},
	},
}

var timestampNow *types.Timestamp

func init() {
	unixNano := time.Now().UnixNano()
	sec := unixNano % int64(time.Nanosecond)
	nsec := int32(unixNano - sec)
	timestampNow = &types.Timestamp{Seconds: sec, Nanos: nsec}
}

func TestSerialization(t *testing.T) {
	e := &EventWrapper{
		Event: &EventWrapper_Log{Log: logEvent},
	}

	data, err := proto.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Size", len(data))
}



func TestJsonSerialization(t *testing.T) {
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	log.Println("Size", len(data))
}

func BenchmarkSerializeProto(b *testing.B) {
	e := &EventWrapper{
		Event: &EventWrapper_Log{Log: logEvent},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := proto.Marshal(e)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSerializeJSON(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(m)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeserializeProto(b *testing.B) {
	e := &EventWrapper{
		Event: &EventWrapper_Log{Log: logEvent},
	}

	data, err := proto.Marshal(e)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var e EventWrapper
		err := proto.Unmarshal(data, &e)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeserializeJSON(b *testing.B) {
	data, err := json.Marshal(m)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]interface{}
		err := json.Unmarshal(data, &m)
		if err != nil {
			b.Fatal(err)
		}
	}
}
