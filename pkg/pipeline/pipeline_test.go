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
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/andrewkroh/go-sawmill/pkg/event"
	"github.com/andrewkroh/go-sawmill/pkg/processor"
	"github.com/andrewkroh/go-sawmill/pkg/processor/registry"

	// Register processors for testing purposes.
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/append"
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/lowercase"
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/set"
)

var generateExpected = flag.Bool("g", false, "generate expected output")

const sampleConfigYAML = `
---
id: logs-sample
description: |-
  Parse sample data.

  Incoming data must follow RFC123 or else!
processors:
  - set:
      target_field: event.id
      value: "1234"
on_failure:
  - set:
      target_field: event.kind
      value: pipeline_error
`

const sampleConfigJSON = `
{
  "id": "logs-sample",
  "description": "Parse sample data.\n\nIncoming data must follow RFC123 or else!",
  "processors": [
    {
      "set": {
        "target_field": "event.id",
        "value": "1234"
      }
    }
  ],
  "on_failure": [
    {
      "set": {
        "target_field": "event.kind",
        "value": "pipeline_error"
      }
    }
  ]
}
`

func samplePipeline() *Config {
	return &Config{
		ID: "logs-sample",
		Description: `Parse sample data.

Incoming data must follow RFC123 or else!`,
		Processors: []ProcessorConfig{
			{
				"set": &ProcessorOptionConfig{
					Config: map[string]interface{}{
						"target_field": "event.id",
						"value":        "1234",
					},
				},
			},
		},
		OnFailure: []ProcessorConfig{
			{
				"set": &ProcessorOptionConfig{
					Config: map[string]interface{}{
						"target_field": "event.kind",
						"value":        "pipeline_error",
					},
				},
			},
		},
	}
}

type failProcessor struct {
	config failConfig
}

type failConfig struct {
	IgnoreFailure bool `config:"ignore_failure"`
}

func (f *failProcessor) Process(event processor.Event) error {
	return errors.New("fail processor failed")
}

func (f *failProcessor) Config() failConfig {
	return f.config
}

var _ processor.Processor = (*failProcessor)(nil)

func init() {
	registry.MustRegister("fail", func(c failConfig) (*failProcessor, error) {
		return &failProcessor{config: c}, nil
	})
}

func TestPipelineConfigYAMLUnmarshal(t *testing.T) {
	var p *Config
	dec := yaml.NewDecoder(bytes.NewBufferString(sampleConfigYAML))
	dec.KnownFields(true)
	require.NoError(t, dec.Decode(&p))

	assert.Equal(t, samplePipeline(), p)
}

func TestPipelineConfigJSONUnmarshal(t *testing.T) {
	var p *Config
	dec := json.NewDecoder(bytes.NewBufferString(sampleConfigJSON))
	dec.DisallowUnknownFields()
	require.NoError(t, dec.Decode(&p))

	assert.Equal(t, samplePipeline(), p)
}

func TestPipelineConfigJSONMarshal(t *testing.T) {
	out := new(bytes.Buffer)
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(samplePipeline())

	assert.Equal(t,
		strings.TrimSpace(sampleConfigJSON),
		strings.TrimSpace(out.String()))
}

func TestNew(t *testing.T) {
	pipe, err := New(samplePipeline())
	require.NoError(t, err)
	require.NotNil(t, pipe)
}

func getProcessor(m map[string]ProcessorOptionConfig) (string, ProcessorOptionConfig, error) {
	if len(m) != 1 {
		return "", ProcessorOptionConfig{}, errors.New("one and only one processor must be specified")
	}
	for k, v := range m {
		return k, v, nil
	}
	return "", ProcessorOptionConfig{}, nil
}

