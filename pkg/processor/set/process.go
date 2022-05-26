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

package set

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

func (p *Set) Process(evt processor.Event) error {
	var v *event.Value
	if p.config.Value != nil {
		// TODO: Make creating a value easier.
		switch x := p.config.Value.(type) {
		case string:
			v = event.String(x)
		default:
			panic("unhandled type")
		}
	} else if p.config.CopyFrom != "" {
		v = evt.Get(p.config.CopyFrom)
		if v == nil {
			return processor.ErrorKeyMissing{Key: p.config.CopyFrom}
		}
	}

	_, err := evt.Put(p.config.TargetField, v)
	return err
}
