package main

import "C"
import (
	"encoding/json"
	"fmt"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/pipeline"

	// Register
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/lowercase"
)

var pipe *pipeline.Pipeline

//export Load
func Load(jsonPipeline *C.char) int32 {
	var config pipeline.Config
	if err := json.Unmarshal([]byte(C.GoString(jsonPipeline)), &config); err != nil {
		fmt.Println(err)
		return 1
	}

	var err error
	pipe, err = pipeline.New(config)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	return 0
}

//export Process
func Process(input *C.char) *C.char {
	jsonData := C.GoString(input)

	var event *event.Event
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		return C.CString(err.Error())
	}

	out, err := pipe.Process(event)
	if err != nil {
		return C.CString(err.Error())
	}

	data, err := json.Marshal(out[0])
	if err != nil {
		return C.CString(err.Error())
	}

	return C.CString(string(data))
}

func main() {}
