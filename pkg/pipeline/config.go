package pipeline

import (
	"encoding/json"
	"errors"
)

type ConditionalExpressionConfig string

type PipelineConfig struct {
	ID          string            `yaml:"id,omitempty"          json:"id,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Processors  []ProcessorConfig `yaml:"processors,omitempty"  json:"processors,omitempty"`
	OnFailure   []ProcessorConfig `yaml:"on_failure,omitempty"  json:"on_failure,omitempty"`
}

type ProcessorConfig map[string]ProcessorOptionConfig

func (c ProcessorConfig) getProcessor() (name string, opts ProcessorOptionConfig, err error) {
	if len(c) != 1 {
		return "", ProcessorOptionConfig{}, errors.New("exactly one processor must be specified")
	}
	for k, v := range c {
		return k, v, nil
	}

	// Never invoked.
	return "", ProcessorOptionConfig{}, nil
}

type ProcessorOptionConfig struct {
	ID        string                      `yaml:"id,omitempty"         json:"id,omitempty"`
	If        ConditionalExpressionConfig `yaml:"if,omitempty"         json:"if,omitempty"`
	OnFailure []ProcessorConfig           `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`
	Config    map[string]interface{}      `yaml:",inline"              json:"-"                    config:",inline"`
}

// UnmarshalJSON contains a workaround for the lack of inline tag support in
// encoding/json.
func (c *ProcessorOptionConfig) UnmarshalJSON(data []byte) error {
	// Prevent another call in this UnmarshalJSON.
	type opts ProcessorOptionConfig
	if err := json.Unmarshal(data, (*opts)(c)); err != nil {
		return err
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	delete(raw, "id")
	delete(raw, "if")
	delete(raw, "on_failure")

	if len(raw) > 0 {
		c.Config = make(map[string]interface{}, len(raw))
	}
	for k, v := range raw {
		c.Config[k] = v
	}
	return nil
}

// MarshalJSON contains a workaround for the lack of inline tag support in
// encoding/json.
func (c ProcessorOptionConfig) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{}
	if c.ID != "" {
		data["id"] = c.ID
	}
	if c.If != "" {
		data["if"] = c.If
	}
	if len(c.OnFailure) > 0 {
		data["on_failure"] = c.OnFailure
	}
	for k, v := range c.Config {
		data[k] = v
	}

	return json.Marshal(data)
}
