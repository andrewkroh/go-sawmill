package pipeline

import (
	"errors"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
)

type Pipeline struct {
	id         string
	processors []*pipelineProcessor
	onFailure  []*pipelineProcessor
}

func New(config *Config) (*Pipeline, error) {
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
func (pipe *Pipeline) Process(evt *event.Event) (*event.Event, error) {
	pipeEvt := &pipelineEvent{data: evt}

	if err := pipe.process(pipeEvt); err != nil {
		return nil, err
	}

	if pipeEvt.dropped {
		return nil, nil
	}

	return pipeEvt.data, nil
}

// TODO: There might need to be an internal and external facing interface.
// Outside users will have an event.Event while internally we will pass
// a processor.Event that allow the additional of metadata and an explicit
// drop method.
func (pipe *Pipeline) process(evt *pipelineEvent) error {
	var err error
	for _, proc := range pipe.processors {
		if err = proc.Process(evt); err != nil {
			// Go to global on_failure handler.
			break
		}
	}

	if err != nil && len(pipe.onFailure) > 0 {
		for _, proc := range pipe.onFailure {
			if err = proc.Process(evt); err != nil {
				// Failure in global on_failure.
				return err
			}
		}
	}

	if err != nil {
		return err
	}

	return nil
}
