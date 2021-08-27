{{/*
Shared settings in configmap for server and api
*/}}
{{- define "stackstate.configmap.server-and-api" }}
{{- $authTypes := list -}}
stackstate.api.authentication.authServer.k8sServiceAccountAuthServer {}
{{- if .Values.caspr.enabled }}
stackstate {
  tenantAware = true
  tenant {
    identifier = {{ .Values.caspr.subscription.tenant }}
    name = "{{ .Values.caspr.subscription.tenantobj.name }}"
    subscription {
{{- if .Values.caspr.subscription.expirationDate }}
      expirationDate = {{ .Values.caspr.subscription.expirationDate }}
{{- end }}
      plan = {{ .Values.caspr.subscription.planobj.name }}
    }
  }
{{- if .Values.caspr.keycloak }}
  {{- $authTypes = append $authTypes "keycloakAuthServer" -}}
  api {
    authentication {
      authServer {
        keycloakAuthServer {
          keycloakBaseUri = "{{ .Values.caspr.keycloak.url }}"
          realm = "{{ .Values.caspr.keycloak.realm }}"
          clientId = "{{ .Values.caspr.keycloak.client }}"
          authenticationMethod = "client_secret_basic"
          secret = "{{ .Values.caspr.keycloak.secret }}"
          redirectUri = "https://{{ .Values.caspr.applicationInstance.host }}/loginCallback"
          jwsAlgorithm = "RS256"
        }
      }
    }
  }
{{- end }}
}
{{- else }}
{{- if .Values.stackstate.authentication.ldap }}
{{ $authTypes = append $authTypes "ldapAuthServer" }}
stackstate.api.authentication.authServer.ldapAuthServer {
  connection {
    host = {{ .Values.stackstate.authentication.ldap.host | required "LDAP configuration requires a host" | quote }}
    port = {{ .Values.stackstate.authentication.ldap.port | required "LDAP configuration requires a port" }}
{{- if .Values.stackstate.authentication.ldap.bind }}
    bindCredentials {
      dn = {{ .Values.stackstate.authentication.ldap.bind.dn | quote }}
      password = {{ .Values.stackstate.authentication.ldap.bind.password | quote }}
    }
{{- end }}
{{- if .Values.stackstate.authentication.ldap.ssl }}
    ssl {
      sslType = {{ required "stackstate.authentication.ldap.ssl.type is required when configuring LDAP SSL" .Values.stackstate.authentication.ldap.ssl.type }}
  {{- if .Values.stackstate.authentication.ldap.ssl.trustCertificates }}
      trustCertificatesPath = "/opt/docker/secrets/ldap-certificates.pem"
  {{- end }}
  {{- if .Values.stackstate.authentication.ldap.ssl.trustStore }}
      trustStorePath = "/opt/docker/secrets/ldap-cacerts"
  {{- end }}
    }
{{- end }}
  }

  userQuery {
    parameters = {{ .Values.stackstate.authentication.ldap.userQuery.parameters | required "LDAP authentication requires the userQuery parameters to be set." | toJson }}
    usernameKey = {{ .Values.stackstate.authentication.ldap.userQuery.usernameKey | required "LDAP authentication requires the userQuery usernameKey to be set." | quote }}
    {{- if .Values.stackstate.authentication.ldap.userQuery.emailKey }}
    emailKey = {{ .Values.stackstate.authentication.ldap.userQuery.emailKey | quote }}
    {{- end }}
  }
  groupQuery {
    parameters = {{ .Values.stackstate.authentication.ldap.groupQuery.parameters | required "LDAP authentication requires the groupQuery parameters to be set." | toJson }}
    rolesKey = {{ .Values.stackstate.authentication.ldap.groupQuery.rolesKey | required "LDAP authentication requires the groupQuery rolesKey to be set." | quote }}
    {{- if .Values.stackstate.authentication.ldap.groupQuery.groupMemberKey }}
    groupMemberKey = {{ .Values.stackstate.authentication.ldap.groupQuery.groupMemberKey | quote }}
    {{- end }}
  }
}
{{- end }}
{{- if .Values.stackstate.authentication.oidc }}
{{ $authTypes = append $authTypes "oidcAuthServer" }}
stackstate.api.authentication.authServer.oidcAuthServer {
  clientId = {{ .Values.stackstate.authentication.oidc.clientId | required "OIDC authentication requires the client id to be set." | quote }}
  secret = {{ .Values.stackstate.authentication.oidc.secret | required "OIDC authentication requires the client secret to be set." | quote }}
  discoveryUri = {{ .Values.stackstate.authentication.oidc.discoveryUri | required "OIDC authentication requires the discovery uri to be set." | quote }}
  {{- if .Values.stackstate.authentication.oidc.redirectUri }}
  redirectUri = {{ .Values.stackstate.authentication.oidc.redirectUri | trimSuffix "/" | quote }}
  {{- else }}
  redirectUri = "{{ .Values.stackstate.baseUrl | default .Values.stackstate.receiver.baseUrl | trimSuffix "/" | required "Could not determine redirectUri for OIDC. Please specify explicitly." }}/loginCallback"
  {{- end }}
  {{- if .Values.stackstate.authentication.oidc.scope }}
  {{- if not (kindIs "slice" .Values.stackstate.authentication.oidc.scope) -}}
    {{- fail "stackstate.authentication.oidc.scope must be an array of scopes" -}}
  {{- end }}
  scope = {{ .Values.stackstate.authentication.oidc.scope | toJson }}
  {{- end }}
  {{- if .Values.stackstate.authentication.oidc.authenticationMethod }}
  authenticationMethod = {{ .Values.stackstate.authentication.oidc.authenticationMethod | quote }}
  {{- end }}
  {{- if .Values.stackstate.authentication.oidc.jwsAlgorithm }}
  jwsAlgorithm = {{ .Values.stackstate.authentication.oidc.jwsAlgorithm | quote }}
  {{- end }}
  {{- if .Values.stackstate.authentication.oidc.jwtClaims }}
  jwtClaims {
    {{- if .Values.stackstate.authentication.oidc.jwtClaims.usernameField }}
    usernameField = {{ .Values.stackstate.authentication.oidc.jwtClaims.usernameField | quote }}
    {{- end }}
    {{- if .Values.stackstate.authentication.oidc.jwtClaims.groupsField }}
    groupsField = {{ .Values.stackstate.authentication.oidc.jwtClaims.groupsField | quote }}
    {{- end }}
  }
  {{- end }}
}
{{- end }}
{{- if .Values.stackstate.authentication.keycloak }}
{{ $authTypes = append $authTypes "keycloakAuthServer" }}
stackstate.api.authentication.authServer.keycloakAuthServer {
  keycloakBaseUri = {{ .Values.stackstate.authentication.keycloak.url | required "Keycloak authentication requires the keycloak url to be set." | quote }}
  realm = {{ .Values.stackstate.authentication.keycloak.realm | required "Keycloak authentication requires the keycloak realm to be set." | quote }}
  clientId = {{ .Values.stackstate.authentication.keycloak.clientId | required "Keycloak authentication requires the client id to be set." | quote }}
  secret = {{ .Values.stackstate.authentication.keycloak.secret | required "Keycloak authentication requires the client secret to be set." | quote }}
  {{- if .Values.stackstate.authentication.keycloak.redirectUri }}
  redirectUri = {{ .Values.stackstate.authentication.keycloak.redirectUri | trimSuffix '/' | quote }}
  {{- else }}
  redirectUri = "{{ .Values.stackstate.baseUrl | default .Values.stackstate.receiver.baseUrl | trimSuffix '/' | required "stackstate.baseUrl is a required value." }}/loginCallback"
  {{- end }}
  {{- if .Values.stackstate.authentication.keycloak.authenticationMethod }}
  authenticationMethod = {{ .Values.stackstate.authentication.keycloak.authenticationMethod | quote }}
  {{- end }}
  {{- if .Values.stackstate.authentication.keycloak.jwsAlgorithm }}
  jwsAlgorithm = {{ .Values.stackstate.authentication.keycloak.jwsAlgorithm | quote }}
  {{- end }}
  {{- if .Values.stackstate.authentication.keycloak.jwtClaims }}
  jwtClaims {
    {{- if .Values.stackstate.authentication.keycloak.jwtClaims.usernameField }}
    usernameField = {{ .Values.stackstate.authentication.keycloak.jwtClaims.usernameField | quote }}
    {{- end }}
    {{- if .Values.stackstate.authentication.keycloak.jwtClaims.groupsField }}
    groupsField = {{ .Values.stackstate.authentication.keycloak.jwtClaims.groupsField | quote }}
    {{- end }}
  }
  {{- end }}
}
{{- end }}

