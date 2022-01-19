package event

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleEvent() {
	e := New()
	e.Put("event.category", Array(String("network"), String("authentication")))
	e.Put("event.type", Array(String("denied")))
	e.Put("event.created", Timestamp(Time{testTimeUnix}))
	e.Put("foo\\.bar", Integer(1))

	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	// Output:
	// {
	//   "event": {
	//     "category": [
	//       "network",
	//       "authentication"
	//     ],
	//     "created": "2022-01-14T00:45:57.123456789Z",
	//     "type": [
	//       "denied"
	//     ]
	//   },
	//   "foo.bar": 1
	// }
}

func TestEventPut(t *testing.T) {
	val := String("val")

	t.Run("put", func(t *testing.T) {
		e := New()
		old, err := e.Put("a", val)
		require.NoError(t, err)
		assert.Nil(t, old)
		assert.Equal(t, val, e.fields["a"])
	})

	t.Run("put nested", func(t *testing.T) {
		e := New()
		old, err := e.Put("a.b", val)
		require.NoError(t, err)
		assert.Nil(t, old)
		assert.Equal(t, val, e.fields["a"].Object["b"])
	})

	t.Run("put returns err for empty key", func(t *testing.T) {
		e := New()
		old, err := e.Put("", val)
		assert.ErrorIs(t, err, ErrEmptyKey)
		assert.Nil(t, old)
	})

	t.Run("put overwrites", func(t *testing.T) {
		original := String("original")
		e := New()
		e.init()
		e.fields["a"] = original

		old, err := e.Put("a", val)
		require.NoError(t, err)
		assert.Equal(t, original, old)
		assert.Equal(t, val, e.fields["a"])
	})

	t.Run("try-put does not overwrite", func(t *testing.T) {
		original := String("original")
		e := New()
		e.init()
		e.fields["a"] = original

		old, err := e.TryPut("a", val)
		require.ErrorIs(t, err, ErrKeyExists)
		assert.Equal(t, original, old)
		assert.Equal(t, original, e.fields["a"])
	})

	t.Run("put returns err on non-object target", func(t *testing.T) {
		original := String("original")
		e := New()
		e.init()
		e.fields["a"] = original

		old, err := e.Put("a.b", val)
		assert.ErrorIs(t, err, ErrTargetKeyNotObject)
		assert.Nil(t, old)
		assert.Equal(t, original, e.fields["a"])
	})
}

func TestEventGet(t *testing.T) {
	val := String("val")

	t.Run("get", func(t *testing.T) {
		e := New()
		e.init()
		e.fields["a"] = val
		e.Get("a")
	})

	t.Run("get nested", func(t *testing.T) {
		e := New()
		e.init()
		e.fields["a"] = Object(map[string]*Value{
			"b": val,
		})
		assert.Equal(t, val, e.Get("a.b"))
	})

	t.Run("get uninitialized event", func(t *testing.T) {
		e := New()
		assert.Nil(t, e.Get("a"))
	})

	t.Run("get key not found returns nil", func(t *testing.T) {
		e := New()
		e.init()
		assert.Nil(t, e.Get("a"))
	})
}

func TestEventDelete(t *testing.T) {
	val := String("val")

	t.Run("delete uninitialized event", func(t *testing.T) {
		e := New()
		assert.Nil(t, e.Delete(""))
	})

	t.Run("delete empty key", func(t *testing.T) {
		e := New()
		e.init()
		assert.Nil(t, e.Delete(""))
	})

	t.Run("delete key found", func(t *testing.T) {
		e := New()
		e.init()
		e.fields["a"] = val
		assert.Equal(t, val, e.Delete("a"))
	})

	t.Run("delete nested", func(t *testing.T) {
		e := New()
		e.init()
		e.fields["a"] = Object(map[string]*Value{
			"b": val,
		})
		assert.Equal(t, val, e.Delete("a.b"))
	})

	t.Run("delete key not found", func(t *testing.T) {
		e := New()
		e.init()
		assert.Nil(t, e.Delete("a"))
	})
}

func TestEvent(t *testing.T) {
	const from, to = "a", "b"
	val := String("val")

	t.Run("rename", func(t *testing.T) {
		e := New()
		e.Put("a", val)

		e.Put(to, e.Delete(from))
		assert.Nil(t, e.Get(from))
		assert.EqualValues(t, val, e.Get(to))
	})

	t.Run("rename no overwrite", func(t *testing.T) {
		original := String("original")
		e := New()
		e.Put(to, original)
		e.Put(from, val)

		if v := e.Delete(from); v != nil {
			if _, err := e.TryPut(to, v); err != nil {
				// Revert
				e.Put(from, v)
			}
		}

		assert.Len(t, e.fields, 2)
		assert.Equal(t, val, e.Get(from))
		assert.Equal(t, original, e.Get(to))
	})

	t.Run("rename non-existent field", func(t *testing.T) {
		e := New()
		e.Put(to, e.Delete(from))
		assert.Nil(t, e.Get(to))
	})
}

func TestKeyToPath(t *testing.T) {
	var testCases = []struct {
		key  string
		path []string
	}{
		{
			key:  `foo`,
			path: []string{"foo"},
		},
		{
			key:  `.foo`,
			path: []string{"foo"},
		},
		{
			key:  `\.foo`,
			path: []string{".foo"},
		},
		{
			key:  `foo.`,
			path: []string{"foo"},
		},
		{
			key:  `foo\.`,
			path: []string{"foo."},
		},
		{
			key:  `.`,
			path: nil,
		},
		{
			key:  `\.`,
			path: []string{"."},
		},
		{
			key:  "foo.bar",
			path: []string{"foo", "bar"},
		},
		{
			key:  `foo\.bar`,
			path: []string{"foo.bar"},
		},
	}

	for _, tc := range testCases {
		observedPath := keyToPath(tc.key)
		assert.Equal(t, tc.path, observedPath, "expected key=%q to produce [%s]", tc.key, strings.Join(tc.path, ", "))
	}
}

func TestPathString(t *testing.T) {
	assert.Equal(t, "/", pathString(nil))
	assert.Equal(t, "/", pathString([]string{}))
	assert.Equal(t, "/event", pathString([]string{"event"}))
	assert.Equal(t, "/event/ingested", pathString([]string{"event", "ingested"}))
	assert.Equal(t, "/ecs.version", pathString([]string{"ecs.version"}))
}
