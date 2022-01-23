package protobuf

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
)

var testTime = time.Now().UTC()

var m = map[string]interface{}{
	"hello":        "world",
	"dotted\\.key": "value",
	"event": map[string]interface{}{
		"created":  testTime.Format(time.RFC3339Nano),
		"ingested": testTime.Format(time.RFC3339Nano),
		"count":    3,
	},
}

func testEvent() *event.Event {
	evt := event.New()
	evt.Put("hello", event.String("world"))
	evt.Put("dotted\\.key", event.String("value"))
	evt.Put("event.created", event.Timestamp(testTime.UnixNano()))
	evt.Put("event.ingested", event.Timestamp(testTime.UnixNano()))
	evt.Put("event.count", event.Integer(3))
	return evt
}

func TestFromToEvent(t *testing.T) {
	evt := testEvent()

	// Encode
	pbData, err := proto.Marshal(FromEvent(evt))
	require.NoError(t, err)

	// Decode
	var m MessageWrapper
	require.NoError(t, proto.Unmarshal(pbData, &m))
	require.NotNil(t, m.GetLog(), "message must contain non-nil Log")

	// Protobuf -> event.Event
	outEvt := ToLogEvent(m.GetLog())

	assert.Equal(t, evt, outEvt)
}

func TestCompareSizes(t *testing.T) {
	// Protobuf
	pbData, err := proto.Marshal(FromEvent(testEvent()))
	require.NoError(t, err)

	// JSON
	jsonBytes, err := json.Marshal(testEvent())
	require.NoError(t, err)

	t.Log("Sizes:")
	t.Log("  Protobuf:", len(pbData))
	t.Log("  JSON:    ", len(jsonBytes))
}

func BenchmarkSerialize(b *testing.B) {
	b.Run("protobuf", func(b *testing.B) {
		m := FromEvent(testEvent())
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := proto.Marshal(m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("event json", func(b *testing.B) {
		evt := testEvent()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(evt)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("mapstr json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := json.Marshal(m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkDeserialize(b *testing.B) {
	b.Run("protobuf", func(b *testing.B) {
		protoBytes, err := proto.Marshal(FromEvent(testEvent()))
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			var m MessageWrapper
			err := proto.Unmarshal(protoBytes, &m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("event json", func(b *testing.B) {
		jsonBytes, err := json.Marshal(testEvent())
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var evt *event.Event
			err := json.Unmarshal(jsonBytes, &evt)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("mapstr json", func(b *testing.B) {
		jsonBytes, err := json.Marshal(m)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var m map[string]interface{}
			err := json.Unmarshal(jsonBytes, &m)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
