package metric

import (
	"slices"

	"go.opentelemetry.io/otel/attribute"
)

type Labeler interface {
	Label() []attribute.KeyValue
}

type NoLabel struct{}

func (NoLabel) Label() []attribute.KeyValue {
	return nil
}

type LabelMap map[string]string

func (l LabelMap) Label() []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(l))
	for k, v := range l {
		attrs = append(attrs, attribute.String(k, v))
	}
	return slices.Clip(attrs)
}
