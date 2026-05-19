package config

import (
	"context"
	"fmt"
	"maps"
	"reflect"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
)

// DynamicConfig holds configuration for a Dynamic instance.
type DynamicConfig struct {
	FetchInterval time.Duration
	FetchTimeout  time.Duration
	// ErrCallback will be called (if any) in case there is any error in the background process.
	// It should not block for too long.
	ErrCallback func(err error)
}

type DynamicConfigOption func(*DynamicConfig)

// WithDynamicFetchInterval sets how often the provider is polled for config changes.
// Defaults to 10s.
func WithDynamicFetchInterval(fetch time.Duration) DynamicConfigOption {
	return func(dc *DynamicConfig) {
		dc.FetchInterval = fetch
	}
}

// WithDynamicFetchTimeout sets the timeout for each provider fetch call.
// Defaults to 5s.
func WithDynamicFetchTimeout(timeout time.Duration) DynamicConfigOption {
	return func(dc *DynamicConfig) {
		dc.FetchTimeout = timeout
	}
}

// WithErrCallback registers a callback that is called whenever a background error occurs
// (e.g. fetch failure, parse failure). The callback must not block for too long.
func WithErrCallback(cb func(error)) DynamicConfigOption {
	return func(dc *DynamicConfig) {
		dc.ErrCallback = cb
	}
}

type registrant[T any] struct {
	factory    func() T
	currentCfg T
	callbacks  []func(any)
}

// Dynamic periodically fetches config from a Provider and keeps registered configs up to date.
// Callers register a key+factory via RegisterConfig or RegisterConfigWithNotify, then read the
// latest parsed value via GetConfig at any time.
//
// Deregistration is not supported as of now; registered configs are expected to live for the
// lifetime of the Dynamic instance.
type Dynamic struct {
	mu sync.RWMutex

	cfg DynamicConfig

	// key is T{}, value is current config
	configRegistry *xsync.MapOf[any, registrant[any]]
	currentCfg     map[string]string

	provider Provider

	closeOnce sync.Once
	closeCh   chan struct{}
}

// NewDynamic creates a Dynamic and performs an initial config fetch from the provider.
// Returns an error if the initial fetch fails.
func NewDynamic(provider Provider, opts ...DynamicConfigOption) (*Dynamic, error) {
	ctx := context.TODO()

	opt := DynamicConfig{
		FetchInterval: 10 * time.Second,
		FetchTimeout:  5 * time.Second,
	}
	for _, optFn := range opts {
		optFn(&opt)
	}

	currentCfg, err := provider.Config(ctx)
	if err != nil {
		return nil, fmt.Errorf("provider.Config: %w", err)
	}

	d := &Dynamic{
		configRegistry: xsync.NewMapOf[any, registrant[any]](),
		provider:       provider,
		closeCh:        make(chan struct{}),
		cfg:            opt,
		currentCfg:     currentCfg,
	}

	return d, nil
}

func (d *Dynamic) fetchConfigPeriodically() {
	ticker := time.NewTicker(d.cfg.FetchInterval)

	for {
		select {
		case <-d.closeCh:
			return
		case <-ticker.C:
		}

		d.updateConfig()
	}
}

func (d *Dynamic) updateConfig() {
	ctx, cancel := context.WithTimeout(context.TODO(), d.cfg.FetchTimeout)
	defer cancel()

	cfgMap, err := d.provider.FetchConfig(ctx)
	if err != nil {
		if d.cfg.ErrCallback != nil {
			d.cfg.ErrCallback(fmt.Errorf("updateConfig: d.provider.FetchConfig: %w", err))
		}
		return
	}

	d.mu.Lock()
	if maps.Equal(d.currentCfg, cfgMap) {
		d.mu.Unlock()
		return
	}

	d.currentCfg = cfgMap
	d.mu.Unlock()

	d.configRegistry.Range(func(key any, value registrant[any]) bool {
		dst := value.factory()

		err := loadConfigFromMapTo(context.TODO(), dst, cfgMap)
		if err != nil {
			if d.cfg.ErrCallback != nil {
				d.cfg.ErrCallback(fmt.Errorf("updateConfig: loadConfigFromMapTo: %w", err))
			}
			return true
		}

		old := value.currentCfg
		value.currentCfg = dst

		d.configRegistry.Store(key, value)

		if !reflect.DeepEqual(old, dst) {
			for _, cb := range value.callbacks {
				cb(dst)
			}
		}

		return true
	})
}

// RegisterConfig registers a key and factory for dynamic config updates.
// Factory must return a pointer to a zero config struct used for parsing.
// The parsed config is immediately available via GetConfig.
func (d *Dynamic) RegisterConfig(key any, factory func() any) error {
	dst := factory()

	d.mu.RLock()
	currentCfg := d.currentCfg
	d.mu.RUnlock()

	err := loadConfigFromMapTo(context.TODO(), dst, currentCfg)
	if err != nil {
		return fmt.Errorf("loadConfigFromMapTo: %w", err)
	}

	d.configRegistry.Store(key, registrant[any]{
		factory:    factory,
		currentCfg: dst,
	})

	return nil
}

// RegisterConfigWithNotify is like RegisterConfig but also returns a callback adder.
// The adder can be called multiple times to register callbacks that are fired whenever
// the config changes. Callbacks are called synchronously during the update cycle, so
// they must not block for too long.
func (d *Dynamic) RegisterConfigWithNotify(key any, factory func() any) (CallbackAdder, error) {
	dst := factory()

	d.mu.RLock()
	currentCfg := d.currentCfg
	d.mu.RUnlock()

	err := loadConfigFromMapTo(context.TODO(), dst, currentCfg)
	if err != nil {
		return nil, fmt.Errorf("loadConfigFromMapTo: %w", err)
	}

	d.configRegistry.Store(key, registrant[any]{
		factory:    factory,
		currentCfg: dst,
	})

	adder := callbackAdderFunc(func(cb func(any)) {
		d.configRegistry.Compute(key, func(oldValue registrant[any], loaded bool) (newValue registrant[any], delete bool) {
			if !loaded {
				return oldValue, true
			}

			oldValue.callbacks = append(oldValue.callbacks, cb)
			return oldValue, false
		})
	})

	return adder, nil
}

// GetConfig returns the latest parsed config for the given key, or nil if not registered.
func (d *Dynamic) GetConfig(key any) any {
	cfg, ok := d.configRegistry.Load(key)
	if !ok {
		return nil
	}
	return cfg.currentCfg
}

// Start begins the background polling loop. Must be called once after NewDynamic.
func (d *Dynamic) Start() {
	go d.fetchConfigPeriodically()
}

// Close stops the background polling loop. Safe to call multiple times.
func (d *Dynamic) Close() {
	d.closeOnce.Do(func() {
		close(d.closeCh)
	})
}
