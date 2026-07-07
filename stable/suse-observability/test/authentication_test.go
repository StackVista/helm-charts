package test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/StackVista/DevOps/helm-charts/helmtestutil"
)

const expectedLdapAuthConfig = `stackstate.api.authentication.authServer.ldapAuthServer {
  connection {
    host = "ldap-server"
    port = 10389
    bindCredentials {
      dn = "ou=acme,dc=com"
      password = ${ldap_password}
    }
    ssl {
      sslType = ssl
      trustStorePath = "/opt/docker/secrets/ldap-cacerts"
    }
  }

  userQuery {
    parameters = [{"ou":"employees"},{"dc":"stackstate"},{"dc":"com"}]
    usernameKey = "cn"
    emailKey = "email"
  }
  groupQuery {
    parameters = [{"ou":"groups"},{"dc":"stackstate"},{"dc":"com"}]
    rolesKey = "cn"
    groupMemberKey = "member"
  }
}`

const expectedLdapAuthEnabled = `stackstate.api.authentication.authServer.authServerType = [ldapAuthServer, k8sServiceAccountAuthServer]`

var expectedLdapSecretContent = map[string]string{"ldap_password": "foobar"}

func TestAuthenticationLdapSplit(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/ldap_authentication.yaml"}, expectedLdapAuthConfig, expectedLdapAuthEnabled)
}

func TestAuthenticationLdap(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/ldap_authentication.yaml", "values/split_disabled.yaml"}, expectedLdapAuthConfig, expectedLdapAuthEnabled)
}

func TestAuthenticationLdapSecret(t *testing.T) {
	RunSecretTest(t, "suse-observability-auth", []string{"values/ldap_authentication.yaml"}, expectedLdapSecretContent)
}

func TestAuthenticationLdapMissingValues(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/ldap_authentication_missing_userQuery_parameters.yaml")
	require.Contains(t, err.Error(), "userQuery parameters")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/ldap_authentication_missing_userQuery_usernameKey.yaml")
	require.Contains(t, err.Error(), "userQuery usernameKey")
}

const expectedOidcAuthConfig = `stackstate.api.authentication.authServer.oidcAuthServer {
  clientId = ${oidc_client_id}
  secret = ${oidc_secret}
  discoveryUri = "http://oidc-provider"
  redirectUri = "http://localhost/loginCallback"
  scope = ["groups","email"]
  authenticationMethod = "client_secret_basic"
  jwsAlgorithm = "RS256"
  jwtClaims {
    usernameField = "email"
    groupsField = "groups"
  }
}`

var expectedOidcAuthSecret = map[string]string{
	"oidc_client_id": "stackstate-client-id",
	"oidc_secret":    "some-secret",
}

