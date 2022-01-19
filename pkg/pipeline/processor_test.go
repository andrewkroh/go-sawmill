package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/andrewkroh/go-event-pipeline/pkg/processor/lowercase"
)

func TestIgnore(t *testing.T) {
	lc, err := lowercase.New(lowercase.Config{IgnoreMissing: true})
	require.NoError(t, err)

	ignoreMissing, ignoreError, err := ignores(lc)
	require.NoError(t, err)
	assert.True(t, *ignoreMissing)
	assert.Nil(t, ignoreError)
}
