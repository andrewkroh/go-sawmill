package processor

import (
	"encoding/json"
	"strings"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
)

type Event interface {
	Put(key string, v *event.Value) (*event.Value, error)
	TryPut(key string, v *event.Value) (*event.Value, error)
	Get(key string) *event.Value
	Delete(key string) *event.Value

	// Cancel any further processing by the pipeline for this event.
	Cancel()

	// Drop the event.
	Drop()
}

type Processor interface {
	Process(event Event) error
}

type SplitProcessor interface {
	// Process a single event and possibly return multiple.
	// TODO: Clarify.
	Process(event Event) ([]Event, error)
}

func ConfigString(name string, v interface{}) string {
	var buf strings.Builder
	buf.WriteString(name)
	buf.WriteByte('=')

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	return buf.String()
}
