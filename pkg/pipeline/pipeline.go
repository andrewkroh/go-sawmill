package pipeline

import (
	"errors"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
)

type Processor interface {
	Process(event *pipelineEvent) ([]*pipelineEvent, error)
}

//var _ Processor = (*Pipeline)(nil)

type Pipeline struct {
	id         string
	processors []*pipelineProcessor
	onFailure  []*pipelineProcessor
}

func New(config Config) (*Pipeline, error) {
	if config.ID == "" {
		return nil, errors.New("pipeline must have a non-empty id")
	}

	processors, err := newPipelineProcessors(config.ID+".processors", config.Processors)
	if err != nil {
		return nil, err
	}
	onFailureProcessors, err := newPipelineProcessors(config.ID+".on_failure", config.OnFailure)
	if err != nil {
		return nil, err
	}

	return &Pipeline{
		id:         config.ID,
		processors: processors,
		onFailure:  onFailureProcessors,
	}, nil
}

// Process transforms an event by processing it through the pipeline. There
// are four cases that callers should expect for return values.
//
//   Event pass through - The input event is returned as index 0 of the slice.
//   Dropped event - Empty slice and nil error.
//   Processing error - Empty slice and non-nil error.
//   Event split - Slice length is greater than 1 and non-nil error.
func (pipe *Pipeline) Process(evt *event.Event) ([]*event.Event, error) {
	pipeEvt := &pipelineEvent{data: evt}

	pipeEvts, err := pipe.process(pipeEvt)
	if err != nil {
		return nil, err
	}

	var out []*event.Event
	for _, evt := range pipeEvts {
		if !evt.dropped {
			out = append(out, evt.data)
		}
	}

	return out, nil
}

// TODO: There might need to be an internal and external facing interface.
// Outside users will have an event.Event while internally we will pass
// a processor.Event that allow the additional of metadata and an explicit
// drop method.
func (pipe *Pipeline) process(evt *pipelineEvent) ([]*pipelineEvent, error) {
	var err error
	for _, proc := range pipe.processors {
		var splitEvents []*pipelineEvent
		splitEvents, err = proc.Process(evt)
		if err != nil {
			// OnFailure
			break
		}
		_ = splitEvents
	}

	if err != nil && len(pipe.onFailure) > 0 {
		for _, proc := range pipe.onFailure {
			var splitEvents []*pipelineEvent
			splitEvents, err = proc.Process(evt)
			if err != nil {
				return nil, err
			}
			_ = splitEvents
		}
	}

	if err != nil {
		return nil, err
	}

	return []*pipelineEvent{evt}, nil
}
