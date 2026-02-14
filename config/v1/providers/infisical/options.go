package infisical

import infisical "github.com/infisical/go-sdk"

type options struct {
	auther      authenticator
	retryConfig *infisical.RetryRequestsConfig
}

type Option func(*options)

// WithKubernetesAuth provides auth with Kubernetes SA.
// identityId and tokenPath is optional. If not provided, it will be fetched from
// INFISICAL_KUBERNETES_IDENTITY_ID and INFISICAL_KUBERNETES_SERVICE_ACCOUNT_TOKEN_PATH_ENV_NAME environment variables.
func WithKubernetesAuth(identityId, tokenPath string) Option {
	return func(o *options) {
		o.auther = &k8sAuth{
			identityId: identityId,
			tokenPath:  tokenPath,
		}
	}
}

// WithKubernetesAuth provides auth with Universal Auth.
// clientId and clientSecret is optional. If not provided, it will be fetched from
// INFISICAL_UNIVERSAL_AUTH_CLIENT_ID and INFISICAL_UNIVERSAL_AUTH_CLIENT_SECRET environment variables.
func WithUniversalAuth(clientId, clientSecret string) Option {
	return func(o *options) {
		o.auther = &universalAuth{
			clientID:     clientId,
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
		auther: &universalAuth{},
	}

	for _, opt := range opts {
		opt(defaultOpt)
	}

	return defaultOpt
}
