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

// credentialProvider implements [auther].
func (u *universalAuth) credentialProvider(auth infisical.AuthInterface) (infisical.MachineIdentityCredential, error) {
	return auth.UniversalAuthLogin(u.clientID, u.clientSecret)
}

type k8sAuth struct {
	identityId string
	tokenPath  string
}

var _ authenticator = (*k8sAuth)(nil)

// credentialProvider implements [auther].
func (k *k8sAuth) credentialProvider(auth infisical.AuthInterface) (infisical.MachineIdentityCredential, error) {
	return auth.KubernetesAuthLogin(k.identityId, k.tokenPath)
}
