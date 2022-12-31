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

	"github.com/prometheus/client_golang/prometheus"

	"github.com/andrewkroh/go-sawmill/pkg/event"
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
//	Event pass through - The input event is returned as index 0 of the slice.
//	Dropped event - Empty slice and nil error.
//	Processing error - Empty slice and non-nil error.
//	Event split - Slice length is greater than 1 and non-nil error.
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

func (pipe *Pipeline) ID() string {
	return pipe.id
}

func (pipe *Pipeline) Metrics() []prometheus.Collector {
	var metrics []prometheus.Collector
	pipe.visitProcessors(func(proc *pipelineProcessor) {
		metrics = append(metrics,
			proc.metricDiscardedEventsTotal,
			proc.metricErrorsTotal,
			proc.metricEventsInTotal,
			proc.metricEventsOutTotal,
		)
	})
	return metrics
}

func (pipe *Pipeline) visitProcessors(visit func(processor *pipelineProcessor)) {
	for _, proc := range pipe.processors {
		visitProcessor(visit, proc)
	}
}

func visitProcessor(visit func(processor *pipelineProcessor), proc *pipelineProcessor) {
	visit(proc)

	for _, proc := range proc.OnFailure {
		visitProcessor(visit, proc)
	}
}
