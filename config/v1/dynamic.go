package config

import (
	"context"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
)

type DynamicConfig struct {
	FetchInterval time.Duration
	FetchTimeout  time.Duration
	// ErrCallback will be called (if any) in case there is any error in the background process.
	// It should not block for too long.
	ErrCallback func(err error)
}

type DynamicConfigOption func(*DynamicConfig)

func WithDynamicFetchInterval(fetch time.Duration) DynamicConfigOption {
	return func(dc *DynamicConfig) {
		dc.FetchInterval = fetch
	}
}

func WithDynamicFetchTimeout(timeout time.Duration) DynamicConfigOption {
	return func(dc *DynamicConfig) {
		dc.FetchTimeout = timeout
	}
}

func WithErrCallback(cb func(error)) DynamicConfigOption {
	return func(dc *DynamicConfig) {
		dc.ErrCallback = cb
	}
}

type registrant[T any] struct {
	factory    func() T
	currentCfg T
}

type Dynamic struct {
	mu sync.RWMutex

	cfg DynamicConfig

	// key is T{}, value is current config
	configRegistry *xsync.MapOf[any, registrant[any]]
	currentCfg     map[string]string

	provider Provider

	closeCh chan struct{}
}

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
	for {
		select {
		case <-d.closeCh:
			return
		case <-time.Tick(d.cfg.FetchInterval):
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

		value.currentCfg = dst

		d.configRegistry.Store(key, value)
		return true
	})
}

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

// GetConfig returns nil if config is not present
func (d *Dynamic) GetConfig(key any) any {
	cfg, ok := d.configRegistry.Load(key)
	if !ok {
		return nil
	}
	return cfg.currentCfg
}

func (d *Dynamic) Start() {
	go d.fetchConfigPeriodically()
}

// Close panics if already closed,
func (d *Dynamic) Close() {
	close(d.closeCh)
}
