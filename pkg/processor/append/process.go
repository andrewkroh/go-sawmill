package append

import (
	"errors"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

func (p *Append) Process(evt processor.Event) error {
	v := evt.Get(p.config.Field)
	if v == nil {
		v = event.Array()
		evt.Put(p.config.Field, v)
	} else if v.Type == event.StringType {
		v = event.Array(v)
	} else if v.Type != event.ArrayType {
		return errors.New("value to append to is not array")
	}

	v.Array = append(v.Array, event.String(p.config.Value))
	return nil
}