{{- if .Values.stackstate.authentication.file }}
{{ $authTypes = append $authTypes "stackstateAuthServer" }}
{{- if not .Values.stackstate.authentication.file.logins -}}
{{- fail "File configuration requires a non-empty list of logins to be specified with fields username, passwordHash and roles specified." -}}
{{- end }}
stackstate.api.authentication.authServer.stackstateAuthServer.logins = [
{{- range .Values.stackstate.authentication.file.logins }}
  {{- if not .roles -}}
  {{- printf "No roles specified for user %s" .username | fail -}}
  {{- end }}
  { username = {{ .username | required "A login requires a username" | quote }}, password = {{ .passwordHash | default .passwordMd5 | required "A login requires a password hash" | quote }}, roles = {{ .roles | toJson }} },
{{- end }}
]
{{- end }}

{{/*
Fallback to a standard 'admin' user with a specified password. Convenient for quick tests and configuration, but
for production this should be replaced with one of the other mechanisms.
*/}}
{{- if eq (len $authTypes) 0 }}
{{- $authTypes = append $authTypes "stackstateAuthServer" }}
stackstate.api.authentication.authServer.stackstateAuthServer {
{{- if .Values.stackstate.authentication.adminPassword }}
  defaultPassword = {{ .Values.stackstate.authentication.adminPassword | quote }}
{{- else }}
{{- if .Values.kots.enabled }}
  defaultPassword = {{ "f6325555cbe33536e95e2c938a4df887" | quote }}
{{- else }}
{{- fail "Helm value 'stackstate.authentication.adminPassword' is required when neither LDAP, OIDC, Keycloak nor file-based authentication has been configured" -}}
{{- end }}
{{- end }}
}
{{- end }}

