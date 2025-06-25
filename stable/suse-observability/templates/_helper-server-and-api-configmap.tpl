{{/*
Shared settings in configmap for server and api
*/}}
{{- define "stackstate.configmap.server-and-api" }}
{{ $files := .Files }}

{{- if and .Values.stackstate.authentication (eq .Values.stackstate.deployment.mode "SelfHosted") }}
stackstate.authorization.staticSubjects.stackstate-admin: {{- $files.Get "sts-authz-permissions/stackstate-admin.json" }}
stackstate.authorization.staticSubjects.stackstate-power-user: {{- $files.Get "sts-authz-permissions/stackstate-power-user.json"}}
stackstate.authorization.staticSubjects.stackstate-guest: {{- $files.Get "sts-authz-permissions/stackstate-guest.json"}}
stackstate.authorization.staticSubjects.stackstate-k8s-troubleshooter: {{- $files.Get "sts-authz-permissions/stackstate-k8s-troubleshooter.json"}}
stackstate.authorization.staticSubjects.stackstate-ingest-telemetry: { systemPermissions: ["update-metrics"] }
stackstate.authorization.staticSubjects.suse-observability-ingest-all: { systemPermissions: ["update-permissions", "update-scoped-permissions", "update-metrics"] }

{{- if .Values.stackstate.k8sAuthorization.enabled }}
stackstate.authorization.staticSubjects.{{ template "stackstate.rbacAgent.roleName" . }}: { systemPermissions: ["update-permissions"] }
{{- end }}

{{ println "" }}
{{/* In SelfHosted mode, append any roles to the stackstate.authorization block, so that we keep the defaults delivered with stackstate. */}}
{{- range .Values.stackstate.authentication.roles.admin }}
stackstate.authorization.staticSubjects.{{ . | quote }}: {{- $files.Get "sts-authz-permissions/stackstate-admin.json" }}
{{- end }}

{{- range .Values.stackstate.authentication.roles.powerUser }}
stackstate.authorization.staticSubjects.{{ . | quote }}: {{- $files.Get "sts-authz-permissions/stackstate-power-user.json" }}
{{- end }}

{{- range .Values.stackstate.authentication.roles.guest }}
stackstate.authorization.staticSubjects.{{ . | quote }}: {{- $files.Get "sts-authz-permissions/stackstate-guest.json" }}
{{- end }}

{{- range .Values.stackstate.authentication.roles.k8sTroubleshooter }}
stackstate.authorization.staticSubjects.{{ . | quote }}: {{- $files.Get "sts-authz-permissions/stackstate-k8s-troubleshooter.json" }}
{{- end }}

{{- if index .Values "anomaly-detection" "enabled" }}
stackstate.authorization.staticSubjects.stackstate-aad: { systemPermissions: ["get-topology", "execute-monitors", "get-monitors", "get-metrics", "get-settings", "update-metrics"] }
{{- end }}

{{- else }}
{{/* In SaaS mode, the stackstate.authorization block will be ignored and we will overwrite the reference to it from the stackstate.api.authorization */}}
stackstate.api.authorization: {}
stackstate.api.authorization.staticSubjects.stackstate-k8s-troubleshooter: {{- $files.Get "sts-authz-permissions/stackstate-k8s-troubleshooter.json" }}
stackstate.api.authorization.staticSubjects.stackstate-k8s-admin: {{- $files.Get "sts-authz-permissions/stackstate-k8s-admin.json" }}
stackstate.api.authorization.staticSubjects.stackstate-ingest-telemetry: { systemPermissions: ["update-metrics"] }
stackstate.authorization.staticSubjects.suse-observability-ingest-all: { systemPermissions: ["update-permissions", "update-scoped-permissions", "update-metrics"] }
{{- if .Values.stackstate.k8sAuthorization.enabled }}
stackstate.api.authorization.staticSubjects.{{ template "stackstate.rbacAgent.roleName" . }}: { systemPermissions: ["update-permissions"] }
{{- end }}

