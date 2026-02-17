package infisical

import infisical "github.com/infisical/go-sdk"

type options struct {
	authenticator authenticator
	retryConfig   *infisical.RetryRequestsConfig
}

type Option func(*options)

// WithKubernetesAuth provides auth with Kubernetes SA.
// identityID and tokenPath is optional. If not provided, it will be fetched from
// INFISICAL_KUBERNETES_IDENTITY_ID and INFISICAL_KUBERNETES_SERVICE_ACCOUNT_TOKEN_PATH_ENV_NAME environment variables.
func WithKubernetesAuth(identityID, tokenPath string) Option {
	return func(o *options) {
		o.authenticator = &k8sAuth{
			identityID: identityID,
			tokenPath:  tokenPath,
		}
	}
}

// WithUniversalAuth provides auth with Universal Auth.
// clientID and clientSecret is optional. If not provided, it will be fetched from
// INFISICAL_UNIVERSAL_AUTH_CLIENT_ID and INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET environment variables.
func WithUniversalAuth(clientID, clientSecret string) Option {
	return func(o *options) {
		o.authenticator = &universalAuth{
			clientID:     clientID,
			clientSecret: clientSecret,
		}
	}
}

func WithRetryConfig(cfg infisical.RetryRequestsConfig) Option {
	return func(o *options) {
		o.retryConfig = &cfg
	}
}

func resolveOptions(opts ...Option) *options {
	defaultOpt := &options{
		authenticator: &universalAuth{},
	}

	for _, opt := range opts {
		opt(defaultOpt)
	}

	return defaultOpt
}
