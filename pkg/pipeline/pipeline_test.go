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

	"github.com/andrewkroh/go-event-pipeline/pkg/util"

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
	t.Run("success", func(t *testing.T) {

		pipeline := pipe{
			id: "root",
			procs: []pipe{
				{
					id: "A",
				},
				{
					id: "B",
				},
				{
					id: "C",
				},
				{
					id: "D",
					fail: []pipe{
						{
							id: "D1",
						},
						{
							id: "D2",
						},
					},
				},
				{
					id: "E",
				},
			},
			fail: []pipe{
				{
					id: "F",
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

		fmt.Println(graphToString(rootProc))

		out, err := execute(newTestEvent(), rootProc)
		if assert.NoError(t, err) {
			fmt.Println()
			for i, e := range out {
				j, _ := e.MarshalJSON()
				assert.Contains(t, string(j), `"root","A","B","C","D","E"`)
				fmt.Printf("%d: %v", i, string(j))
			}
			fmt.Println()
		}
	})

	t.Run("global-on_failure", func(t *testing.T) {
		pipeline := pipe{
			id: "root",
			procs: []pipe{
				{
					id: "A",
				},
				{
					id: "B-fail",
					fail: []pipe{
						{
							id: "C-fail",
						},
					},
				},
			},
			fail: []pipe{
				{
					id: "D",
				},
			},
		}

		p := pipelineToGraph(&pipeline)
		fmt.Println(graphToString(p))

		out, err := execute(newTestEvent(), p)
		if assert.NoError(t, err) {
			fmt.Println()
			for i, e := range out {
				j, _ := e.MarshalJSON()
				assert.Contains(t, string(j), `"root","A","B-fail","C-fail","D"`)
				fmt.Printf("%d: %v", i, string(j))
			}
			fmt.Println()
		}
	})

	t.Run("branching failures", func(t *testing.T) {
		root := makeNode("root", nil)
		a := makeNode("A", root)
		b := makeNode("B-fail", a)
		b.fail = makeNode("B.1", nil)
		fail2 := makeNode("B.2-fail", b.fail)
		fail2.fail = makeNode("B.2.1", nil)

		_ = makeNode("C", b)

		fmt.Println(graphToString(root))

		out, err := execute(newTestEvent(), root)
		if assert.NoError(t, err) {
			fmt.Println()
			for i, e := range out {
				j, _ := e.MarshalJSON()
				assert.Contains(t, string(j), `"root","A","B-fail","B.1","B.2-fail","B.2.1","C"`)
				fmt.Printf("%d: %v", i, string(j))
			}
			fmt.Println()
		}
	})
}

func graphToString(p *proc) string {
	var sb strings.Builder
	addGraphNode(p, 0, &sb)
	return sb.String()
}

func addGraphNode(p *proc, indent int, sb *strings.Builder) {
	sb.WriteString(strings.Repeat(" ", indent))
	sb.WriteString(p.id)
	sb.WriteByte('\n')
	if p.fail != nil {
		sb.WriteString(strings.Repeat(" ", indent))
		sb.WriteString("|--\n")
		addGraphNode(p.fail, indent+2, sb)
	}
	if p.next != nil {
		addGraphNode(p.next, indent, sb)
	}
}

func pipelineToGraph(pipe *pipe) *proc {
	return &proc{
		id:   pipe.id,
		do:   makeDo(pipe.id),
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

func makeNode(name string, parent *proc) *proc {
	n := &proc{
		id: name,
		do: makeDo(name),
	}
	if parent != nil {
		parent.next = n
	}
	return n
}

func makeDo(name string) func(event *event.Event) ([]*event.Event, error) {
	return func(evt *event.Event) ([]*event.Event, error) {
		//fmt.Printf("%s->", name)
		util.Append(evt, "path", event.String(name))
		if strings.Contains(name, "fail") {
			return nil, fmt.Errorf("failed in %s", name)
		}
		return []*event.Event{evt}, nil
	}
}

type pipe struct {
	id    string
	procs []pipe
	fail  []pipe
}

type proc struct {
	id   string
	do   func(event *event.Event) ([]*event.Event, error)
	next *proc
	fail *proc
}

func execute(evt *event.Event, p *proc) ([]*event.Event, error) {
	stack := &stack{}
	if p.fail != nil {
		stack.Push(p.fail)
	}
	stack.Push(p)
	return executeStack(evt, stack)
}

func executeStack(evt *event.Event, stack *stack) ([]*event.Event, error) {
	for stack.Len() > 0 {
		x := stack.Pop()

		if x.next != nil {
			stack.Push(x.next)
		}

		out, err := x.do(evt)
		if err != nil {
			if x.fail != nil {
				stack.Push(x.fail)
				continue
			}
			fmt.Println(len(stack.data))
			for _, node := range stack.data {
				fmt.Println(node.id, node.fail != nil)
			}
			return nil, err
		}

		switch {
		case len(out) > 1:
			// Process each event individually from here.
			var accumulate []*event.Event
			for _, evt := range out {
				splitOut, err := executeStack(evt, stack.Clone())
				if err != nil {
					return nil, err
				}
				accumulate = append(accumulate, splitOut...)
			}
			return accumulate, nil
		case len(out) == 1:
			evt = out[0]
		case len(out) == 0:
			return nil, nil
		}
	}

	return []*event.Event{evt}, nil
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

func (s *stack) Clone() *stack {
	data := make([]*proc, len(s.data))
	copy(data, s.data)
	return &stack{
		data: data,
	}
}
