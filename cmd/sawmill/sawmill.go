package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/pipeline"

	// Register processors:
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/append"
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/community_id"
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/lowercase"
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/remove"
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/set"
	_ "github.com/andrewkroh/go-event-pipeline/pkg/processor/uppercase"
)

var pipelineFile string

func init() {
	flag.StringVar(&pipelineFile, "p", "", "pipeline definition file")
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	c, err := loadPipeline(pipelineFile)
	if err != nil {
		log.Fatal("Error:", err)
	}

	// TODO: Consider making New accept a pointer.
	p, err := pipeline.New(c)
	if err != nil {
		log.Fatal("Error:", err)
	}

	if err := processInput(os.Stdin, p); err != nil {
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

func processInput(in io.Reader, pipe *pipeline.Pipeline) error {
	s := bufio.NewScanner(in)
	var lineNumber uint64

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
		out, err := pipe.Process(evt)
		if err != nil {
			log.Printf("Error processing line %d: %v", lineNumber, err)
			continue
		}

		data, err := json.Marshal(out)
		if err != nil {
			log.Printf("Unexpected error marshaling event from line %d to JSON: %v", lineNumber, err)
			continue
		}

		fmt.Println(string(data))
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("failed reading from input: %w", err)
	}

	return nil
}