{{- if gt (len $authTypes) 1 -}}
{{- fail "More than 1 authentication mechanism specified. Please configure only one from: keycloak, oidc or ldap. If none are configured the default admin user will be made available with the stackstate.authentication.adminPassword." -}}
{{- end }}

stackstate.api.authentication.sessionLifetime =  {{ .Values.stackstate.authentication.sessionLifetime | toJson }}

{{- if .Values.stackstate.authentication.roles.admin }}
stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} {{ append .Values.stackstate.authentication.roles.admin "stackstate-aad" | toJson }}
{{- else }}
{{/*
  - 'stackstate-aad' is required for anomaly-detection
  - '${stackstate.authorization.adminGroups}' is required to preserve default groups, e.g. stackstate-admin
*/}}
stackstate.authorization.adminGroups = ${stackstate.authorization.adminGroups} ["stackstate-aad"]
{{- end }}

{{- if .Values.stackstate.authentication.roles.platformAdmin }}
stackstate.authorization.platformAdminGroups = ${stackstate.authorization.platformAdminGroups} {{ .Values.stackstate.authentication.roles.platformAdmin | toJson }}
{{- end }}

{{- if .Values.stackstate.authentication.roles.powerUser }}
stackstate.authorization.powerUserGroups = ${stackstate.authorization.powerUserGroups} {{ .Values.stackstate.authentication.roles.powerUser | toJson }}
{{- end }}

{{- if .Values.stackstate.authentication.roles.guest }}
stackstate.authorization.guestGroups = ${stackstate.authorization.guestGroups} {{ .Values.stackstate.authentication.roles.guest | toJson }}
{{- end }}

{{- $authTypes = append $authTypes "k8sServiceAccountAuthServer" }}
stackstate.api.authentication.authServer.authServerType = [ {{- $authTypes | compact | join ", " -}} ]
{{- end }}

{{- with .Values.stackstate.stackpacks.installed }}
stackstate.stackPacks {
  {{- range . }}
  installOnStartUp += {{ .name | quote }}
  {{- end }}

  installOnStartUpConfig {
    {{- range . }}
    {{ .name }} = {{- toPrettyJson .configuration | indent 4 }}
    {{- end }}
  }
}
{{- end }}
{{- end -}}
