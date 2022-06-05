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
	OnFailure     []*pipelineProcessor

	proc processor.Processor

	// Metrics
	metricDiscardedEventsTotal prometheus.Counter // Explicit drops by the processor.
	metricErrorsTotal          prometheus.Counter // Total errors (not including ignored or recovered errors).
	metricEventsInTotal        prometheus.Counter // Received events.
	metricEventsOutTotal       prometheus.Counter // Successfully output events.
}

func (p *pipelineProcessor) Process(event *pipelineEvent) error {
	// TODO: Check Condition before executing.
	p.metricEventsInTotal.Inc()

	if err := p.proc.Process(event); err != nil {
		if event.dropped {
			p.metricDiscardedEventsTotal.Inc()
			return nil
		}

		// Ignore Missing
		if p.IgnoreMissing && errors.Is(err, processor.ErrorKeyMissing{}) {
			p.metricEventsOutTotal.Inc()
			return nil
		}

		// On Failure
		if len(p.OnFailure) > 0 {
			for _, proc := range p.OnFailure {
				if err = proc.Process(event); err != nil {
					break
				}
			}
		}

		// Ignore Failure
		if p.IgnoreFailure && err != nil {
			p.metricEventsOutTotal.Inc()
			return nil
		}

		// Could not recover from the error or ignore it.
		p.metricErrorsTotal.Inc()
		return err
	}

	p.metricEventsOutTotal.Inc()
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

	proc, err := registry.NewProcessor(processorType, config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed constructing processor with ID %s: %w", id, err)
	}

	ignoreMissingPtr, ignoreFailurePtr, err := ignores(proc)
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
		OnFailure: onFailureProcessors,
		proc:      proc,
		metricDiscardedEventsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "es",
			Name:        "component_discarded_events_total",
			Help:        "Total number of events dropped by component.",
			ConstLabels: labels,
		}),
		metricErrorsTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "es",
			Name:        "component_errors_total",
			Help:        "Total number of errors by component.",
			ConstLabels: labels,
		}),
		metricEventsInTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "es",
			Name:        "component_received_events_total",
			Help:        "Total number of events received by component.",
			ConstLabels: labels,
		}),
		metricEventsOutTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   "es",
			Name:        "component_sent_events_total",
			Help:        "Total number of events sent by component.",
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