{{- if index .Values "anomaly-detection" "enabled" }}
stackstate.api.authorization.staticSubjects.stackstate-aad: { systemPermissions: ["get-topology", "execute-monitors", "get-monitors", "get-metrics", "get-settings"] }
{{- end }}
{{- end }}

stackstate.stackPacks {
  {{- if eq .Values.hbase.deployment.mode "Distributed" }}
  localStackPacksUri = "hdfs://{{ .Release.Name }}-hbase-hdfs-nn-headful:9000/stackpacks"
  {{- else }}
  localStackPacksUri = "file:///var/stackpacks_local"
  {{- end }}

  {{- if eq .Values.stackstate.stackpacks.source "docker-image" }}
  latestVersionsStackPackStoreUri = "file:///var/stackpacks"
  {{- else if eq .Values.stackstate.stackpacks.source "s3-bucket"}}
  latestVersionsStackPackStoreUri = "s3://{{ .Values.stackstate.stackpacks.s3.bucket }}"
  {{- else }}
  {{- fail "The value for stackstate.stackpacks.source must be one of `docker-image` or `s3-bucket`" -}}
  {{- end }}

  updateStackPacksInterval = {{ .Values.stackstate.stackpacks.updateInterval | quote }}

{{- with .Values.stackstate.stackpacks.installed }}

  {{- range . }}
  installOnStartUp += {{ .name | quote }}
  {{- end }}

  installOnStartUpConfig {
    {{- range . }}
    {{ .name }} = {{- toPrettyJson .configuration | indent 4 }}
    {{- end }}
  }
{{- end }}

  upgradeOnStartUp = {{ toJson .Values.stackstate.stackpacks.upgradeOnStartup }}

  {{- $editionStackPack := printf "%s-kubernetes" (lower .Values.stackstate.deployment.edition) }}
  installOnStartUp += {{ $editionStackPack | quote }}
  upgradeOnStartUp += {{ $editionStackPack | quote }}
}

{{- if .Values.stackstate.components.api.docslink }}
stackstate.webUIConfig.docLinkUrlPrefix = "{{- .Values.stackstate.components.api.docslink -}}"
{{- end }}

{{- if .Values.stackstate.ui.defaultTimeRange }}
stackstate.webUIConfig.defaultTimeRange = "{{- .Values.stackstate.ui.defaultTimeRange -}}"
{{- end }}

{{- if .Values.stackstate.components.api.supportMode }}
stackstate.webUIConfig.supportMode = "{{- .Values.stackstate.components.api.supportMode -}}"
{{- end }}

stackstate.deploymentMode = "{{- .Values.stackstate.deployment.mode -}}"
stackstate.edition = "{{- .Values.stackstate.deployment.edition -}}"
{{- include "stackstate.service.configmap.clickhouseconfig" . }}
{{- include "stackstate.config.email" . }}

