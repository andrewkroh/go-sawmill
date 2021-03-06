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
	"encoding/json"
	"errors"
)

type Config struct {
	ID          string            `yaml:"id,omitempty"          json:"id,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Processors  []ProcessorConfig `yaml:"processors,omitempty"  json:"processors,omitempty"`
	OnFailure   []ProcessorConfig `yaml:"on_failure,omitempty"  json:"on_failure,omitempty"`
}

type ProcessorConfig map[string]*ProcessorOptionConfig

type ProcessorOptionConfig struct {
	ID        string                      `yaml:"id,omitempty"         json:"id,omitempty"`
	If        ConditionalExpressionConfig `yaml:"if,omitempty"         json:"if,omitempty"`
	OnFailure []ProcessorConfig           `yaml:"on_failure,omitempty" json:"on_failure,omitempty"`
	Config    map[string]interface{}      `yaml:",inline"              json:"-"                    config:",inline"`
}

type ConditionalExpressionConfig string

func (c ProcessorConfig) getProcessor() (name string, opts *ProcessorOptionConfig, err error) {
	if len(c) == 0 {
		return "", nil, errors.New("processor cannot be empty")
	}
	if len(c) > 1 {
		return "", nil, errors.New("only one processor must be specified")
	}
	for k, v := range c {
		if v == nil {
			v = &ProcessorOptionConfig{}
		}
		return k, v, nil
	}

	// Never invoked.
	return "", nil, errors.New("unexpected number of keys in processor")
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
