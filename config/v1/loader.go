package config

import (
	"context"
	"fmt"
)

// LoadConfigTo loads config to T from the provider. The loaded config is the one initially read by the provider.
func LoadConfigTo[T any](provider Provider) (T, error) {
	ctx := context.TODO()

	var dst T

	cfg, err := provider.Config(ctx)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("provider.Config: %w", err)
	}

	if err := loadConfigFromMapTo(ctx, &dst, cfg); err != nil {
		var zero T
		return zero, fmt.Errorf("loadConfigFromMapTo: %w", err)
	}

	return dst, nil
}

type DynamicConfigGetter[T any] interface {
	Get() T
}

type getterFunc[T any] func() T

func (f getterFunc[T]) Get() T {
	return f()
}

// LoadDynamicConfigTo loads dynamic config to T from the provider.
// It is a syntactic sugar for mgr.RegisterConfig and GetConfig.
// Config is updated periodically by the manager.
func LoadDynamicConfigTo[T any](mgr DynamicConfigManager) (DynamicConfigGetter[T], error) {
	var key T
	if err := mgr.RegisterConfig(key, func() any {
		var zero T
		return &zero
	}); err != nil {
		return nil, fmt.Errorf("mgr.RegisterConfig: %w", err)
	}

	return getterFunc[T](func() T {
		cfg := mgr.GetConfig(key).(*T) // config is always present upon success registry, no need to check for nil
		return *cfg
	}), nil
}
