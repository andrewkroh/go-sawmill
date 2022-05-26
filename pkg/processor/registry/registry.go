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

package registry

import (
	"fmt"
	"reflect"

	"github.com/elastic/go-ucfg"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

var (
	errorInterface     = reflect.TypeOf((*error)(nil)).Elem()
	processorInterface = reflect.TypeOf((*processor.Processor)(nil)).Elem()
)

var constructors = NewRegistry()

func MustRegister(name string, constructorFunc interface{}) {
	if err := constructors.Register(name, constructorFunc); err != nil {
		panic(err)
	}
}

func NewProcessor(name string, config map[string]interface{}) (interface{}, error) {
	return constructors.NewProcessor(name, config)
}

type processorConstructor struct {
	// newConfig returns a pointer to a zero value config.
	newConfig func() reflect.Value

	newProc func(config reflect.Value) (procPtr interface{}, err error)
}

type Registry struct {
	procs map[string]processorConstructor
}

func NewRegistry() *Registry {
	return &Registry{procs: map[string]processorConstructor{}}
}

func (r *Registry) Register(name string, constructorFunc interface{}) error {
	if _, found := r.procs[name]; found {
		return fmt.Errorf("%q processor is already registered", name)
	}
	pc, err := validateConfig(constructorFunc)
	if err != nil {
		return err
	}
	r.procs[name] = *pc
	return nil
}

func (r *Registry) NewProcessor(name string, config map[string]interface{}) (interface{}, error) {
	pc, found := r.procs[name]
	if !found {
		return nil, fmt.Errorf("processor type %q not found", name)
	}

	procConfigValue := pc.newConfig()

	uConf, err := ucfg.NewFrom(config)
	if err != nil {
		return nil, err
	}

	if err := uConf.Unpack(procConfigValue.Interface()); err != nil {
		return nil, err
	}

	return pc.newProc(procConfigValue)
}

func (r *Registry) clear() {
	r.procs = map[string]processorConstructor{}
}

func validateConfig(i interface{}) (*processorConstructor, error) {
	v := reflect.ValueOf(i)
	t := v.Type()

	if t.Kind() != reflect.Func {
		return nil, fmt.Errorf("value must be a function, but got %v", t.Kind())
	}

	// Check config.
	if t.NumIn() != 1 {
		return nil, fmt.Errorf("function must accept one arg")
	}
	configType := t.In(0)
	if configType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("function must be config struct")
	}

	if t.NumOut() != 2 {
		return nil, fmt.Errorf("function must return 2 values")
	}
	processorType := t.Out(0)
	if processorType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("function should return a pointer to a processor struct")
	}
	if !processorType.Implements(processorInterface) {
		return nil, fmt.Errorf("processor must implement processor.Processor")
	}

	errorType := t.Out(1)
	if !errorType.Implements(errorInterface) {
		return nil, fmt.Errorf("function should return an error")
	}

	pc := &processorConstructor{
		newConfig: func() reflect.Value {
			return reflect.New(configType)
		},
		newProc: func(config reflect.Value) (procPtr interface{}, err error) {
			if config.Type().Kind() == reflect.Ptr {
				config = config.Elem()
			}
			out := v.Call([]reflect.Value{config})
			if !out[1].IsNil() {
				return nil, out[1].Interface().(error)
			}
			return out[0].Interface(), nil
		},
	}
	v.Call([]reflect.Value{reflect.New(configType).Elem()})

	// TODO: Validate that each processor implements a 'Conifg() struct' method.
	return pc, nil
}
