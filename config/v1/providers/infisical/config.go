package infisical

import infisical "github.com/infisical/go-sdk"

type Config struct {
	SiteUrl      string
	ClientID     string
	ClientSecret string
	ProjectSlug  string
	Environment  string
	ConfigPath   string

	RetryConfig *infisical.RetryRequestsConfig
}

type SecretConfig struct {
	ProjectSlug string
	Environment string
	ConfigPath  string
}
