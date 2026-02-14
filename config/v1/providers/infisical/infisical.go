package infisical

import (
	"context"
	"fmt"
	"maps"

	infisical "github.com/infisical/go-sdk"
)

type ConfigProvider struct {
	secretConfig SecretConfig
	client       infisical.InfisicalClientInterface
	cancel       context.CancelFunc
	initial      map[string]string
}

func New(config Config) (*ConfigProvider, error) {
	opts := []Option{
		WithUniversalAuth(config.ClientID, config.ClientSecret),
	}

	if config.RetryConfig != nil {
		opts = append(opts, WithRetryConfig(*config.RetryConfig))
	}

	return NewWithOptions(config.SiteUrl, SecretConfig{
		ProjectSlug: config.ProjectSlug,
		Environment: config.Environment,
		ConfigPath:  config.ConfigPath,
	}, opts...)
}

func NewWithOptions(siteURL string, secretCfg SecretConfig, options ...Option) (*ConfigProvider, error) {
	ctx, cancel := context.WithCancel(context.TODO())

	opts := resolveOptions(options...)

	client := infisical.NewInfisicalClient(ctx, infisical.Config{
		SiteUrl:              siteURL,
		AutoTokenRefresh:     true,
		RetryRequestsConfig:  opts.retryConfig,
		CacheExpiryInSeconds: 0, // no cache
	})

	_, err := opts.auther.credentialProvider(client.Auth())
	if err != nil {
		cancel()
		return nil, fmt.Errorf("opts.auther.credentialProvider: %w", err)
	}

	provider := ConfigProvider{
		client:       client,
		cancel:       cancel,
		secretConfig: secretCfg,
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
		ProjectSlug: c.secretConfig.ProjectSlug,
		Environment: c.secretConfig.Environment,
		SecretPath:  c.secretConfig.ConfigPath,
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
