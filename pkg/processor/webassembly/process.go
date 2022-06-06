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

package webassembly

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

func newProcessor(config Config) (*Webassembly, error) {
	wasmOrWAT, err := readFile(config.File)
	if err != nil {
		return nil, err
	}

	s, err := newWazeroSession(wasmOrWAT)
	if err != nil {
		return nil, err
	}

	return &Webassembly{
		config:  config,
		session: s,
	}, nil
}

func (p *Webassembly) Process(event processor.Event) error {
	return p.session.guestProcess(event)
}

func readFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var r io.ReadCloser = f
	if strings.ToLower(filepath.Ext(filename)) == ".gz" {
		r, err = gzip.NewReader(f)
		if err != nil {
			return nil, err
		}
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err = r.Close(); err != nil {
		return nil, err
	}

	return data, nil
}
