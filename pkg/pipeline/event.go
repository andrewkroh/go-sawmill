package pipeline

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

var _ processor.Event = (*Event)(nil)

type Event struct {
	data      *event.Event
	cancelled bool
	dropped   bool
}

func (e *Event) Put(key string, v *event.Value) (*event.Value, error) {
	return e.data.Put(key, v)
}

func (e *Event) TryPut(key string, v *event.Value) (*event.Value, error) {
	return e.data.TryPut(key, v)
}

func (e *Event) Get(key string) *event.Value {
	return e.data.Get(key)
}

func (e *Event) Delete(key string) *event.Value {
	return e.data.Delete(key)
}

func (e *Event) Cancel() {
	e.cancelled = true
}

func (e Event) Drop() {
	e.dropped = true
}
