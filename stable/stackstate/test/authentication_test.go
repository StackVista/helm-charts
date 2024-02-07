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
      password = "foobar"
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

func TestAuthenticationLdapSplit(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-api", []string{"values/ldap_authentication.yaml"}, expectedLdapAuthConfig, expectedLdapAuthEnabled)
}

func TestAuthenticationLdap(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-server", []string{"values/ldap_authentication.yaml", "values/split_disabled.yaml"}, expectedLdapAuthConfig, expectedLdapAuthEnabled)
}

func TestAuthenticationLdapMissingValues(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/ldap_authentication_missing_userQuery_parameters.yaml")
	require.Contains(t, err.Error(), "userQuery parameters")
	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/ldap_authentication_missing_userQuery_usernameKey.yaml")
	require.Contains(t, err.Error(), "userQuery usernameKey")
}

const expectedOidcAuthConfig = `stackstate.api.authentication.authServer.oidcAuthServer {
  clientId = "stackstate-client-id"
  secret = "some-secret"
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

const expectedOidcEnabled = `stackstate.api.authentication.authServer.authServerType = [oidcAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationOidcSplit(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-api", []string{"values/oidc_authentication.yaml"}, expectedOidcAuthConfig, expectedOidcEnabled)
}

func TestAuthenticationOidc(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-server", []string{"values/oidc_authentication.yaml", "values/split_disabled.yaml"}, expectedOidcAuthConfig, expectedOidcEnabled)
}

func TestAuthenticationOidcInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/oidc_authentication_invalid_scope.yaml")
	require.Contains(t, err.Error(), "stackstate.authentication.oidc.scope must be an array of scopes")

	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/oidc_authentication_missing_clientid.yaml")
	require.Contains(t, err.Error(), "the client id to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/oidc_authentication_missing_secret.yaml")
	require.Contains(t, err.Error(), "the client secret to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/oidc_authentication_missing_discoveryuri.yaml")
	require.Contains(t, err.Error(), "the discovery uri to be set")
}

const expectedKeycloakAuthConfig = `stackstate.api.authentication.authServer.keycloakAuthServer {
  keycloakBaseUri = "http://keycloak"
  realm = "test"
  clientId = "stackstate-client-id"
  secret = "some-secret"
  redirectUri = "http://localhost/loginCallback"
  authenticationMethod = "client_secret_basic"
  jwsAlgorithm = "RS256"
  jwtClaims {
    groupsField = "groups"
  }
}`

const expectedKeycloakEnabled = `stackstate.api.authentication.authServer.authServerType = [keycloakAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationKeycloakSplit(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-api", []string{"values/keycloak_authentication.yaml"}, expectedKeycloakAuthConfig, expectedKeycloakEnabled)
}

func TestAuthenticationKeycloak(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-server", []string{"values/keycloak_authentication.yaml", "values/split_disabled.yaml"}, expectedKeycloakAuthConfig, expectedKeycloakEnabled)
}

func TestAuthenticationKeycloakInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/keycloak_authentication_missing_clientid.yaml")
	require.Contains(t, err.Error(), "the client id to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/keycloak_authentication_missing_secret.yaml")
	require.Contains(t, err.Error(), "the client secret to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/keycloak_authentication_missing_realm.yaml")
	require.Contains(t, err.Error(), "the keycloak realm to be set")
	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/keycloak_authentication_missing_url.yaml")
	require.Contains(t, err.Error(), "the keycloak url to be set")
}

const expectedFileAuthConfig = `stackstate.api.authentication.authServer.stackstateAuthServer.logins = [
  { username = "administrator", password = "098f6bcd4621d373cade4e832627b4f6", roles = ["stackstate-admin"] },
  { username = "platformadmin", password = "098f6bcd4621d373cade4e832627b4f6", roles = ["stackstate-platform-admin"] },
  { username = "guest1", password = "098f6bcd4621d373cade4e832627b4f6", roles = ["stackstate-guest"] },
  { username = "guest2", password = "098f6bcd4621d373cade4e832627b4f6", roles = ["stackstate-guest"] },
  { username = "maintainer", password = "098f6bcd4621d373cade4e832627b4f6", roles = ["stackstate-power-user","stackstate-guest"] },
]`

const expectedFileEnabled = `stackstate.api.authentication.authServer.authServerType = [stackstateAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationFileSplit(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-api", []string{"values/file_authentication.yaml"}, expectedFileAuthConfig, expectedFileEnabled)
}

func TestAuthenticationFile(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-server", []string{"values/file_authentication.yaml", "values/split_disabled.yaml"}, expectedFileAuthConfig, expectedFileEnabled)
}

func TestAuthenticationFileInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/file_authentication_no_logins.yaml")
	require.Contains(t, err.Error(), "requires a non-empty list of logins")

	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/file_authentication_no_roles.yaml")
	require.Contains(t, err.Error(), "No roles specified for user administrator")

	err = helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/file_authentication_no_password.yaml")
	require.Contains(t, err.Error(), "A login requires a password hash")
}

