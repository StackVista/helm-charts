package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const expectedBootstrapTokenConfig = `stackstate.api.authentication.authServer.serviceTokenAuthServer.bootstrap {
  token = ${bootstrap_token}
  roles = [stackstate-k8s-admin]
  ttl = "24h"
}`

const expectedBootstrapTokenFallbackEnabled = `stackstate.api.authentication.authServer.authServerType = [stackstateAuthServer, serviceTokenAuthServer, k8sServiceAccountAuthServer]`

func TestBootstrapTokenSecretWithFallback(t *testing.T) {
	RunSecretTestF(t, "suse-observability-auth", []string{"values/bootstrap_token.yaml"}, func(inSecret map[string]string) {
		assert.Contains(t, inSecret, "default_password")
		assert.Equal(t, "ea8ec1f6dc01149fc0a99cc55fb494b4", inSecret["bootstrap_token"])
	})
}

func TestBootstrapTokenConfigWithFallback(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/bootstrap_token.yaml"}, expectedBootstrapTokenConfig, expectedBootstrapTokenFallbackEnabled)
}

const expectedBootstrapTokenRancherEnabled = `stackstate.api.authentication.authServer.authServerType = [oidcAuthServer, serviceTokenAuthServer, k8sServiceAccountAuthServer]`

func TestBootstrapTokenSecretWithRancher(t *testing.T) {
	RunSecretTestF(t, "suse-observability-auth", []string{"values/bootstrap_token_with_rancher.yaml"}, func(inSecret map[string]string) {
		assert.Equal(t, "stackstate-client-id", inSecret["oidc_client_id"])
		assert.Equal(t, "some-secret", inSecret["oidc_secret"])
		assert.Equal(t, "ea8ec1f6dc01149fc0a99cc55fb494b4", inSecret["bootstrap_token"])
		assert.NotContains(t, inSecret, "default_password")
	})
}

func TestBootstrapTokenConfigWithRancher(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/bootstrap_token_with_rancher.yaml"}, expectedBootstrapTokenConfig, expectedBootstrapTokenRancherEnabled)
}

const expectedBootstrapTokenOidcEnabled = `stackstate.api.authentication.authServer.authServerType = [oidcAuthServer, serviceTokenAuthServer, k8sServiceAccountAuthServer]`

func TestBootstrapTokenSecretWithOidc(t *testing.T) {
	RunSecretTestF(t, "suse-observability-auth", []string{"values/bootstrap_token_with_oidc.yaml"}, func(inSecret map[string]string) {
		assert.Equal(t, "stackstate-client-id", inSecret["oidc_client_id"])
		assert.Equal(t, "some-secret", inSecret["oidc_secret"])
		assert.Equal(t, "ea8ec1f6dc01149fc0a99cc55fb494b4", inSecret["bootstrap_token"])
		assert.NotContains(t, inSecret, "default_password")
	})
}

func TestBootstrapTokenConfigWithOidc(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/bootstrap_token_with_oidc.yaml"}, expectedBootstrapTokenConfig, expectedBootstrapTokenOidcEnabled)
}

const expectedBootstrapTokenKeycloakEnabled = `stackstate.api.authentication.authServer.authServerType = [keycloakAuthServer, serviceTokenAuthServer, k8sServiceAccountAuthServer]`

func TestBootstrapTokenSecretWithKeycloak(t *testing.T) {
	RunSecretTestF(t, "suse-observability-auth", []string{"values/bootstrap_token_with_keycloak.yaml"}, func(inSecret map[string]string) {
		assert.Equal(t, "stackstate-client-id", inSecret["keycloak_client_id"])
		assert.Equal(t, "some-secret", inSecret["keycloak_secret"])
		assert.Equal(t, "ea8ec1f6dc01149fc0a99cc55fb494b4", inSecret["bootstrap_token"])
		assert.NotContains(t, inSecret, "default_password")
	})
}

func TestBootstrapTokenConfigWithKeycloak(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/bootstrap_token_with_keycloak.yaml"}, expectedBootstrapTokenConfig, expectedBootstrapTokenKeycloakEnabled)
}

const expectedBootstrapTokenLdapEnabled = `stackstate.api.authentication.authServer.authServerType = [ldapAuthServer, serviceTokenAuthServer, k8sServiceAccountAuthServer]`

func TestBootstrapTokenSecretWithLdap(t *testing.T) {
	RunSecretTestF(t, "suse-observability-auth", []string{"values/bootstrap_token_with_ldap.yaml"}, func(inSecret map[string]string) {
		assert.Equal(t, "foobar", inSecret["ldap_password"])
		assert.Equal(t, "ea8ec1f6dc01149fc0a99cc55fb494b4", inSecret["bootstrap_token"])
		assert.NotContains(t, inSecret, "default_password")
	})
}

func TestBootstrapTokenConfigWithLdap(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/bootstrap_token_with_ldap.yaml"}, expectedBootstrapTokenConfig, expectedBootstrapTokenLdapEnabled)
}

const expectedBootstrapTokenFileEnabled = `stackstate.api.authentication.authServer.authServerType = [stackstateAuthServer, serviceTokenAuthServer, k8sServiceAccountAuthServer]`

func TestBootstrapTokenSecretWithFile(t *testing.T) {
	RunSecretTestF(t, "suse-observability-auth", []string{"values/bootstrap_token_with_file.yaml"}, func(inSecret map[string]string) {
		assert.Equal(t, "098f6bcd4621d373cade4e832627b4f6", inSecret["file_administrator_password"])
		assert.Equal(t, "ea8ec1f6dc01149fc0a99cc55fb494b4", inSecret["bootstrap_token"])
		assert.NotContains(t, inSecret, "default_password")
	})
}

func TestBootstrapTokenConfigWithFile(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/bootstrap_token_with_file.yaml"}, expectedBootstrapTokenConfig, expectedBootstrapTokenFileEnabled)
}

func TestBootstrapTokenSecretNameIsCorrect(t *testing.T) {
	output := helmtestutil.RenderHelmTemplate(t, "suse-observability", "values/full.yaml", "values/bootstrap_token_with_rancher.yaml")
	resources := helmtestutil.NewKubernetesResources(t, output)

	_, hasAuthSecret := resources.Secrets["suse-observability-auth"]
	require.True(t, hasAuthSecret, "Expected secret 'suse-observability-auth' to exist")

	_, hasBadSecret := resources.Secrets["suse-observability-suse-observability"]
	require.False(t, hasBadSecret, "Secret 'suse-observability-suse-observability' should not exist (indicates fromYaml failure)")
}
