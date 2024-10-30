{{/*
Pick an external or predefined internal secret.
*/}}
{{- define "stackstate.secret.externalOrInternal" -}}
{{- if .externalSecret }}
{{- .externalSecret }}
{{- else }}
{{- template "common.fullname.short" . }}-{{ .internalSecretName }}
{{- end }}
{{- end }}

{{/*
Secret for license.
*/}}
{{- define "stackstate.secret.name.license" -}}
{{ include "stackstate.secret.externalOrInternal" (merge (dict "externalSecret" .Values.stackstate.license.fromExternalSecret "internalSecretName" "license") .) | quote }}
{{- end }}

{{/*
Secret for api key.
*/}}
{{- define "stackstate.secret.name.apiKey" -}}
{{ include "stackstate.secret.externalOrInternal" (merge (dict "externalSecret" .Values.stackstate.apiKey.fromExternalSecret "internalSecretName" "api-key") .) | quote }}
{{- end }}

{{/*
Secret for auth.
*/}}
{{- define "stackstate.secret.name.auth" -}}
{{ include "stackstate.secret.externalOrInternal" (merge (dict "externalSecret" .Values.stackstate.authentication.fromExternalSecret "internalSecretName" "auth") .) | quote }}
{{- end }}
