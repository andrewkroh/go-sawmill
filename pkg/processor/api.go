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

package processor

import (
	"encoding/json"
	"strings"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
)

type Event interface {
	Put(key string, v *event.Value) (*event.Value, error)
	TryPut(key string, v *event.Value) (*event.Value, error)
	Get(key string) *event.Value
	Delete(key string) *event.Value

	// Cancel any further processing by the pipeline for this event.
	Cancel()

	// Drop the event.
	Drop()
}

type Processor interface {
	Process(event Event) error
}

func ConfigString(name string, v interface{}) string {
	var buf strings.Builder
	buf.WriteString(name)
	buf.WriteByte('=')

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
	return buf.String()
}