{{/*
Authentication config
*/}}
{{- $apiAuth := .Values.stackstate.authentication -}}
{{- $authnPrefix := "stackstate.api.authentication" -}}
{{- $authzPrefix := "stackstate.authorization" -}}
{{- $global := . -}}
{{- $authTypes := list -}}
{{ $authnPrefix }}.authServer.k8sServiceAccountAuthServer {}
{{- if $apiAuth.ldap }}
{{ $authTypes = append $authTypes "ldapAuthServer" }}
{{ $authnPrefix }}.authServer.ldapAuthServer {
  connection {
    host = {{ $apiAuth.ldap.host | required "LDAP configuration requires a host" | quote }}
    port = {{ $apiAuth.ldap.port | required "LDAP configuration requires a port" }}
{{- if $apiAuth.ldap.bind }}
    bindCredentials {
      dn = {{ $apiAuth.ldap.bind.dn | quote }}
      password = ${ldap_password}
    }
{{- end }}
{{- if $apiAuth.ldap.ssl }}
    ssl {
      sslType = {{ required "stackstate.authentication.ldap.ssl.type is required when configuring LDAP SSL" $apiAuth.ldap.ssl.type }}
  {{- if or $apiAuth.ldap.ssl.trustCertificates $apiAuth.ldap.ssl.trustCertificatesBase64Encoded }}
      trustCertificatesPath = "/opt/docker/secrets/ldap-certificates.pem"
  {{- end }}
  {{- if or $apiAuth.ldap.ssl.trustStore $apiAuth.ldap.ssl.trustStoreBase64Encoded }}
      trustStorePath = "/opt/docker/secrets/ldap-cacerts"
  {{- end }}
    }
{{- end }}
  }

  userQuery {
    parameters = {{ $apiAuth.ldap.userQuery.parameters | required "LDAP authentication requires the userQuery parameters to be set." | toJson }}
    usernameKey = {{ $apiAuth.ldap.userQuery.usernameKey | required "LDAP authentication requires the userQuery usernameKey to be set." | quote }}
    {{- if $apiAuth.ldap.userQuery.emailKey }}
    emailKey = {{ $apiAuth.ldap.userQuery.emailKey | quote }}
    {{- end }}
  }
  groupQuery {
    parameters = {{ $apiAuth.ldap.groupQuery.parameters | required "LDAP authentication requires the groupQuery parameters to be set." | toJson }}
    rolesKey = {{ $apiAuth.ldap.groupQuery.rolesKey | required "LDAP authentication requires the groupQuery rolesKey to be set." | quote }}
    {{- if $apiAuth.ldap.groupQuery.groupMemberKey }}
    groupMemberKey = {{ $apiAuth.ldap.groupQuery.groupMemberKey | quote }}
    {{- end }}
  }
}
{{- end }}
{{- if $apiAuth.oidc }}
{{ $authTypes = append $authTypes "oidcAuthServer" }}
{{ $authnPrefix }}.authServer.oidcAuthServer {
  clientId = ${oidc_client_id}
  secret = ${oidc_secret}
  discoveryUri = {{ $apiAuth.oidc.discoveryUri | required "OIDC authentication requires the discovery uri to be set." | quote }}
  {{- if $apiAuth.oidc.redirectUri }}
  redirectUri = {{ $apiAuth.oidc.redirectUri | trimSuffix "/" | quote }}
  {{- else }}
  redirectUri = "{{ $global.Values.stackstate.baseUrl | default $global.Values.stackstate.receiver.baseUrl | trimSuffix "/" | required "Could not determine redirectUri for OIDC. Please specify explicitly." }}/loginCallback"
  {{- end }}
  {{- if $apiAuth.oidc.scope }}
  {{- if not (kindIs "slice" $apiAuth.oidc.scope) -}}
    {{- fail "stackstate.authentication.oidc.scope must be an array of scopes" -}}
  {{- end }}
  scope = {{ $apiAuth.oidc.scope | toJson }}
  {{- end }}
  {{- if $apiAuth.oidc.authenticationMethod }}
  authenticationMethod = {{ $apiAuth.oidc.authenticationMethod | quote }}
  {{- end }}
  {{- if $apiAuth.oidc.jwsAlgorithm }}
  jwsAlgorithm = {{ $apiAuth.oidc.jwsAlgorithm | quote }}
  {{- end }}
  {{- if $apiAuth.oidc.customParameters }}
  customParams {{ $apiAuth.oidc.customParameters | toJson }}
  {{- end }}
  {{- if $apiAuth.oidc.jwtClaims }}
  jwtClaims {
    {{- if $apiAuth.oidc.jwtClaims.usernameField }}
    usernameField = {{ $apiAuth.oidc.jwtClaims.usernameField | quote }}
    {{- end }}
    {{- if $apiAuth.oidc.jwtClaims.groupsField }}
    groupsField = {{ $apiAuth.oidc.jwtClaims.groupsField | quote }}
    {{- end }}
  }
  {{- end }}
}
{{- end }}
{{- if $apiAuth.rancher }}
  {{- if has "oidcAuthServer" $authTypes }}
  {{- fail "Cannot configure both stackstate.authentication.oidc and stackstate.authentication.rancher simultaneously. Please choose one authentication method." -}}
  {{- else }}
  {{- $authTypes = append $authTypes "oidcAuthServer" }}
{{ $authnPrefix }}.authServer.oidcAuthServer {
  clientId = ${oidc_client_id}
  secret = ${oidc_secret}
  {{- if $apiAuth.rancher.discoveryUri }}
  discoveryUri = {{ $apiAuth.rancher.discoveryUri | trimSuffix "/" | quote }}
  {{- else if $apiAuth.rancher.baseUrl }}
  discoveryUri = "{{ $apiAuth.rancher.baseUrl | trimSuffix "/" }}/oidc/.well-known/openid-configuration"
  {{- else }}
  {{- fail "Cannot configure Rancher authentication: either discoveryUri or baseUrl must be provided." }}
  {{- end }}
  {{- if $apiAuth.rancher.redirectUri }}
  redirectUri = {{ $apiAuth.rancher.redirectUri | trimSuffix "/" | quote }}
  {{- else }}
  redirectUri = "{{ $global.Values.stackstate.baseUrl | default $global.Values.stackstate.receiver.baseUrl | trimSuffix "/" | required "Cannot configure Rancher authentication: baseUrl or receiver.baseUrl must be provided to construct the redirectUri." }}/loginCallback"
  {{- end }}
  scope = ["openid", "profile", "offline_access"]
  jwsAlgorithm = "RS256"
  jwtClaims {
    usernameField = "sub"
    displayNameField = "preferred_username"
    groupsField = "groups"
  }
  {{- if $apiAuth.rancher.customParameters }}
  customParams {{ $apiAuth.rancher.customParameters | toJson }}
  {{- end }}
}
  {{- end }}
{{- end }}
{{- if $apiAuth.keycloak }}
{{ $authTypes = append $authTypes "keycloakAuthServer" }}
{{ $authnPrefix }}.authServer.keycloakAuthServer {
  keycloakBaseUri = {{ $apiAuth.keycloak.url | required "Keycloak authentication requires the keycloak url to be set." | quote }}
  realm = {{ $apiAuth.keycloak.realm | required "Keycloak authentication requires the keycloak realm to be set." | quote }}
  clientId = ${keycloak_client_id}
  secret = ${keycloak_secret}
  {{- if $apiAuth.keycloak.redirectUri }}
  redirectUri = {{ $apiAuth.keycloak.redirectUri | trimSuffix "/" | quote }}
  {{- else }}
  redirectUri = "{{ $global.Values.stackstate.baseUrl | default $global.Values.stackstate.receiver.baseUrl | trimSuffix "/" | required "stackstate.baseUrl is a required value." }}/loginCallback"
  {{- end }}
  {{- if $apiAuth.keycloak.scope }}
  {{- if not (kindIs "slice" $apiAuth.keycloak.scope) -}}
    {{- fail "stackstate.authentication.keycloak.scope must be an array of scopes" -}}
  {{- end }}
  scope = {{ $apiAuth.keycloak.scope | toJson }}
  {{- end }}
  {{- if $apiAuth.keycloak.authenticationMethod }}
  authenticationMethod = {{ $apiAuth.keycloak.authenticationMethod | quote }}
  {{- end }}
  {{- if $apiAuth.keycloak.jwsAlgorithm }}
  jwsAlgorithm = {{ $apiAuth.keycloak.jwsAlgorithm | quote }}
  {{- end }}
  {{- if $apiAuth.keycloak.jwtClaims }}
  jwtClaims {
    {{- if $apiAuth.keycloak.jwtClaims.usernameField }}
    usernameField = {{ $apiAuth.keycloak.jwtClaims.usernameField | quote }}
    {{- end }}
    {{- if $apiAuth.keycloak.jwtClaims.groupsField }}
    groupsField = {{ $apiAuth.keycloak.jwtClaims.groupsField | quote }}
    {{- end }}
  }
  {{- end }}
}
{{- end }}