func TestPipeline(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		pipe, err := New(samplePipeline())
		require.NoError(t, err)

		evt, err := pipe.Process(newTestEvent())
		require.NoError(t, err)

		eventId := evt.Get("event.id")
		require.NotNil(t, eventId)
		assert.Equal(t, "1234", eventId.String)
	})

	t.Run("processor err", func(t *testing.T) {
		pipeline := &Config{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": &ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
					},
				},
			},
		}
		pipe, err := New(pipeline)
		require.NoError(t, err)

		evt, err := pipe.Process(newTestEvent())
		require.Error(t, err)
		require.Nil(t, evt)

		assert.Contains(t, err.Error(), "non_existent")
	})

	t.Run("processor err with local on_failure", func(t *testing.T) {
		pipeline := &Config{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": &ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
						OnFailure: []ProcessorConfig{
							{
								"set": &ProcessorOptionConfig{
									Config: map[string]interface{}{
										"target_field": "event.kind",
										"value":        "pipeline_error",
									},
								},
							},
						},
					},
				},
			},
		}
		pipe, err := New(pipeline)
		require.NoError(t, err)

		evt, err := pipe.Process(newTestEvent())
		require.NoError(t, err)

		kind := evt.Get("event.kind")
		require.NotNil(t, kind)
		assert.Equal(t, "pipeline_error", kind.String)
	})

	t.Run("pipeline err with global on_failure", func(t *testing.T) {
		pipeline := &Config{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": &ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
					},
				},
			},
			OnFailure: []ProcessorConfig{
				{
					"set": &ProcessorOptionConfig{
						Config: map[string]interface{}{
							"target_field": "event.kind",
							"value":        "pipeline_error",
						},
					},
				},
			},
		}
		pipe, err := New(pipeline)
		require.NoError(t, err)

		evt, err := pipe.Process(newTestEvent())
		require.NoError(t, err)

		kind := evt.Get("event.kind")
		require.NotNil(t, kind)
		assert.Equal(t, "pipeline_error", kind.String)
	})

	t.Run("pipeline err no global on_failure", func(t *testing.T) {
		pipeline := &Config{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": &ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
					},
				},
			},
		}
		pipe, err := New(pipeline)
		require.NoError(t, err)

		evts, err := pipe.Process(newTestEvent())
		require.Error(t, err)
		require.Nil(t, evts)

		assert.Contains(t, err.Error(), "non_existent")
	})
}

func newTestEvent() *event.Event {
	evt := event.New()
	evt.Put("vehicle.vin", event.String("1234"))
	evt.Put("vehicle.tag", event.String("VCX-9833"))
	return evt
}

func TestPipelines(t *testing.T) {
	pipelineFiles, err := filepath.Glob("testdata/*.pipeline.yml")
	require.NoError(t, err)

	for _, name := range pipelineFiles {
		t.Run(name, func(t *testing.T) {
			data, err := ioutil.ReadFile(name)
			require.NoError(t, err)

			var pipelineConfig *Config
			err = yaml.Unmarshal(data, &pipelineConfig)
			require.NoError(t, err)

			pipe, err := New(pipelineConfig)
			require.NoError(t, err)

			// Load events.
			prefix := strings.TrimSuffix(name, ".pipeline.yml")
			data, err = ioutil.ReadFile(prefix + ".events.json")
			require.NoError(t, err)

			type testEvents struct {
				Events []map[string]interface{}
			}
			var events []*event.Event
			err = json.Unmarshal(data, &events)
			require.NoError(t, err)

			type outputEvent struct {
				Index int
				Event *event.Event `json:"event,omitempty"`
				Error string       `json:"error,omitempty"`
			}
			outputs := make([]outputEvent, 0, len(events))
			for i, evt := range events {
				oe := outputEvent{Index: i}
				oe.Event, err = pipe.Process(evt)
				if err != nil {
					oe.Error = err.Error()
				}
				outputs = append(outputs, oe)
			}

			buf := new(bytes.Buffer)
			enc := json.NewEncoder(buf)
			enc.SetEscapeHTML(false)
			enc.SetIndent("", "  ")
			require.NoError(t, enc.Encode(outputs))

			actualJSON := buf.Bytes()

			if *generateExpected {
				err = ioutil.WriteFile(prefix+".events-expected.json", buf.Bytes(), 0o644)
				require.NoError(t, err)
			}

			expectedJSON, err := ioutil.ReadFile(prefix + ".events-expected.json")
			if err != nil {
				if os.IsNotExist(err) {
					t.Fatal("run tests with -g to generate expected file")
				}
				t.Fatal(err)
			}
			if diff := cmp.Diff(string(expectedJSON), string(actualJSON)); diff != "" {
				t.Fatalf("Found differences:\n%s", diff)
			}
		})
	}
}
