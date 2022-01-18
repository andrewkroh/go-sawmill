// Foo License
// Code generated by processor/generate.go - DO NOT EDIT.
package uppercase

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
)

func init() {
	registry.MustRegister(processorName, New)
}

const (
	processorName = "uppercase"
)

// Config contains the configuration options for the uppercase processor.
type Config struct {
	// Source field to process.
	Field string `config:"field" validate:"required"`

	// If true and field does not exist or is null, the processor quietly
	// returns without modifying the document.
	IgnoreMissing bool `config:"ignore_missing"`

	// The field to assign the output value to, by default field is updated
	// in-place.
	TargetField string `config:"target_field"`
}

// InitDefaults initializes the configuration options to their default values.
func (c *Config) InitDefaults() {
	c.IgnoreMissing = false
}

// Uppercase converts a string to its uppercase equivalent. If the field is an
// array of strings, all members of the array will be converted.
type Uppercase struct {
	config Config
}

// New returns a new Uppercase processor.
func New(config Config) (*Uppercase, error) {
	return &Uppercase{config: config}, nil
}

// Config returns the Uppercase processor config.
func (p *Uppercase) Config() Config {
	return p.config
}

func (p *Uppercase) String() string {
	return processor.ConfigString(processorName, p.config)
}

func (p *Uppercase) Process(event processor.Event) error {
	return nil
}
