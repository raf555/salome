package infisical

import infisical "github.com/infisical/go-sdk"

type Config struct {
	SiteUrl      string
	ClientID     string
	ClientSecret string
	ProjectSlug  string
	Environment  string

	RetryConfig *infisical.RetryRequestsConfig
}