const expectedOidcEnabled = `stackstate.api.authentication.authServer.authServerType = [oidcAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationOidcSplit(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/oidc_authentication.yaml"}, expectedOidcAuthConfig, expectedOidcEnabled)
}

func TestAuthenticationOidcSecret(t *testing.T) {
	RunSecretTest(t, "suse-observability-auth", []string{"values/oidc_authentication.yaml"}, expectedOidcAuthSecret)
}

func TestAuthenticationOidc(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/oidc_authentication.yaml", "values/split_disabled.yaml"}, expectedOidcAuthConfig, expectedOidcEnabled)
}

func TestAuthenticationOidcInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/oidc_authentication_invalid_scope.yaml")
	require.Contains(t, err.Error(), "stackstate.authentication.oidc.scope must be an array of scopes")

	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/oidc_authentication_missing_clientid.yaml")
	require.Contains(t, err.Error(), "the client id to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/oidc_authentication_missing_secret.yaml")
	require.Contains(t, err.Error(), "the client secret to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/oidc_authentication_missing_discoveryuri.yaml")
	require.Contains(t, err.Error(), "the discovery uri to be set")
}

const expectedOidcLogoutDisabled = `stackstate.api.authentication.oidcLogout = false`
const expectedSkipLoginPageDisabled = `stackstate.api.authentication.skipLoginPage = false`
const expectedOidcLogoutEnabled = `stackstate.api.authentication.oidcLogout = true`
const expectedSkipLoginPageEnabled = `stackstate.api.authentication.skipLoginPage = true`

// The oidcLogout and skipLoginPage options are configured inside the active OIDC provider block
// (oidc, keycloak or rancher) and map straight to the backend authentication config. They default
// to false, so unless a provider opts in, logging out of SUSE Observability does not log the user
// out at the provider and the login page is shown.
func TestAuthenticationOidcLogoutFlagsDefaultDisabled(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/oidc_authentication.yaml"}, expectedOidcLogoutDisabled, expectedSkipLoginPageDisabled)
}

// When enabled inside the oidc provider block, both options render as true in the backend config.
func TestAuthenticationOidcLogoutFlagsEnabled(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/oidc_authentication.yaml", "values/oidc_logout_flags.yaml"}, expectedOidcLogoutEnabled, expectedSkipLoginPageEnabled)
}

// When enabled inside the keycloak provider block, both options render as true in the backend config.
func TestAuthenticationKeycloakLogoutFlagsEnabled(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/keycloak_authentication.yaml", "values/keycloak_logout_flags.yaml"}, expectedOidcLogoutEnabled, expectedSkipLoginPageEnabled)
}

// When enabled inside the rancher provider block, both options render as true in the backend config.
func TestAuthenticationRancherLogoutFlagsEnabled(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/rancher_authentication.yaml", "values/rancher_logout_flags.yaml"}, expectedOidcLogoutEnabled, expectedSkipLoginPageEnabled)
}

// Guard for the rancher SaaS shape (e.g. the prime-test tenant): a rancher provider that does not
// opt in keeps both options false, so existing rancher tenants retain the default behaviour.
func TestAuthenticationRancherLogoutFlagsDefaultDisabled(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/rancher_authentication.yaml"}, expectedOidcLogoutDisabled, expectedSkipLoginPageDisabled)
}

// Regression guard: the chart must not derive these options from deployment.mode. Even in Saas
// mode they stay false unless a provider opts in; SaaS deployments enable them per provider.
func TestAuthenticationOidcLogoutFlagsSaasDefaultDisabled(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/authentication_saas_noroles.yaml"}, expectedOidcLogoutDisabled, expectedSkipLoginPageDisabled)
}

// Gating guard: the options only apply to OIDC-based providers. For a non-OIDC provider (LDAP) they
// stay false even when the legacy top-level stackstate.authentication.oidcLogout/skipLoginPage keys
// are set, proving the chart reads them from the provider block and ignores the old top-level keys.
func TestAuthenticationLogoutFlagsIgnoredForNonOidc(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/ldap_authentication.yaml", "values/authentication_logout_flags.yaml"}, expectedOidcLogoutDisabled, expectedSkipLoginPageDisabled)
}

const expectedRancherAuthConfig = `stackstate.api.authentication.authServer.oidcAuthServer {
  clientId = ${oidc_client_id}
  secret = ${oidc_secret}
  discoveryUri = "https://rancher-hostname.com/oidc/.well-known/openid-configuration"
  redirectUri = "http://localhost/loginCallback"
  scope = ["openid", "profile", "offline_access"]
  jwsAlgorithm = "RS256"
  jwtClaims {
    usernameField = "sub"
    groupsField = "groups"
  }
}`

var expectedRancherAuthSecret = map[string]string{
	"oidc_client_id": "stackstate-client-id",
	"oidc_secret":    "some-secret",
}

const expectedRancherAuthEnabled = `stackstate.api.authentication.authServer.authServerType = [oidcAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationRancherSplit(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/rancher_authentication.yaml"}, expectedRancherAuthConfig, expectedRancherAuthEnabled)
}

func TestAuthenticationRancherSecret(t *testing.T) {
	RunSecretTest(t, "suse-observability-auth", []string{"values/rancher_authentication.yaml"}, expectedRancherAuthSecret)
}

func TestAuthenticationRancher(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/rancher_authentication.yaml", "values/split_disabled.yaml"}, expectedRancherAuthConfig, expectedRancherAuthEnabled)
}

func TestAuthenticationRancherInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/rancher_authentication_multiple_oidc_providers.yaml")
	require.Contains(t, err.Error(), "Cannot configure both stackstate.authentication.oidc and stackstate.authentication.rancher simultaneously")

	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/rancher_authentication_missing_clientId.yaml")
	require.Contains(t, err.Error(), "the client id to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/rancher_authentication_missing_secret.yaml")
	require.Contains(t, err.Error(), "the client secret to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/rancher_authentication_missing_baseUrl.yaml")
	require.Contains(t, err.Error(), "either discoveryUri or baseUrl must be provided")
}

