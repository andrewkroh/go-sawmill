package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	// Register processors for testing purposes.
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/lowercase"
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/set"
)

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

func samplePipeline() PipelineConfig {
	return PipelineConfig{
		ID: "logs-sample",
		Description: `Parse sample data.

Incoming data must follow RFC123 or else!`,
		Processors: []ProcessorConfig{
			{
				"set": ProcessorOptionConfig{
					Config: map[string]interface{}{
						"target_field": "event.id",
						"value":        "1234",
					},
				},
			},
		},
		OnFailure: []ProcessorConfig{
			{
				"set": ProcessorOptionConfig{
					Config: map[string]interface{}{
						"target_field": "event.kind",
						"value":        "pipeline_error",
					},
				},
			},
		},
	}
}

func TestPipelineConfigYAMLUnmarshal(t *testing.T) {
	var p PipelineConfig
	dec := yaml.NewDecoder(bytes.NewBufferString(sampleConfigYAML))
	dec.KnownFields(true)
	require.NoError(t, dec.Decode(&p))

	assert.Equal(t, samplePipeline(), p)
}

func TestPipelineConfigJSONUnmarshal(t *testing.T) {
	var p PipelineConfig
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
		evt := newTestEvent()

		pipe, err := New(samplePipeline())
		require.NoError(t, err)

		pipelineEvent := &Event{data: evt}
		evts, err := pipe.Process(pipelineEvent)
		require.NoError(t, err)
		require.Len(t, evts, 1)

		out := evts[0]
		eventId := out.Get("event.id")
		require.NotNil(t, eventId)
		assert.Equal(t, "1234", eventId.Bytes)
	})

	t.Run("processor err", func(t *testing.T) {
		evt := newTestEvent()

		pipeline := PipelineConfig{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
					},
				},
			},
		}
		pipe, err := New(pipeline)
		require.NoError(t, err)

		pipelineEvent := &Event{data: evt}
		evts, err := pipe.Process(pipelineEvent)
		require.Error(t, err)
		require.Nil(t, evts)

		assert.Contains(t, err.Error(), "non_existent")
	})

	t.Run("processor err with local on_failure", func(t *testing.T) {
		evt := newTestEvent()

		pipeline := PipelineConfig{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
						OnFailure: []ProcessorConfig{
							{
								"set": ProcessorOptionConfig{
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

		pipelineEvent := &Event{data: evt}
		evts, err := pipe.Process(pipelineEvent)
		require.NoError(t, err)
		require.Len(t, evts, 1)

		out := evts[0]
		kind := out.Get("event.kind")
		require.NotNil(t, kind)
		assert.Equal(t, "pipeline_error", kind.Bytes)
	})

	t.Run("pipeline err with global on_failure", func(t *testing.T) {
		evt := newTestEvent()

		pipeline := PipelineConfig{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
					},
				},
			},
			OnFailure: []ProcessorConfig{
				{
					"set": ProcessorOptionConfig{
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

		pipelineEvent := &Event{data: evt}
		evts, err := pipe.Process(pipelineEvent)
		require.NoError(t, err)
		require.Len(t, evts, 1)

		out := evts[0]
		kind := out.Get("event.kind")
		require.NotNil(t, kind)
		assert.Equal(t, "pipeline_error", kind.Bytes)
	})

	t.Run("pipeline err no global on_failure", func(t *testing.T) {
		evt := newTestEvent()

		pipeline := PipelineConfig{
			ID: "lowercase-non-existent",
			Processors: []ProcessorConfig{
				{
					"lowercase": ProcessorOptionConfig{
						Config: map[string]interface{}{
							"field": "non_existent",
						},
					},
				},
			},
		}
		pipe, err := New(pipeline)
		require.NoError(t, err)

		pipelineEvent := &Event{data: evt}
		evts, err := pipe.Process(pipelineEvent)
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
