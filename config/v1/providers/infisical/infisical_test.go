package infisical

import (
	"testing"

	infisical "github.com/infisical/go-sdk"
	"github.com/infisical/go-sdk/packages/models"
	"github.com/stretchr/testify/assert"
)

// fakeClient implements only Secrets; any other method panics through the nil embedded interface.
type fakeClient struct {
	infisical.InfisicalClientInterface
	secrets *fakeSecrets
}

func (f *fakeClient) Secrets() infisical.SecretsInterface {
	return f.secrets
}

type fakeSecrets struct {
	infisical.SecretsInterface
	got     infisical.ListSecretsOptions
	secrets []models.Secret
}

func (f *fakeSecrets) List(options infisical.ListSecretsOptions) ([]models.Secret, error) {
	f.got = options
	return f.secrets, nil
}

func TestFetchConfigPassesSecretConfig(t *testing.T) {
	secrets := &fakeSecrets{
		secrets: []models.Secret{
			{SecretKey: "KEY", SecretValue: "value"},
		},
	}
	provider := ConfigProvider{
		client: &fakeClient{secrets: secrets},
		secretConfig: SecretConfig{
			ProjectSlug:    "opm",
			Environment:    "prod",
			ConfigPath:     "/api-backend",
			IncludeImports: true,
		},
	}

	cfg, err := provider.FetchConfig(t.Context())
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"KEY": "value"}, cfg)
	assert.Equal(t, infisical.ListSecretsOptions{
		ProjectSlug:    "opm",
		Environment:    "prod",
		SecretPath:     "/api-backend",
		IncludeImports: true,
	}, secrets.got)
}
