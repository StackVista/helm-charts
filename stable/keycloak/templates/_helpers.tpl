{{/*
Return the proper Keycloak image name
*/}}
{{- define "keycloak.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "keycloak.imagePullSecrets" -}}
{{- include "common.images.renderPullSecrets" (dict "images" (list .Values.image) "context" $) -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "keycloak.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "common.names.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Return the secret containing the Keycloak admin bootstrap password
*/}}
{{- define "keycloak.secretName" -}}
{{- $secretName := .Values.auth.existingSecret -}}
{{- if $secretName -}}
    {{- printf "%s" (tpl $secretName $) -}}
{{- else -}}
    {{- printf "%s" (include "common.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Return the secret key that contains the Keycloak admin bootstrap password
*/}}
{{- define "keycloak.secretKey" -}}
{{- $secretName := .Values.auth.existingSecret -}}
{{- if and $secretName .Values.auth.passwordSecretKey -}}
    {{- printf "%s" .Values.auth.passwordSecretKey -}}
{{- else -}}
    {{- print "admin-password" -}}
{{- end -}}
{{- end -}}

{{/*
Return the external database host
*/}}
{{- define "keycloak.databaseHost" -}}
{{- tpl .Values.externalDatabase.host $ -}}
{{- end -}}

{{/*
Return the external database port
*/}}
{{- define "keycloak.databasePort" -}}
{{- .Values.externalDatabase.port | quote -}}
{{- end -}}

{{/*
Return the external database name
*/}}
{{- define "keycloak.databaseName" -}}
{{- .Values.externalDatabase.database -}}
{{- end -}}

{{/*
Return the external database user
*/}}
{{- define "keycloak.databaseUser" -}}
{{- .Values.externalDatabase.user -}}
{{- end -}}

{{/*
Return the name of the secret holding the external database password
*/}}
{{- define "keycloak.databaseSecretName" -}}
{{- default (printf "%s-externaldb" .Release.Name) (tpl .Values.externalDatabase.existingSecret $) -}}
{{- end -}}

{{/*
Return the secret key holding the external database password
*/}}
{{- define "keycloak.databaseSecretPasswordKey" -}}
{{- default "db-password" .Values.externalDatabase.existingSecretPasswordKey -}}
{{- end -}}

{{- define "keycloak.databaseSecretHostKey" -}}
{{- default "db-host" .Values.externalDatabase.existingSecretHostKey -}}
{{- end -}}
{{- define "keycloak.databaseSecretPortKey" -}}
{{- default "db-port" .Values.externalDatabase.existingSecretPortKey -}}
{{- end -}}
{{- define "keycloak.databaseSecretUserKey" -}}
{{- default "db-user" .Values.externalDatabase.existingSecretUserKey -}}
{{- end -}}
{{- define "keycloak.databaseSecretDatabaseKey" -}}
{{- default "db-database" .Values.externalDatabase.existingSecretDatabaseKey -}}
{{- end -}}

{{/*
Return the name of the ConfigMap holding the custom Infinispan cache configuration
*/}}
{{- define "keycloak.cacheConfigmapName" -}}
{{- if .Values.cache.existingConfigmap -}}
    {{- printf "%s" (tpl .Values.cache.existingConfigmap $) -}}
{{- else -}}
    {{- printf "%s-cache" (include "common.names.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Return true if a cache ConfigMap object should be created
*/}}
{{- define "keycloak.createCacheConfigmap" -}}
{{- if and .Values.cache.enabled (not .Values.cache.existingConfigmap) -}}
    {{- true -}}
{{- end -}}
{{- end -}}

{{/*
Return the JGroups DNS query pointing at the headless service (used by the Kubernetes/DNS_PING stack)
*/}}
{{- define "keycloak.jgroupsDnsQuery" -}}
{{- printf "%s-headless.%s.svc.%s" (include "common.names.fullname" .) (include "common.names.namespace" .) .Values.clusterDomain -}}
{{- end -}}

{{/*
Compile all warnings into a single message.
*/}}
{{- define "keycloak.validateValues" -}}
{{- $messages := list -}}
{{- $messages := append $messages (include "keycloak.validateValues.database" .) -}}
{{- $messages := without $messages "" -}}
{{- $message := join "\n" $messages -}}

{{- if $message -}}
{{-   printf "\nVALUES VALIDATION:\n%s" $message | fail -}}
{{- end -}}
{{- end -}}

{{/* Validate values of Keycloak - database */}}
{{- define "keycloak.validateValues.database" -}}
{{- if and (not .Values.externalDatabase.host) (not .Values.externalDatabase.existingSecretHostKey) -}}
keycloak: externalDatabase.host
    An external PostgreSQL database is required. Set externalDatabase.host (--set externalDatabase.host=FOO)
    or provide it through an existing secret (--set externalDatabase.existingSecret=BAR
    --set externalDatabase.existingSecretHostKey=db-host).
{{- end -}}
{{- if and (not .Values.externalDatabase.password) (not .Values.externalDatabase.existingSecret) -}}
{{ printf "\n" }}keycloak: externalDatabase.password
    An external PostgreSQL password is required. Set externalDatabase.password (--set externalDatabase.password=BAR)
    or provide an existing secret (--set externalDatabase.existingSecret=BAR).
{{- end -}}
{{- end -}}