const expectedKeycloakAuthConfig = `stackstate.api.authentication.authServer.keycloakAuthServer {
  keycloakBaseUri = "http://keycloak"
  realm = "test"
  clientId = ${keycloak_client_id}
  secret = ${keycloak_secret}
  redirectUri = "http://localhost/loginCallback"
  scope = ["openid","profile","email"]
  authenticationMethod = "client_secret_basic"
  jwsAlgorithm = "RS256"
  jwtClaims {
    groupsField = "groups"
  }
}`

var expectedKeycloakAuthSecret = map[string]string{
	"keycloak_client_id": "stackstate-client-id",
	"keycloak_secret":    "some-secret",
}

const expectedKeycloakEnabled = `stackstate.api.authentication.authServer.authServerType = [keycloakAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationKeycloakSplit(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/keycloak_authentication.yaml"}, expectedKeycloakAuthConfig, expectedKeycloakEnabled)
}

func TestAuthenticationKeycloakSecret(t *testing.T) {
	RunSecretTest(t, "suse-observability-auth", []string{"values/keycloak_authentication.yaml"}, expectedKeycloakAuthSecret)
}

func TestAuthenticationKeycloak(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/keycloak_authentication.yaml", "values/split_disabled.yaml"}, expectedKeycloakAuthConfig, expectedKeycloakEnabled)
}

func TestAuthenticationKeycloakInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/keycloak_authentication_missing_clientid.yaml")
	require.Contains(t, err.Error(), "the client id to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/keycloak_authentication_missing_secret.yaml")
	require.Contains(t, err.Error(), "the client secret to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/keycloak_authentication_missing_realm.yaml")
	require.Contains(t, err.Error(), "the keycloak realm to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/keycloak_authentication_missing_url.yaml")
	require.Contains(t, err.Error(), "the keycloak url to be set")
}

const expectedFileAuthConfig = `stackstate.api.authentication.authServer.stackstateAuthServer.logins = [
  { username = "administrator",
    password = ${file_administrator_password},
    roles = ["stackstate-admin"] },
  { username = "guest1",
    password = ${file_guest1_password},
    roles = ["stackstate-guest"] },
  { username = "guest2",
    password = ${file_guest2_password},
    roles = ["stackstate-guest"] },
  { username = "maintainer",
    password = ${file_maintainer_password},
    roles = ["stackstate-power-user","stackstate-guest"] },
]`

var expectedFileAuthSecret = map[string]string{
	"file_administrator_password": "098f6bcd4621d373cade4e832627b4f6",
	"file_guest1_password":        "098f6bcd4621d373cade4e832627b4f6",
	"file_guest2_password":        "098f6bcd4621d373cade4e832627b4f6",
	"file_maintainer_password":    "098f6bcd4621d373cade4e832627b4f6",
}

const expectedFileEnabled = `stackstate.api.authentication.authServer.authServerType = [stackstateAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationFileSplit(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{"values/file_authentication.yaml"}, expectedFileAuthConfig, expectedFileEnabled)
}

func TestAuthenticationFileSecret(t *testing.T) {
	RunSecretTest(t, "suse-observability-auth", []string{"values/file_authentication.yaml"}, expectedFileAuthSecret)
}

func TestAuthenticationFile(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/file_authentication.yaml", "values/split_disabled.yaml"}, expectedFileAuthConfig, expectedFileEnabled)
}

func TestAuthenticationFileInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/file_authentication_no_logins.yaml")
	require.Contains(t, err.Error(), "requires a non-empty list of logins")

	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/file_authentication_no_roles.yaml")
	require.Contains(t, err.Error(), "No roles specified for user administrator")

	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/file_authentication_no_password.yaml")
	require.Contains(t, err.Error(), "A login requires a password hash")

	err = helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/file_authentication_invalid.yaml")
	require.Contains(t, err.Error(), "Only alphanumeric and _ are allowed for user names: administrator@mycompany.com.")

}

const expectedFallbackAuthConfig = `stackstate.api.authentication.authServer.stackstateAuthServer {
  defaultPassword = ${default_password}
}`

var expectedFallbackAuthSecret = map[string]string{
	"default_password": "098f6bcd4621d373cade4e832627b4f6",
}

