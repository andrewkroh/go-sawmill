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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

var _ processor.Processor = (*DummyProc)(nil)

type DummyConfig struct {
	IgnoreFailure bool `config:"ignore_failure"`
}

type DummyProc struct {
	config DummyConfig
}

func (d *DummyProc) Process(event processor.Event) error {
	panic("implement me")
}

func newDummyProc(conf DummyConfig) (*DummyProc, error) {
	return &DummyProc{config: conf}, nil
}

func TestRegister(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register("dummy", newDummyProc))

	config := map[string]interface{}{
		"ignore_failure": true,
	}
	p, err := r.NewProcessor("dummy", config)
	require.NoError(t, err)

	d, ok := p.(*DummyProc)
	require.True(t, ok)
	assert.True(t, d.config.IgnoreFailure)
}
