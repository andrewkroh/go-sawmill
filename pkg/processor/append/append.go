// Foo License
// Code generated by processor/generate.go - DO NOT EDIT.
package append

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor/registry"
)

func init() {
	registry.MustRegister(processorName, New)
}

const (
	processorName = "append"
)

// Config contains the configuration options for the append processor.
type Config struct {
	// If false, the processor does not append values already present in the
	// field.
	AllowDuplicates bool `config:"allow_duplicates"`

	// Source field to process.
	Field string `config:"field" validate:"required"`

	// If true and field does not exist or is null, the processor quietly
	// returns without modifying the document.
	IgnoreMissing bool `config:"ignore_missing"`

	// The value to be appended.
	Value string `config:"value"`
}

// InitDefaults initializes the configuration options to their default values.
func (c *Config) InitDefaults() {
	c.AllowDuplicates = false
	c.IgnoreMissing = false
}

// Appends one or more values to an existing array if the field already exists
// and it is an array. Converts a scalar to an array and appends one or more
// values to it if the field exists and it is a scalar. Creates an array
// containing the provided values if the field doesn’t exist. Accepts a single
// value or an array of values.
type Append struct {
	config Config
}

// New returns a new Append processor.
func New(config Config) (*Append, error) {
	return &Append{config: config}, nil
}

// Config returns the Append processor config.
func (p *Append) Config() Config {
	return p.config
}

func (p *Append) String() string {
	return processor.ConfigString(processorName, p.config)
}
