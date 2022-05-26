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

//go:build js && wasm
// +build js,wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/pipeline"

	// Register
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/lowercase"
)

func process(this js.Value, args []js.Value) interface{} {
	if len(args) != 2 {
		return "pipeline_execute requires two args"
	}

	pipelineJSON := args[0].String()
	eventJSON := args[1].String()

	var pipelineConfig *pipeline.Config
	if err := json.Unmarshal([]byte(pipelineJSON), &pipelineConfig); err != nil {
		return "failed to unmarshal pipeline: " + err.Error()
	}

	pipe, err := pipeline.New(pipelineConfig)
	if err != nil {
		return "failed to create new pipeline: " + err.Error()
	}

	var event *event.Event
	if err = json.Unmarshal([]byte(eventJSON), &event); err != nil {
		return "failed to unmarshal event JSON: " + err.Error()
	}

	out, err := pipe.Process(event)
	if err != nil {
		return "failed to process event: " + err.Error()
	}

	return toObject(out[0])
}

// toObject converts a struct to a map[string]interface{} using JSON
// marshal/unmarshal.
func toObject(v interface{}) map[string]interface{} {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	var out map[string]interface{}
	if err = json.Unmarshal(data, &out); err != nil {
		panic(err)
	}

	return out
}

func registerCallbacks() {
	js.Global().Set("pipeline_execute", js.FuncOf(process))
}

func main() {
	println("Pipeline WASM Demo loaded.")
	println("Invoke the pipeline_execute(pipeline_json, input_event_json) function to test.")
	registerCallbacks()
	<-make(chan bool)
}
