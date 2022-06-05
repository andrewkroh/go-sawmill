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

// Code generated by processor/generate.go - DO NOT EDIT.
package set

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/config"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
)

func init() {
	registry.MustRegister(processorName, New)
}

const (
	processorName = "set"
)

// Config contains the configuration options for the set processor.
type Config struct {
	// The origin field which will be copied to target_field.
	CopyFrom string `config:"copy_from"`

	// Ignore failures for the processor.
	IgnoreFailure bool `config:"ignore_failure"`

	// If true and field does not exist or is null, the processor quietly
	// returns without modifying the document.
	IgnoreMissing bool `config:"ignore_missing"`

	// The field to assign the output value to, by default field is updated
	// in-place.
	TargetField string `config:"target_field"`

	// The value to be set for the field.
	Value config.EventValue `config:"value"`
}

// InitDefaults initializes the configuration options to their default values.
func (c *Config) InitDefaults() {
	c.IgnoreFailure = false
	c.IgnoreMissing = false
}

// Sets one field and associates it with the specified value. If the field
// already exists, its value will be replaced with the provided one.
type Set struct {
	config Config
}

// New returns a new Set processor.
func New(config Config) (*Set, error) {
	return &Set{config: config}, nil
}

// Config returns the Set processor config.
func (p *Set) Config() Config {
	return p.config
}

func (p *Set) String() string {
	return processor.ConfigString(processorName, p.config)
}
