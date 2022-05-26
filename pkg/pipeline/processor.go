// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package pipeline

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
)

type pipelineProcessor struct {
	ID            string
	Condition     string // TODO: Not implemented.
	IgnoreFailure bool
	IgnoreMissing bool
	process       func(event *pipelineEvent) error
	OnFailure     []*pipelineProcessor

	// Metrics
	eventsIn      prometheus.Counter
	eventsOut     prometheus.Counter
	eventsDropped prometheus.Counter
}

func (p *pipelineProcessor) Process(event *pipelineEvent) error {
	// TODO: Check Condition before executing.

	err := p.process(event)

	// Ignore Missing
	if err != nil && p.IgnoreMissing && errors.Is(err, processor.ErrorKeyMissing{}) {
		return nil
	}

	// On Failure
	if err != nil && len(p.OnFailure) > 0 {
		for _, proc := range p.OnFailure {
			if err = proc.process(event); err != nil {
				break
			}
		}
	}

	// Ignore Failure
	if err != nil && !p.IgnoreFailure {
		return err
	}

	return nil
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

func newPipelineProcessor(baseID string, processorIndex int, processorType string, config *ProcessorOptionConfig) (*pipelineProcessor, error) {
	// Pseudo JSON XPath expression.
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
	// TODO: Change the interface to return a processor since splitProcessor is removed.
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
		process: func(event *pipelineEvent) error {
			if err := proc.Process(event); err != nil {
				return err
			}
			return nil
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
