package infisical

import infisical "github.com/infisical/go-sdk"

type authenticator interface {
	credentialProvider(auth infisical.AuthInterface) (infisical.MachineIdentityCredential, error)
}

type universalAuth struct {
	clientID     string
	clientSecret string
}

var _ authenticator = (*universalAuth)(nil)

// credentialProvider implements [authenticator].
func (u *universalAuth) credentialProvider(auth infisical.AuthInterface) (infisical.MachineIdentityCredential, error) {
	return auth.UniversalAuthLogin(u.clientID, u.clientSecret)
}

type k8sAuth struct {
	identityID string
	tokenPath  string
}

var _ authenticator = (*k8sAuth)(nil)

// credentialProvider implements [authenticator].
func (k *k8sAuth) credentialProvider(auth infisical.AuthInterface) (infisical.MachineIdentityCredential, error) {
	return auth.KubernetesAuthLogin(k.identityID, k.tokenPath)
}
