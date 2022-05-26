package uppercase

import (
	"errors"
	"strings"

	"github.com/andrewkroh/go-event-pipeline/pkg/event"
	"github.com/andrewkroh/go-event-pipeline/pkg/processor"
)

func (p *Uppercase) Process(evt processor.Event) error {
	v := evt.Get(p.config.Field)
	if v == nil {
		return processor.ErrorKeyMissing{Key: p.config.Field}
	}

	if v.Type != event.StringType {
		return errors.New("value to uppercase is not a string")
	}

	out := strings.ToUpper(v.String)

	targetField := p.config.Field
	if p.config.TargetField != "" {
		targetField = p.config.TargetField
	}

	_, err := evt.Put(targetField, event.String(out))
	return err
}
