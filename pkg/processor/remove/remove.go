// Foo License
// Code generated by processor/generate.go - DO NOT EDIT.
package remove

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
)

func init() {
	registry.MustRegister(processorName, New)
}

const (
	processorName = "remove"
)

// Config contains the configuration options for the remove processor.
type Config struct {
	// Source fields to remove.
	Fields []string `config:"fields" validate:"required"`

	// If true and field does not exist or is null, the processor quietly
	// returns without modifying the document.
	IgnoreMissing bool `config:"ignore_missing"`
}

// InitDefaults initializes the configuration options to their default values.
func (c *Config) InitDefaults() {
	c.IgnoreMissing = false
}

// Removes existing fields. If one field doesn’t exist the processor
// will fail.
type Remove struct {
	config Config
}

// New returns a new Remove processor.
func New(config Config) (*Remove, error) {
	return &Remove{config: config}, nil
}

// Config returns the Remove processor config.
func (p *Remove) Config() Config {
	return p.config
}

func (p *Remove) String() string {
	return processor.ConfigString(processorName, p.config)
}

func (p *Remove) Process(event processor.Event) error {
	return nil
}
