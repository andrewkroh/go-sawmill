package util

import "github.com/andrewkroh/go-event-pipeline/pkg/event"

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