const expectedFallbackEnabled = `stackstate.api.authentication.authServer.authServerType = [stackstateAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationFallbackSplit(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-api", []string{}, expectedFallbackAuthConfig, expectedFallbackEnabled)
}

func TestAuthenticationFallbackSecret(t *testing.T) {
	RunSecretTest(t, "suse-observability-auth", []string{}, expectedFallbackAuthSecret)
}

func TestAuthenticationFallback(t *testing.T) {
	RunConfigMapTest(t, "suse-observability-server", []string{"values/split_disabled.yaml"}, expectedFallbackAuthConfig, expectedFallbackEnabled)
}

func TestAuthenticationFallbackInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/fallback_authentication_no_password.yaml")
	require.Contains(t, err.Error(), "Admin password is required")
}

const expectedRolesAuthConfig = `stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} ["extra-admin","stackstate-aad"]
stackstate.authorization.powerUserGroups = ${stackstate.authorization.powerUserGroups} ["extra-power"]
stackstate.authorization.guestGroups = ${stackstate.authorization.guestGroups} ["guest1","guest2"]`

func TestAuthenticationRolesSplit(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-api", []string{"values/authentication_roles.yaml"}, func(stringData string) {
		// check that the roles are added
		require.NotContains(t, stringData, "stackstate-aad")
		require.Contains(t, stringData, "extra-admin")
		require.Contains(t, stringData, "extra-power")
		require.Contains(t, stringData, "guest1")
		require.Contains(t, stringData, "guest2")
	})
}

const expectedRolesWhenEmptyAuthConfig = `stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} ["stackstate-aad"]`

func TestAuthenticationRolesEmptySplit(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-api", []string{"values/authentication_roles_empty.yaml"}, func(stringData string) {
		require.NotContains(t, stringData, "stackstate-aad")
	})
}

const expectedRolesWhenUndefinedAdminAuthConfig = `stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} ["stackstate-aad"]`

func TestAuthenticationRolesUndefinedAdminSplit(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-api", []string{"values/authentication_roles_no_admin.yaml"}, func(stringData string) {
		require.NotContains(t, stringData, "stackstate-aad")
	})
}

func TestNoAuthenticationRolesSaas(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-server", []string{"values/authentication_saas_noroles.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the suse-observability-troubleshooter role is added
		require.Contains(t, stringData, "stackstate-k8s-troubleshooter")
		require.NotContains(t, stringData, "stackstate-aad")
	})
}

func TestIgnoredAuthenticationRolesSaas(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-server", []string{"values/authentication_saas_noroles.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the suse-observability-troubleshooter role is added
		require.Contains(t, stringData, "stackstate-k8s-troubleshooter")
		require.NotContains(t, stringData, "stackstate-aad")
		require.NotContains(t, stringData, "extra-admin")
		require.NotContains(t, stringData, "guest1")
		require.NotContains(t, stringData, "extra-power-user")
	})
}

func TestCustomAuthenticationRolesSaas(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-server", []string{"values/authentication_saas_custom.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the suse-observability-troubleshooter role is added
		require.Contains(t, stringData, "stackstate-k8s-troubleshooter")
		require.Contains(t, stringData, "stackstate-k8s-admin")
	})
}

func TestAuthenticationRoles(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-server", []string{"values/authentication_roles.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the roles are added
		require.NotContains(t, stringData, "stackstate-aad")
		require.Contains(t, stringData, "extra-admin")
		require.Contains(t, stringData, "extra-power")
		require.Contains(t, stringData, "guest1")
		require.Contains(t, stringData, "guest2")
	})
}

func TestAuthenticationRolesWithDots(t *testing.T) {
	RunConfigMapTestF(t, "suse-observability-server", []string{"values/authentication_roles_dots.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		require.Contains(t, stringData, "stackstate.authorization.staticSubjects.\"extra.admin\"")
		require.Contains(t, stringData, "stackstate.authorization.staticSubjects.\"extra.power\"")
		require.Contains(t, stringData, "stackstate.authorization.staticSubjects.\"guest.1\"")
		require.Contains(t, stringData, "stackstate.authorization.staticSubjects.\"guest.2\"")
		require.Contains(t, stringData, "stackstate.authorization.staticSubjects.\"one.two.three\"")
	})
}

func TestMultipleAuthConfigsNotAllowed(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "suse-observability", "values/full.yaml", "values/multiple_auth_configs.yaml")
	require.Contains(t, err.Error(), "More than 1 authentication mechanism specified")
}
