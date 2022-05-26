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

package event

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTimeUnix = 1642121157123456789
	testTimeISO  = "2022-01-14T00:45:57.123456789Z"
)

func TestTimeMarshalJSON(t *testing.T) {
	v := Time{testTimeUnix}

	data, err := json.Marshal(v)
	require.NoError(t, err)
	t.Log(string(data))

	assert.Equal(t, `"`+testTimeISO+`"`, string(data))
}
