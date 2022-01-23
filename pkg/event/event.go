package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrKeyExists          = errors.New("key already exists")
	ErrTargetKeyNotObject = errors.New("target key is not an object")
	ErrEmptyKey           = errors.New("key name is empty")
)

type Event struct {
	fields map[string]*Value
}

// New returns a new Event.
func New() *Event {
	return &Event{}
}

func (e *Event) init() {
	e.fields = map[string]*Value{}
}

// Put puts a value into the map. If the key already exists then it will be
// overwritten. If the target value is not an object then ErrTargetKeyNotObject
// is returned.
func (e *Event) Put(key string, val *Value) (old *Value, err error) {
	return e.put(keyToPath(key), val, true)
}

// TryPut puts a value into the map if the key does not exist. If the key
// already exists it will return the existing value and ErrKeyExists. If the
// target value is not an object then ErrTargetKeyNotObject is returned.
func (e *Event) TryPut(key string, val *Value) (existing *Value, err error) {
	return e.put(keyToPath(key), val, false)
}

func (e *Event) put(path []string, val *Value, overwrite bool) (old *Value, err error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("event put failed: %w", ErrEmptyKey)
	}
	if val == nil {
		return nil, nil
	}

	if e.fields == nil {
		e.init()
	}

	old = e.get(path)
	if old != nil && !overwrite {
		return old, fmt.Errorf("event put failed for path <%s>: %w", pathString(path), ErrKeyExists)
	}

	m := e.fields
	for _, key := range path[:len(path)-1] {
		inner, found := m[key]
		if found {
			if inner.Type != ObjectType {
				return nil, fmt.Errorf("event put failed for path <%s>: %w", pathString(path), ErrTargetKeyNotObject)
			}
			m = inner.Object
			continue
		}
		newObj := Object(map[string]*Value{})
		m[key] = newObj
		m = newObj.Object
	}
	m[path[len(path)-1]] = val
	return old, nil
}

// Get returns the value associated to the given key. It returns nil if the key
// does not exist.
func (e *Event) Get(key string) *Value {
	return e.get(keyToPath(key))
}

func (e *Event) get(path []string) *Value {
	if len(e.fields) == 0 {
		return nil
	}
	if len(path) == 0 {
		// Return root object.
		return Object(e.fields)
	}

	m := e.fields
	for _, key := range path[:len(path)-1] {
		inner, found := m[key]
		if !found || len(inner.Object) == 0 {
			return nil
		}
		m = inner.Object
	}

	return m[path[len(path)-1]]
}

// Delete removes the given key. It returns the deleted value if it existed.
func (e *Event) Delete(key string) (deleted *Value) {
	return e.delete(keyToPath(key))
}

func (e *Event) delete(path []string) (deleted *Value) {
	if len(path) == 0 || len(e.fields) == 0 {
		return nil
	}

	v := e.get(path)
	if v != nil {
		delete(e.fields, path[0])
	}

	return v
}

func (e *Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.fields)
}

func (e *Event) UnmarshalJSON(data []byte) error {
	var fields map[string]*Value
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	e.fields = fields
	return nil
}

// keyToPath creates an object path from a dot-separated key. If a path name
// contains a dot then the dot must be escaped by a backslash.
//
//    foo.bar = [foo, bar]
//    foo\.bar = [foo.bar]
func keyToPath(key string) []string {
	var paths []string
	var pathScratch []byte
	var escape bool
	for i := 0; i < len(key); i++ {
		switch c := key[i]; c {
		case '\\':
			escape = true
		case '.':
			if escape {
				pathScratch = append(pathScratch, '.')
			} else {
				if len(pathScratch) > 0 {
					paths = append(paths, string(pathScratch))
					pathScratch = pathScratch[:0]
				}
			}
			escape = false
		default:
			pathScratch = append(pathScratch, c)
			escape = false
		}
	}

	if len(pathScratch) > 0 {
		paths = append(paths, string(pathScratch))
	}

	return paths
}

func pathString(path []string) string {
	switch len(path) {
	case 0:
		return "/"
	case 1:
		return "/" + path[0]
	}

	var sb strings.Builder
	for _, elem := range path {
		sb.WriteByte('/')
		sb.WriteString(elem)
	}

	return sb.String()
}
