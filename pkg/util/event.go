package util

import (
	"strings"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
)

// Append appends a value to an existing array. If a non-array value
// already exists then the value becomes the first element in the new
// array. This function does not deduplicate.
func Append(evt *event.Event, key string, item *event.Value) error {
	if item == nil {
		return nil
	}

	// Make the target value into an array.
	targetValue := evt.Get(key)
	if targetValue == nil {
		targetValue = event.Array()
	} else if targetValue.Type != event.ArrayType {
		targetValue = event.Array(targetValue)
	}

	// Append the new item to the array.
	if item.Type == event.ArrayType {
		for _, v := range item.Array {
			targetValue.Array = append(targetValue.Array, v)
		}
	} else {
		targetValue.Array = append(targetValue.Array, item)
	}

	// Overwrite the existing value.
	_, err := evt.Put(key, targetValue)
	return err
}

// EscapeKey escapes dots contained in a key. This is non-idempotent
// (do not use it on a key that is already escaped).
func EscapeKey(key string) string {
	if strings.IndexByte(key, '.') == -1 {
		return key
	}
	return strings.ReplaceAll(key, ".", `\.`)
}
