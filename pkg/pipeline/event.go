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
	"github.com/andrewkroh/go-sawmill/pkg/event"
	"github.com/andrewkroh/go-sawmill/pkg/processor"
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
