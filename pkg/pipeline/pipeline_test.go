package pipeline

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
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

func samplePipeline() Config {
	return Config{
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
	var p Config
	dec := yaml.NewDecoder(bytes.NewBufferString(sampleConfigYAML))
	dec.KnownFields(true)
	require.NoError(t, dec.Decode(&p))

	assert.Equal(t, samplePipeline(), p)
}

func TestPipelineConfigJSONUnmarshal(t *testing.T) {
	var p Config
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

		evts, err := pipe.Process(newTestEvent())
		require.NoError(t, err)
		require.Len(t, evts, 1)

		out := evts[0]
		eventId := out.Get("event.id")
		require.NotNil(t, eventId)
		assert.Equal(t, "1234", eventId.String)
	})

	t.Run("processor err", func(t *testing.T) {
		pipeline := Config{
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

		evts, err := pipe.Process(newTestEvent())
		require.Error(t, err)
		require.Nil(t, evts)

		assert.Contains(t, err.Error(), "non_existent")
	})

	t.Run("processor err with local on_failure", func(t *testing.T) {
		pipeline := Config{
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

		evts, err := pipe.Process(newTestEvent())
		require.NoError(t, err)
		require.Len(t, evts, 1)

		out := evts[0]
		kind := out.Get("event.kind")
		require.NotNil(t, kind)
		assert.Equal(t, "pipeline_error", kind.String)
	})

	t.Run("pipeline err with global on_failure", func(t *testing.T) {
		pipeline := Config{
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

		evts, err := pipe.Process(newTestEvent())
		require.NoError(t, err)
		require.Len(t, evts, 1)

		out := evts[0]
		kind := out.Get("event.kind")
		require.NotNil(t, kind)
		assert.Equal(t, "pipeline_error", kind.String)
	})

	t.Run("pipeline err no global on_failure", func(t *testing.T) {
		pipeline := Config{
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

			var pipelineConfig Config
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
				Index  int
				Events []*event.Event
				Error  string `json:"error,omitempty"`
			}
			outputs := make([]outputEvent, 0, len(events))
			for i, evt := range events {
				oe := outputEvent{Index: i}
				oe.Events, err = pipe.Process(evt)
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
				err = ioutil.WriteFile(prefix+".events-expected.json", buf.Bytes(), 0644)
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

var generateExpected = flag.Bool("g", false, "generate expected output")

func TestExecute(t *testing.T) {
	pipeline := pipe{
		id: "root",
		procs: []pipe{
			{
				id:    "A",
			},
			{
				id:    "B",
			},
			{
				id:    "C",
			},
			{
				id:    "D",
				fail: []pipe{
					{
						id:    "D1",
					},
					{
						id:    "D2",
					},
				},
			},
			{
				id:    "E",
			},
		},
		fail: []pipe{
			{
				id:    "F",
			},
		},
	}

	rootProc := pipelineToGraph(&pipeline)
	assert.Equal(t, "A", rootProc.next.id)
	assert.Equal(t, "B", rootProc.next.next.id)
	assert.Equal(t, "C", rootProc.next.next.next.id)
	assert.Equal(t, "D", rootProc.next.next.next.next.id)
	assert.Equal(t, "D1", rootProc.next.next.next.next.fail.id)
	assert.Equal(t, "D2", rootProc.next.next.next.next.fail.next.id)
	assert.Equal(t, "E", rootProc.next.next.next.next.next.id)
	assert.Equal(t, "F", rootProc.fail.id)
	fmt.Println(execute(rootProc))


	p := &proc{
		do: makeDo("A", false),
		next: &proc{
			do:   makeDo("B", true),
			fail: &proc{do: makeDo("C", true)},
		},
	}
	fmt.Println(execute(p))

	root := makeNode("root", false, nil)
	a := makeNode("A", false, root)
	b := makeNode("B", true, a)
	b.fail = makeNode("B.1", false, nil)
	fail2 := makeNode("B.2", true, b.fail)
	fail2.fail = makeNode("B.2.1", false, nil)

	_ = makeNode("C", false, b)
	fmt.Println(execute(root))
}

func pipelineToGraph(pipe *pipe) *proc {
	return &proc{
		id: pipe.id,
		do: makeDo(pipe.id, false),
		next: pipelineProcsToGraph(pipe.procs),
		fail: pipelineProcsToGraph(pipe.fail),
	}
}

func pipelineProcsToGraph(pipes []pipe) *proc {
	var list []*proc
	for _, p := range pipes {
		list = append(list, pipelineToGraph(&p))
	}
	if len(list) == 0 {
		return nil
	}
	item := list[0]
	for _, pr := range list[1:] {
		item.next = pr
		item = pr
	}
	return list[0]
}

func makeNode(name string, fail bool, parent *proc) *proc {
	n := &proc{
		id: name,
		do: makeDo(name, fail),
	}
	if parent != nil {
		parent.next = n
	}
	return n
}

func makeDo(name string, fail bool) func() error {
	return func() error {
		fmt.Printf("%s->", name)
		if fail {
			return fmt.Errorf("failed in %s", name)
		}
		return nil
	}
}

type pipe struct {
	id    string
	procs []pipe
	fail  []pipe
}

type proc struct {
	id   string
	do   func() error
	next *proc
	fail *proc
}

func execute(p *proc) error {
	stack := &stack{}
	stack.Push(p)

	for stack.Len() > 0 {
		x := stack.Pop()

		if x.next != nil {
			stack.Push(x.next)
		}

		if err := x.do(); err != nil {
			if x.fail != nil {
				stack.Push(x.fail)
				continue
			}
			return err
		}
	}
	return nil
}

type stack struct {
	data []*proc
}

func (s *stack) Push(p *proc) {
	s.data = append(s.data, p)
}

func (s *stack) Pop() *proc {
	size := len(s.data)
	if size == 0 {
		return nil
	}
	item := s.data[size-1]
	s.data = s.data[:size-1]
	return item
}

func (s *stack) Len() int {
	return len(s.data)
}
