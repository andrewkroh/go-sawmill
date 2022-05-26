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

package lowercase

import (
	"errors"
	"strings"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

func (p *Lowercase) Process(evt processor.Event) error {
	v := evt.Get(p.config.Field)
	if v == nil {
		return processor.ErrorKeyMissing{Key: p.config.Field}
	}

	if v.Type != event.StringType {
		return errors.New("value to lowercase is not a string")
	}

	out := strings.ToLower(v.String)

	targetField := p.config.Field
	if p.config.TargetField != "" {
		targetField = p.config.TargetField
	}

	_, err := evt.Put(targetField, event.String(out))
	return err
}
