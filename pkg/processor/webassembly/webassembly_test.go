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

package webassembly

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andrewkroh/go-sawmill/pkg/event"
)

func init() {
	timeNow = fakeTimeNow
}

func TestWebassembly_Process(t *testing.T) {
	conf := Config{}
	conf.InitDefaults()
	conf.File = "testdata/modify_fields.wasm.gz"

	p, err := New(conf)
	require.NoError(t, err)

	evt := &pipelineEvent{
		data: event.New(),
	}

	err = p.Process(evt)
	require.NoError(t, err)

	data, err := json.Marshal(evt.data)
	require.NoError(t, err)
	t.Log(string(data))

	// TODO: Add a smaller, simpler module to testdata.
	expected := `{"bool":true,"float":1.2,"integer":1,"null":null,"object":{"hello":"world!"},"string":"hello"}`
	assert.JSONEq(t, expected, string(data))
}

func TestWebassemblyProcess(t *testing.T) {
	conf := Config{}
	conf.InitDefaults()
	conf.File = "testdata/demo.wasm.gz"

	p, err := New(conf)
	require.NoError(t, err)

	evt := &pipelineEvent{
		data: event.New(),
	}
	evt.data.Put("event.original", event.String("hello world"))

	err = p.Process(evt)
	require.NoError(t, err)

	data, err := json.Marshal(evt.data)
	require.NoError(t, err)
	t.Log(string(data))

	expected := `{"event":{"created":"2022-06-05T22:00:33.123456789+00:00","original":"hello world"},"message":"hello world"}`
	assert.JSONEq(t, expected, string(data))
}

// TODO: Clean this up. It was copied from the pipeline package.

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

func fakeTimeNow() time.Time {
	return time.Date(2022, 6, 5, 22, 0, 33, 123456789, time.UTC)
}
