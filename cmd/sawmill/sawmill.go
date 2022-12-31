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

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/andrewkroh/go-sawmill/pkg/event"
	"github.com/andrewkroh/go-sawmill/pkg/metrics"
	"github.com/andrewkroh/go-sawmill/pkg/pipeline"

	// Register processors:
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/append"
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/community_id"
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/lowercase"
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/remove"
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/set"
	_ "github.com/andrewkroh/go-sawmill/pkg/processor/uppercase"
)

var (
	pipelineFile      string
	metricsListenAddr string
	cpuProfile        string
	memProfile        string
)

func init() {
	flag.StringVar(&pipelineFile, "p", "", "pipeline definition file")
	flag.StringVar(&metricsListenAddr, "metrics-addr", "localhost:9003", "Metrics listen address.")

	flag.StringVar(&cpuProfile, "cpuprofile", "", "CPU profile output")
	flag.StringVar(&memProfile, "memprofile", "", "memory profile output")
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	if cpuProfile != "" {
		bw, flush := bufferedFileWriter(cpuProfile)
		pprof.StartCPUProfile(bw)
		defer flush()
		defer pprof.StopCPUProfile()
	}

	if memProfile != "" {
		bw, flush := bufferedFileWriter(memProfile)
		defer func() {
			runtime.GC() // materialize all statistics
			if err := pprof.WriteHeapProfile(bw); err != nil {
				log.Fatal(err)
			}
			flush()
		}()
	}

	c, err := loadPipeline(pipelineFile)
	if err != nil {
		log.Fatal("Error:", err)
	}

	p, err := pipeline.New(c)
	if err != nil {
		log.Fatal("Error:", err)
	}

	metrics.Register(p.Metrics()...)
	defer metrics.Unregister(p.Metrics()...)
	metrics.Listen(metricsListenAddr)

	if err := processInput(os.Stdin, os.Stdout, p); err != nil {
		log.Fatal("Error:", err)
	}
}

func loadPipeline(path string) (*pipeline.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var c *pipeline.Config
	if err := yaml.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}

	return c, nil
}

func processInput(in io.Reader, out io.Writer, pipe *pipeline.Pipeline) error {
	s := bufio.NewScanner(in)
	var lineNumber uint64

	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)

	for s.Scan() {
		lineNumber++

		// Skip empty lines.
		line := strings.TrimSpace(s.Text())
		if line == "" {
			continue
		}

		// Create new event.
		evt := event.New()
		evt.Put("@metadata.line_number", event.UnsignedInteger(lineNumber))
		evt.Put("event.original", event.String(line))

		// Process the event.
		evt, err := pipe.Process(evt)
		if err != nil {
			log.Printf("Error processing line %d: %v", lineNumber, err)
			continue
		}

		if err := enc.Encode(evt); err != nil {
			log.Printf("Unexpected error marshaling event from line %d to JSON: %v", lineNumber, err)
			continue
		}
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("failed reading from input: %w", err)
	}

	return nil
}

func bufferedFileWriter(dest string) (w io.Writer, close func()) {
	f, err := os.Create(dest)
	if err != nil {
		log.Fatal(err)
	}
	bw := bufio.NewWriter(f)
	return bw, func() {
		if err := bw.Flush(); err != nil {
			log.Fatalf("error flushing %v: %v", dest, err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
