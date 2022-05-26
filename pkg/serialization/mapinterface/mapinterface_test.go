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

package mapinterface

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
)

var testTime = time.Now().UTC()

type myInt int32

type myStruct struct {
	Path string
	Hash struct {
		SHA256 string
	}
}

type Agent struct {
	BuildOriginal string `event:"build.original"`
}

type myTime time.Time

var m = map[string]interface{}{
	"hello":        "world",
	"dotted\\.key": "value",
	"event": map[string]interface{}{
		"created":    testTime,
		"ingested":   testTime,
		"start":      myTime(testTime),
		"count":      3,
		"risk_score": 0.51,
		"sequence":   myInt(14),
	},
	"related": map[string]interface{}{
		"ip": []string{
			"1.1.1.1",
			"8.8.8.8",
		},
	},
	"final":  true,
	"source": nil,
	"destination": map[interface{}]interface{}{
		"port": 53,
		18:     "foo",
	},
	"file": &myStruct{
		Path: "/root",
		Hash: struct {
			SHA256 string
		}{"a6a036e31e8bb26cf47dd8d3bee915debee906e4cf399ff7da29ca2b785d8cee"},
	},
	"user": map[string]interface{}{
		"name": interface{}(nil),
	},
	"error": map[string]interface{}{
		"message": fmt.Errorf("failure in pipeline"),
	},
	"agent": Agent{
		BuildOriginal: "7.16.2",
	},
}

func testEvent() *event.Event {
	evt := event.New()
	evt.Put("hello", event.String("world"))
	evt.Put("dotted\\.key", event.String("value"))
	evt.Put("event.created", event.Timestamp(testTime.UnixNano()))
	evt.Put("event.ingested", event.Timestamp(testTime.UnixNano()))
	evt.Put("event.start", event.Timestamp(testTime.UnixNano()))
	evt.Put("event.count", event.Integer(3))
	evt.Put("event.risk_score", event.Float(0.51))
	evt.Put("event.sequence", event.Integer(14))
	evt.Put("final", event.Bool(true))
	evt.Put("source", event.NullValue)
	evt.Put("related.ip", event.Array(event.String("1.1.1.1"), event.String("8.8.8.8")))
	evt.Put("destination.port", event.Integer(53))
	evt.Put("destination.18", event.String("foo"))
	evt.Put("file.path", event.String("/root"))
	evt.Put("file.hash.sha256", event.String("a6a036e31e8bb26cf47dd8d3bee915debee906e4cf399ff7da29ca2b785d8cee"))
	evt.Put("user.name", event.NullValue)
	evt.Put("error.message", event.String("failure in pipeline"))

	// TODO: It would be nice to allow dots in tag names and convert them
	// to nested.
	evt.Put("agent.build\\.original", event.String("7.16.2"))
	return evt
}

func TestToEvent(t *testing.T) {
	out, err := ToEvent(m)
	require.NoError(t, err)
	assert.Equal(t, testEvent(), out)
}

func TestFromEvent(t *testing.T) {
	mapifc := FromEvent(testEvent())

	expected := map[string]interface{}{
		"hello":      "world",
		"dotted.key": "value",
		"event": map[string]interface{}{
			"created":    testTime,
			"ingested":   testTime,
			"start":      testTime,
			"count":      int64(3),
			"risk_score": 0.51,
			"sequence":   int64(14),
		},
		"final":  true,
		"source": nil,
		"related": map[string]interface{}{
			"ip": []interface{}{
				"1.1.1.1",
				"8.8.8.8",
			},
		},
		"destination": map[string]interface{}{
			"port": int64(53),
			"18":   "foo",
		},
		"file": map[string]interface{}{
			"path": "/root",
			"hash": map[string]interface{}{
				"sha256": "a6a036e31e8bb26cf47dd8d3bee915debee906e4cf399ff7da29ca2b785d8cee",
			},
		},
		"user": map[string]interface{}{
			"name": nil,
		},
		"error": map[string]interface{}{
			"message": "failure in pipeline",
		},
		"agent": map[string]interface{}{
			"build.original": "7.16.2",
		},
	}

	assert.Equal(t, expected, mapifc)
}
