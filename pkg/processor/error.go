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

// ErrorKeyMissing is returned by processors when source field is missing from
// the event.
//
// Use errors.Is(err, ErrorKeyMissing{}) to test if a returned error is an
// ErrorKeyMissing.
type ErrorKeyMissing struct {
	Key string // Key that was not found in the event.
}

func (e ErrorKeyMissing) Error() string {
	return "key <" + e.Key + "> is missing from event"
}

func (e ErrorKeyMissing) Is(target error) bool {
	_, ok := target.(ErrorKeyMissing)
	return ok
}
