package processor

// ErrorKeyMissing is returned by processors when source field is missing from
// the event.
//
// Use errors.Is(err, ErrorKeyMissing{}) to test if a returned error is an
// ErrorKeyMissing.
type ErrorKeyMissing struct {
	Key string // Key that was not found in the event.
}

func (e ErrorKeyMissing) Error() string {
	return "key <" + e.Key + "> is missing from event"
}

func (e ErrorKeyMissing) Is(target error) bool {
	_, ok := target.(ErrorKeyMissing)
	return ok
}
