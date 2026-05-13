package config

import "context"

//go:generate go tool mockgen -typed -source provider.go -destination provider.mock.gen.go -package config

type Provider interface {
	// Config provides a config that is initially read by the provider.
	Config(ctx context.Context) (map[string]string, error)
	// FetchConfig provides a config that is read by provider immediately upon call.
	FetchConfig(ctx context.Context) (map[string]string, error)
}

type DynamicConfigManager interface {
	// GetConfig provides a config from the provided key.
	// It may return nil if not found.
	GetConfig(key any) any
	// RegisterConfig registers a config to be dynamically updated.
	// Factory must provide a pointer to zero config struct to be used for parsing configs.
	RegisterConfig(key any, factory func() any) error
	// RegisterConfigWithNotify registers a config to be dynamically updated and returns a CallbackAdder.
	// The returned adder can be used to register callbacks that fire whenever the config changes.
	RegisterConfigWithNotify(key any, factory func() any) (CallbackAdder, error)
}

// CallbackAdder registers callbacks to be called when a config changes.
type CallbackAdder interface {
	Add(func(any))
}

type callbackAdderFunc func(func(any))

func (f callbackAdderFunc) Add(cb func(any)) {
	f(cb)
}

type DynamicConfigGetter[T any] interface {
	Get() T
}

type DynamicConfigGetterWithNotify[T any] interface {
	DynamicConfigGetter[T]

	RegisterCallback(func(T))
}

type getterFunc[T any] func() T

func (f getterFunc[T]) Get() T {
	return f()
}

type getterCallbackRegistrar[T any] struct {
	mgr   DynamicConfigManager
	key   *T
	adder CallbackAdder
}

func (g *getterCallbackRegistrar[T]) Get() T {
	cfg := g.mgr.GetConfig(g.key).(*T)
	return *cfg
}

func (g *getterCallbackRegistrar[T]) RegisterCallback(cb func(T)) {
	g.adder.Add(func(v any) {
		cb(*v.(*T))
	})
}
