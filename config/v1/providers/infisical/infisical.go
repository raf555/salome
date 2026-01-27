package infisical

import (
	"context"
	"fmt"
	"maps"

	infisical "github.com/infisical/go-sdk"
)

type ConfigProvider struct {
	client  infisical.InfisicalClientInterface
	config  Config
	cancel  context.CancelFunc
	initial map[string]string
}

func New(config Config) (*ConfigProvider, error) {
	ctx, cancel := context.WithCancel(context.TODO())

	client := infisical.NewInfisicalClient(ctx, infisical.Config{
		SiteUrl:              config.SiteUrl,
		AutoTokenRefresh:     true,
		RetryRequestsConfig:  config.RetryConfig,
		CacheExpiryInSeconds: 0, // no cache
	})

	_, err := client.Auth().UniversalAuthLogin(config.ClientID, config.ClientSecret)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("client.Auth.UniversalAuthLogin: %w", err)
	}

	provider := ConfigProvider{
		client: client,
		cancel: cancel,
		config: config,
	}

	provider.initial, err = provider.FetchConfig(context.TODO())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("provider.FetchConfig: %w", err)
	}

	return &provider, nil
}

func (c *ConfigProvider) Close() {
	c.cancel()
}

func (c *ConfigProvider) Config(_ context.Context) (map[string]string, error) {
	return maps.Clone(c.initial), nil
}

func (c *ConfigProvider) FetchConfig(_ context.Context) (map[string]string, error) {
	secrets, err := c.client.Secrets().List(infisical.ListSecretsOptions{
		ProjectSlug: c.config.ProjectSlug,
		Environment: c.config.Environment,
	})
	if err != nil {
		return nil, fmt.Errorf("c.client.Secrets.List: %w", err)
	}

	out := make(map[string]string, len(secrets))
	for _, secret := range secrets {
		out[secret.SecretKey] = secret.SecretValue
	}

	return out, nil
}
