package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestConfig(t *testing.T) {
	const yml = `
---
processors:
  - fail:
  - set:
on_failure:
  - set:
`
	var conf Config
	err := yaml.Unmarshal([]byte(yml), &conf)
	require.NoError(t, err)
	require.Len(t, conf.Processors, 2)
	require.Len(t, conf.OnFailure, 1)

	name, _, err := conf.Processors[0].getProcessor()
	require.NoError(t, err)
	assert.Equal(t, "fail", name)
}
