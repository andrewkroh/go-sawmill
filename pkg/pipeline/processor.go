package pipeline

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
)

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

	ignoreMissingPtr, ignoreFailurePtr, err := ignores(procIfc)
	if err != nil {
		return nil, err
	}

	onFailureProcessors, err := newPipelineProcessors(id+".on_failure", config.OnFailure)
	if err != nil {
		return nil, err
	}

	p := &pipelineProcessor{
		ID:        id,
		Condition: string(config.If),
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
	if ignoreMissingPtr != nil {
		p.IgnoreMissing = *ignoreMissingPtr
	}
	if ignoreFailurePtr != nil {
		p.IgnoreFailure = *ignoreFailurePtr
	}

	return p, nil
}

func ignores(proc interface{}) (ignoreMissing, ignoreFailure *bool, err error) {
	zeroValue := reflect.Value{}
	p := reflect.ValueOf(proc)

	m := p.MethodByName("Config")
	if m == zeroValue {
		return nil, nil, fmt.Errorf("error Config() method not found")
	}

	if m.Type().NumOut() != 1 {
		return nil, nil, fmt.Errorf("error Config() must return one value")
	}
	out := m.Call(nil)[0]

	if v := out.FieldByName("IgnoreMissing"); v != zeroValue && v.Kind() == reflect.Bool {
		boolValue := v.Bool()
		ignoreMissing = &boolValue
	}
	if v := out.FieldByName("IgnoreFailure"); v != zeroValue && v.Kind() == reflect.Bool {
		boolValue := v.Bool()
		ignoreFailure = &boolValue
	}
	return ignoreMissing, ignoreFailure, nil
}
