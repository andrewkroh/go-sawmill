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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/sys"

	"github.com/andrewkroh/go-sawmill/pkg/eventutil"
	"github.com/andrewkroh/go-sawmill/pkg/processor"
)

var timeNow = time.Now

type Status int32

const (
	StatusOK Status = iota
	StatusInternalFailure
	StatusInvalidArgument
	StatusNotFound
)

var statusNames = map[Status]string{
	StatusOK:              "OK",
	StatusInternalFailure: "Internal Failure",
	StatusInvalidArgument: "Invalid Argument",
	StatusNotFound:        "Not Found",
}

func (s Status) String() string {
	if name, found := statusNames[s]; found {
		return name
	}
	return "Status " + strconv.Itoa(int(s))
}

func (s Status) Error() string {
	return s.String()
}

type contextKey string

type wazeroSession struct {
	runtime      wazero.Runtime
	processFunc  api.Function
	mallocFunc   api.Function
	registerFunc api.Function
}

func newWazeroSession(wasm []byte) (*wazeroSession, error) {
	// Choose the context to use for function calls.
	ctx := context.Background()

	// Create a new WebAssembly Runtime.
	r := wazero.NewRuntime()

	// Instantiate a Go-defined module named "env" that exports a function to
	// log to the console.
	elasticMod, err := r.NewModuleBuilder("elastic").
		ExportFunction("elastic_get_field", elasticGetField).
		ExportFunction("elastic_put_field", elasticPutField).
		ExportFunction("elastic_get_current_time_nanoseconds", elasticGetCurrentTimeNanoseconds).
		ExportFunction("elastic_log", elasticLog).
		Compile(ctx, wazero.NewCompileConfig())
	if err != nil {
		return nil, err
	}

	_, err = r.InstantiateModule(nil, elasticMod, wazero.NewModuleConfig())
	if err != nil {
		return nil, err
	}

	guestMod, err := r.CompileModule(ctx, wasm, wazero.NewCompileConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to compile: %w", err)
	}

	mod, err := r.InstantiateModule(ctx, guestMod, wazero.NewModuleConfig().WithName("instance_1"))
	if err != nil {
		return nil, err
	}

	// TODO: List export functions to determine the ABI version.

	// Get references to WebAssembly functions we'll use in this example.
	process := mod.ExportedFunction("process")
	malloc := mod.ExportedFunction("malloc")
	register := mod.ExportedFunction("register")

	return &wazeroSession{
		runtime:      r,
		processFunc:  process,
		mallocFunc:   malloc,
		registerFunc: register,
	}, nil
}

func elasticGetField(ctx context.Context, m api.Module, keyAddr, keySize, rtnPtrPtr, rtnSizePtr uint32) int32 {
	log.Println("elastic_get_field")

	var key string
	if data, ok := m.Memory().Read(ctx, keyAddr, keySize); ok {
		key = string(data)
	} else {
		return int32(StatusInvalidArgument)
	}
	log.Printf("elastic_get_field key=%v", key)

	evt := ctx.Value(contextKey("event")).(processor.Event)

	var data []byte
	if value := evt.Get(key); value != nil {
		var err error
		if data, err = json.Marshal(value); err != nil {
			log.Println("Error:", err)
			return int32(StatusInternalFailure)
		}
	} else {
		log.Println("Error:", StatusNotFound)
		return int32(StatusNotFound)
	}

	sess := ctx.Value(contextKey("session")).(*wazeroSession)

	addr, err := sess.malloc(len(data))
	if err != nil {
		log.Println("Error:", err)
		return int32(StatusInternalFailure)
	}

	if !m.Memory().Write(nil, addr, data) {
		return int32(StatusInternalFailure)
	}

	if !m.Memory().WriteUint32Le(nil, rtnPtrPtr, addr) {
		return int32(StatusInternalFailure)
	}

	if !m.Memory().WriteUint32Le(nil, rtnSizePtr, uint32(len(data))) {
		return int32(StatusInternalFailure)
	}

	log.Println("elastic_get_field status OK")
	return int32(StatusOK)
}

func elasticPutField(ctx context.Context, m api.Module, keyAddr, keySize, valueAddr, valueSize uint32) int32 {
	log.Println("elastic_put_field")

	var key string
	if data, ok := m.Memory().Read(ctx, keyAddr, keySize); ok {
		key = string(data)
	} else {
		log.Println("Failed read")
		return int32(StatusInvalidArgument)
	}

	var value interface{}
	if data, ok := m.Memory().Read(ctx, valueAddr, valueSize); ok {
		if err := json.Unmarshal(data, &value); err != nil {
			log.Println("Failed JSON unmarshal", string(data))
			return int32(StatusInvalidArgument)
		}
	} else {
		log.Println("Failed read")
		return int32(StatusInvalidArgument)
	}

	eventValue, err := eventutil.ReflectValue(value)
	if err != nil {
		log.Printf("Failed to create event.Value from %#v", value)
		return int32(StatusInvalidArgument)
	}

	evt := ctx.Value(contextKey("event")).(processor.Event)
	evt.Put(key, eventValue)

	return int32(StatusOK)
}

//     fn elastic_get_current_time_nanoseconds(return_time: *mut u64) -> Status;
func elasticGetCurrentTimeNanoseconds(ctx context.Context, m api.Module, returnTimePtr int32) int32 {
	log.Println("elastic_get_current_time_nanoseconds")
	now := timeNow().UnixNano()
	if !m.Memory().WriteUint64Le(nil, uint32(returnTimePtr), uint64(now)) {
		return int32(StatusInternalFailure)
	}
	return int32(StatusOK)
}

func elasticLog(ctx context.Context, m api.Module, level, messageDataAddr, messageSize uint32) int32 {
	var msg string
	if data, ok := m.Memory().Read(ctx, messageDataAddr, messageSize); ok {
		msg = string(data)
	} else {
		log.Println("Failed read")
		return int32(StatusInvalidArgument)
	}

	log.Printf("elastic_log [%v] %s", level, msg)

	return int32(StatusOK)
}

func (s *wazeroSession) Close() error {
	return s.Close()
}

func (s *wazeroSession) malloc(size int) (uint32, error) {
	rtn, err := s.mallocFunc.Call(context.Background(), uint64(size))
	if err != nil {
		return 0, err
	}

	// Return address of allocated memory.
	return uint32(rtn[0]), nil
}

func (s *wazeroSession) guestProcess(event processor.Event) error {
	ctx := context.WithValue(context.Background(), contextKey("session"), s)
	ctx = context.WithValue(ctx, contextKey("event"), event)

	rtns, err := s.processFunc.Call(ctx)
	if err != nil {
		var exitErr *sys.ExitError
		if !errors.As(err, &exitErr) || exitErr.ExitCode() != 0 {
			return err
		}
	}

	if len(rtns) > 0 {
		status := Status(rtns[0])
		if status != StatusOK {
			return status
		}
	}

	return nil
}
