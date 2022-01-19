package pipeline

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

var _ processor.Event = (*pipelineEvent)(nil)

// pipelineEvent implements the processor.Event interface. It wraps an
// event.Event and contains state about the event w.r.t. the pipeline.
type pipelineEvent struct {
	data      *event.Event
	cancelled bool
	dropped   bool
}

func (e *pipelineEvent) Put(key string, v *event.Value) (*event.Value, error) {
	return e.data.Put(key, v)
}

func (e *pipelineEvent) TryPut(key string, v *event.Value) (*event.Value, error) {
	return e.data.TryPut(key, v)
}

func (e *pipelineEvent) Get(key string) *event.Value {
	return e.data.Get(key)
}

func (e *pipelineEvent) Delete(key string) *event.Value {
	return e.data.Delete(key)
}

func (e *pipelineEvent) Cancel() {
	e.cancelled = true
}

func (e pipelineEvent) Drop() {
	e.dropped = true
}
