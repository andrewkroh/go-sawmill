package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

var _ processor.Processor = (*DummyProc)(nil)

type DummyConfig struct {
	IgnoreFailure bool `config:"ignore_failure"`
}

type DummyProc struct {
	config DummyConfig
}

func (d *DummyProc) Process(event processor.Event) error {
	panic("implement me")
}

func newDummyProc(conf DummyConfig) (*DummyProc, error) {
	return &DummyProc{config: conf}, nil
}

func TestRegister(t *testing.T) {
	r := NewRegistry()
	require.NoError(t, r.Register("dummy", newDummyProc))

	config := map[string]interface{}{
		"ignore_failure": true,
	}
	p, err := r.NewProcessor("dummy", config)
	require.NoError(t, err)

	d, ok := p.(*DummyProc)
	require.True(t, ok)
	assert.True(t, d.config.IgnoreFailure)
}
