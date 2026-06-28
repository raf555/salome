package config

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sethvargo/go-envconfig"
)

func loadConfigFromMapTo(ctx context.Context, dst any, cfg map[string]string) error {
	if err := processConfig(ctx, dst, cfg); err != nil {
		return fmt.Errorf("processConfig: %w", err)
	}

	validatorOnce.Do(func() {
		vl = validator.New()
	})

	if err := vl.StructCtx(ctx, dst); err != nil {
		return fmt.Errorf("vl.StructCtx: %w", err)
	}

	return nil
}

func processConfig(ctx context.Context, dst any, cfg map[string]string) error {
	if err := envconfig.ProcessWith(ctx, &envconfig.Config{
		Target: dst,
		Lookuper: envconfig.MultiLookuper(
			envconfig.OsLookuper(), // if env is specified, it takes precedence
			envconfig.MapLookuper(cfg),
		),
	}); err != nil {
		return fmt.Errorf("envconfig.Process: %w", err)
	}
	return nil
}
