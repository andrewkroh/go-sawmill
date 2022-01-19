package pipeline

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
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

func New(config PipelineConfig) (*Pipeline, error) {
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
		fmt.Println("running on_failure")
		for _, proc := range pipe.onFailure {
			fmt.Println(proc.ID)
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

type pipelineProcessor struct {
	ID            string
	Condition     string
	IgnoreFailure bool
	IgnoreMissing bool
	process       func(event *pipelineEvent) ([]*pipelineEvent, error)
	OnFailure     []*pipelineProcessor

	// Metrics
	eventsIn      prometheus.Counter
	eventsOut     prometheus.Counter
	eventsDropped prometheus.Counter
}

func (p *pipelineProcessor) Process(event *pipelineEvent) ([]*pipelineEvent, error) {
	// TODO: Check the type of processor (single vs split).
	_, err := p.process(event)

	if err != nil && len(p.OnFailure) > 0 {
		for _, proc := range p.OnFailure {
			if _, err = proc.process(event); err != nil {
				break
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return []*pipelineEvent{event}, nil
}

func newPipelineProcessors(baseID string, procConfigs []ProcessorConfig) ([]*pipelineProcessor, error) {
	if len(procConfigs) == 0 {
		return nil, nil
	}

	processors := make([]*pipelineProcessor, 0, len(procConfigs))
	for i, processorConfig := range procConfigs {
		procType, options, err := processorConfig.getProcessor()
		if err != nil {
			return nil, err
		}

		pipeProc, err := newPipelineProcessor(baseID, i, procType, options)
		if err != nil {
			return nil, err
		}

		processors = append(processors, pipeProc)
	}

	return processors, nil
}

func newPipelineProcessor(baseID string, processorIndex int, processorType string, config ProcessorOptionConfig) (*pipelineProcessor, error) {
	// Psuedo JSON XPath expression.
	id := baseID + "[" + strconv.Itoa(processorIndex) + "]." + processorType

	labels := map[string]string{
		"component_kind": "processor",
		"component_type": processorType,
		"component_id":   id,
	}

	procIfc, err := registry.NewProcessor(processorType, config.Config)
	if err != nil {
		return nil, err
	}
	proc := procIfc.(processor.Processor)

	onFailureProcessors, err := newPipelineProcessors(id+".on_failure", config.OnFailure)
	if err != nil {
		return nil, err
	}

	p := &pipelineProcessor{
		ID: id,
		//Condition: config.If,
		process: func(event *pipelineEvent) ([]*pipelineEvent, error) {
			if err := proc.Process(event); err != nil {
				return nil, err
			}
			return []*pipelineEvent{event}, nil
		},
		OnFailure: onFailureProcessors,
		eventsIn: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "es",
			Name:        "component_events_in_total",
			Help:        "Total number of events in to component.",
			ConstLabels: labels,
		}),

		eventsOut: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "es",
			Name:        "component_events_out_total",
			Help:        "Total number of events out of component.",
			ConstLabels: labels,
		}),

		eventsDropped: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "es",
			Name:        "component_events_dropped_total",
			Help:        "Total number of events dropped by component.",
			ConstLabels: labels,
		}),
	}

	return p, nil
}
