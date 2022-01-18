package event

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTimeUnix = 1642121157123456789
	testTimeISO  = "2022-01-14T00:45:57.123456789Z"
)

func TestTimeMarshalJSON(t *testing.T) {
	v := Time{testTimeUnix}

	data, err := json.Marshal(v)
	require.NoError(t, err)
	t.Log(string(data))

	assert.Equal(t, `"`+testTimeISO+`"`, string(data))
}
