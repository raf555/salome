package metric

import (
	"fmt"
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

type LabelMap map[string]any

func (l LabelMap) Label() []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(l))
	for k, v := range l {
		attrs = append(attrs, l.toAttribute(k, v))
	}
	return slices.Clip(attrs)
}

func (l LabelMap) toAttribute(key string, v any) attribute.KeyValue {
	switch val := v.(type) {
	case bool:
		return attribute.Bool(key, val)
	case int:
		return attribute.Int(key, val)
	case int64:
		return attribute.Int64(key, val)
	case float64:
		return attribute.Float64(key, val)
	case string:
		return attribute.String(key, val)
	case error:
		if val != nil {
			return attribute.String(key, val.Error())
		}
		return attribute.String(key, "")
	}

	if stringer, ok := v.(fmt.Stringer); ok {
		return attribute.Stringer(key, stringer)
	}

	return attribute.String(key, fmt.Sprintf("%+v", v))
}
