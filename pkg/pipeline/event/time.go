package event

import (
	"strconv"
	"time"
)

type Time struct {
	UnixNanos int64
}

func TimeNow() Time {
	return Time{UnixNanos: time.Now().UnixNano()}
}

func (t Time) GoTime() time.Time {
	return time.Unix(0, t.UnixNanos).UTC()
}

func (t Time) String() string {
	return strconv.FormatInt(t.UnixNanos, 10)
}

func (t Time) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.GoTime().Format(time.RFC3339Nano) + `"`), nil
}
