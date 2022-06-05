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

package proctemplate

import (
	"bufio"
	"os"

	"gopkg.in/yaml.v3"
)

type Processors struct {
	CommonFields map[string]interface{} `yaml:"common_fields"`
	Processors   []map[string]Processor
}

type Processor struct {
	Description   string
	Configuration []ConfigurationOption
}

type ConfigurationOption struct {
	Name        string
	Type        string
	Required    bool
	Optional    bool
	Default     interface{}
	Description string
}

func cleanSlice(in []interface{}) []interface{} {
	result := make([]interface{}, len(in))
	for i, v := range in {
		result[i] = cleanValue(v)
	}
	return result
}

func cleanMapInterface(in map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		key := k.(string)
		result[key] = cleanValue(v)
	}
	return result
}

func cleanMapString(in map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range in {
		result[k] = cleanValue(v)
	}
	return result
}

func cleanValue(v interface{}) interface{} {
	switch v := v.(type) {
	case []interface{}:
		return cleanSlice(v)
	case map[interface{}]interface{}:
		return cleanMapInterface(v)
	case map[string]interface{}:
		return cleanMapString(v)
	default:
		return v
	}
}

func ReadProcessorsYAMLFile(path string) (*Processors, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Decode the processors.yml and validate all fields are known.
	dec := yaml.NewDecoder(bufio.NewReader(f))
	dec.KnownFields(true)
	var p *Processors
	if err := dec.Decode(&p); err != nil {
		return nil, err
	}

	return p, nil
}
