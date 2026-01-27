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

func (d *Dynamic) swapCurrentConfig(newCfg map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.currentCfg = newCfg
}

func (d *Dynamic) readCurrentConfig() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.currentCfg
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
		// TODO: log
		return
	}

	if maps.Equal(d.readCurrentConfig(), cfgMap) {
		return
	}

	d.swapCurrentConfig(cfgMap)

	d.configRegistry.Range(func(key any, value registrant[any]) bool {
		dst := value.factory()

		err := loadConfigFromMapTo(context.TODO(), dst, cfgMap)
		if err != nil {
			// TODO: log
			return true
		}

		value.currentCfg = dst

		d.configRegistry.Store(key, value)
		return true
	})
}

func (d *Dynamic) RegisterConfig(key any, factory func() any) error {
	dst := factory()

	err := loadConfigFromMapTo(context.TODO(), dst, d.readCurrentConfig())
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