const expectedFallbackAuthConfig = `stackstate.api.authentication.authServer.stackstateAuthServer {
  defaultPassword = "098f6bcd4621d373cade4e832627b4f6"
}`

const expectedFallbackEnabled = `stackstate.api.authentication.authServer.authServerType = [stackstateAuthServer, k8sServiceAccountAuthServer]`

func TestAuthenticationFallbackSplit(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-api", []string{}, expectedFallbackAuthConfig, expectedFallbackEnabled)
}

func TestAuthenticationFallback(t *testing.T) {
	RunSecretsConfigTest(t, "stackstate-server", []string{"values/split_disabled.yaml"}, expectedFallbackAuthConfig, expectedFallbackEnabled)
}

func TestAuthenticationFallbackInvalid(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/fallback_authentication_no_password.yaml")
	require.Contains(t, err.Error(), "Helm value 'stackstate.authentication.adminPassword' is required")
}

const expectedRolesAuthConfig = `stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} ["extra-admin","stackstate-aad"]
stackstate.authorization.platformAdminGroups = ${stackstate.authorization.platformAdminGroups} ["extra-platform-admin"]
stackstate.authorization.powerUserGroups = ${stackstate.authorization.powerUserGroups} ["extra-power"]
stackstate.authorization.guestGroups = ${stackstate.authorization.guestGroups} ["guest1","guest2"]`

func TestAuthenticationRolesSplit(t *testing.T) {
	RunSecretsConfigTestF(t, "stackstate-api", []string{"values/authentication_roles.yaml"}, func(stringData string) {
		// check that the roles are added
		require.Contains(t, stringData, "stackstate-aad")
		require.Contains(t, stringData, "extra-admin")
		require.Contains(t, stringData, "extra-platform-admin")
		require.Contains(t, stringData, "extra-power")
		require.Contains(t, stringData, "guest1")
		require.Contains(t, stringData, "guest2")
	})
}

const expectedRolesWhenEmptyAuthConfig = `stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} ["stackstate-aad"]`

func TestAuthenticationRolesEmptySplit(t *testing.T) {
	RunSecretsConfigTestF(t, "stackstate-api", []string{"values/authentication_roles_empty.yaml"}, func(stringData string) {
		require.Contains(t, stringData, "stackstate-aad")
	})
}

const expectedRolesWhenUndefinedAdminAuthConfig = `stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} ["stackstate-aad"]`

func TestAuthenticationRolesUndefinedAdminSplit(t *testing.T) {
	RunSecretsConfigTestF(t, "stackstate-api", []string{"values/authentication_roles_no_admin.yaml"}, func(stringData string) {
		require.Contains(t, stringData, "stackstate-aad")
	})
}

func TestNoAuthenticationRolesSaaS(t *testing.T) {
	RunSecretsConfigTestF(t, "stackstate-server", []string{"values/authentication_saas_noroles.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the stackstate-k8s-troubleshooter role is added
		require.NotContains(t, stringData, "stackstate-aad")
	})
}

func TestIgnoredAuthenticationRolesSaaS(t *testing.T) {
	RunSecretsConfigTestF(t, "stackstate-server", []string{"values/authentication_saas_noroles.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the stackstate-k8s-troubleshooter role is added
		require.NotContains(t, stringData, "stackstate-aad")
		require.NotContains(t, stringData, "extra-admin")
		require.NotContains(t, stringData, "guest1")
		require.NotContains(t, stringData, "extra-platform-admin")
		require.NotContains(t, stringData, "extra-power-user")
	})
}

func TestCustomAuthenticationRolesSaaS(t *testing.T) {
	RunSecretsConfigTestF(t, "stackstate-server", []string{"values/authentication_saas_custom.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the stackstate-k8s-troubleshooter role is added
		require.Contains(t, stringData, "stackstate-admin")
	})
}

func TestAuthenticationRoles(t *testing.T) {
	RunSecretsConfigTestF(t, "stackstate-server", []string{"values/authentication_roles.yaml", "values/split_disabled.yaml"}, func(stringData string) {
		// check that the roles are added
		require.Contains(t, stringData, "stackstate-aad")
		require.Contains(t, stringData, "extra-admin")
		require.Contains(t, stringData, "extra-platform-admin")
		require.Contains(t, stringData, "extra-power")
		require.Contains(t, stringData, "guest1")
		require.Contains(t, stringData, "guest2")
	})
}

func TestMultipleAuthConfigsNotAllowed(t *testing.T) {
	err := helmtestutil.RenderHelmTemplateError(t, "stackstate", "values/full.yaml", "values/multiple_auth_configs.yaml")
	require.Contains(t, err.Error(), "More than 1 authentication mechanism specified")
}
