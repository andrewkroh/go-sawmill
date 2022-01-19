package registry

import (
	"fmt"
	"reflect"

	"github.com/elastic/go-ucfg"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

var (
	errorInterface          = reflect.TypeOf((*error)(nil)).Elem()
	processorInterface      = reflect.TypeOf((*processor.Processor)(nil)).Elem()
	splitProcessorInterface = reflect.TypeOf((*processor.SplitProcessor)(nil)).Elem()
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
		return nil, fmt.Errorf("proc not found")
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
	if !processorType.Implements(processorInterface) && !processorType.Implements(splitProcessorInterface) {
		return nil, fmt.Errorf("processor must implement processor.Processor or processor.SplitProcessor")
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
	return pc, nil
}