{{- if $apiAuth.file }}
{{ $authTypes = append $authTypes "stackstateAuthServer" }}
{{- if not $apiAuth.file.logins -}}
{{- fail "File configuration requires a non-empty list of logins to be specified with fields username, passwordHash and roles specified." -}}
{{- end }}
{{ $authnPrefix }}.authServer.stackstateAuthServer.logins = [
{{- range $apiAuth.file.logins }}
  {{- if not .roles -}}
  {{- printf "No roles specified for user %s" .username | fail -}}
  {{- end }}
  {{- if not (regexMatch "^[A-Za-z0-9_]+$" (.username | required "A login requires a username")) }}
    {{ printf "Only alphanumeric and _ are allowed for user names: %s." .username | fail }}
  {{- end }}
  { username = {{ .username | quote }},
    password = ${file_{{ .username }}_password},
    roles = {{ .roles | toJson }} },
{{- end }}
]
{{- end }}

{{/*
Fallback to a standard 'admin' user with a specified password. Convenient for quick tests and configuration, but
for production this should be replaced with one of the other mechanisms.
*/}}
{{- if eq (len $authTypes) 0 }}
{{- $authTypes = append $authTypes "stackstateAuthServer" }}
{{ $authnPrefix }}.authServer.stackstateAuthServer {
{{- if $global.Values.stackstate.authentication.adminPassword }}
  defaultPassword = ${default_password}
{{- else }}
{{- fail "Helm value 'stackstate.authentication.adminPassword' is required when neither LDAP, OIDC, Keycloak nor file-based authentication has been configured" -}}
{{- end }}
}
{{- end }}

