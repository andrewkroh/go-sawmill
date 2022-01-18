package set

import (
	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

func (p *Set) Process(evt processor.Event) error {
	var v *event.Value
	if p.config.Value != nil {
		// TODO: Make creating a value easier.
		switch x := p.config.Value.(type) {
		case string:
			v = event.String(x)
		default:
			panic("unhandled type")
		}
	} else if p.config.CopyFrom != "" {
		v = evt.Get(p.config.CopyFrom)
		if v == nil {
			return processor.ErrorKeyMissing{Key: p.config.CopyFrom}
		}
	}

	_, err := evt.Put(p.config.TargetField, v)
	return err
}