{{- if gt (len $authTypes) 1 -}}
{{- fail "More than 1 authentication mechanism specified. Please configure only one from: keycloak, oidc, rancher or ldap. If none are configured the default admin user will be made available with the stackstate.authentication.adminPassword." -}}
{{- end }}

{{ $authnPrefix }}.sessionLifetime =  {{ $apiAuth.sessionLifetime | default "7d" | toJson }}

{{- range $k, $v := $apiAuth.roles.custom }}
{{ $authzPrefix }}.staticSubjects.{{ $k | quote }}: { systemPermissions: {{ $v.systemPermissions | toJson }}{{ if $v.resourcePermissions }}, resourcePermissions: {{ $v.resourcePermissions | toJson }}{{end}}{{ if $v.viewPermissions }}, viewPermissions: {{ $v.viewPermissions | toJson }}{{end}}{{ if $v.topologyScope }}, query: {{ $v.topologyScope | quote }}{{end}} }
{{- end }}

{{- if $apiAuth.serviceToken.bootstrap.token }}
{{- $authTypes = append $authTypes "serviceTokenAuthServer" }}
{{ $authnPrefix }}.authServer.serviceTokenAuthServer.bootstrap {
  token = ${bootstrap_token}
  roles = [ {{- $apiAuth.serviceToken.bootstrap.roles | compact | join ", " -}} ]
  ttl = {{ $apiAuth.serviceToken.bootstrap.ttl | quote }}
{{- if $apiAuth.serviceToken.bootstrap.dedicatedSubject }}
  dedicatedSubject = {{ $apiAuth.serviceToken.bootstrap.dedicatedSubject | quote }}
{{- end }}
}
{{- end }}

{{- $authTypes = append $authTypes "k8sServiceAccountAuthServer" }}
{{ $authnPrefix }}.authServer.authServerType = [ {{- $authTypes | compact | join ", " -}} ]

{{- end -}}
