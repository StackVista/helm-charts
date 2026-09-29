{{/*
Pick an external or predefined internal secret.
Legacy interface retained for templates outside this chart. Chart-owned
selectors below use dedicated internal resource-name helpers.
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
{{- if .Values.stackstate.license.fromExternalSecret -}}
{{ .Values.stackstate.license.fromExternalSecret | quote }}
{{- else -}}
{{ include "stackstate.license.secret.fullname" . | quote }}
{{- end -}}
{{- end }}

{{/*
Secret for api key.
*/}}
{{- define "stackstate.secret.name.apiKey" -}}
{{- if .Values.stackstate.apiKey.fromExternalSecret -}}
{{ .Values.stackstate.apiKey.fromExternalSecret | quote }}
{{- else -}}
{{ include "stackstate.apiKey.secret.fullname" . | quote }}
{{- end -}}
{{- end }}

{{/*
Secret for auth.
*/}}
{{- define "stackstate.secret.name.auth" -}}
{{- if .Values.stackstate.authentication.fromExternalSecret -}}
{{ .Values.stackstate.authentication.fromExternalSecret | quote }}
{{- else -}}
{{ include "stackstate.auth.secret.fullname" . | quote }}
{{- end -}}
{{- end }}


{{/*
Secret for email.
*/}}
{{- define "stackstate.secret.name.email" -}}
{{- if .Values.stackstate.email.server.auth.fromExternalSecret -}}
{{ .Values.stackstate.email.server.auth.fromExternalSecret | quote }}
{{- else -}}
{{ include "stackstate.email.secret.fullname" . | quote }}
{{- end -}}
{{- end }}

{{/*
Resolvers for the trust stores and certificates mounted into the StackState services.

Each returns a `name`/`key` document identifying the secret holding the blob, or nothing at all when
the blob is not configured, so callers can use them both as an enabled-check and as a secret
reference. An external secret takes precedence over the inline values: inline values end up in the
Helm release secret, which for a trust store of a few hundred KB can push the release secret past
the 1MB limit imposed on the underlying etcd object.
*/}}
{{- define "stackstate.trustStore.java" -}}
{{- $external := default (dict) .Values.stackstate.java.trustStoreFromExternalSecret -}}
{{- if $external.name }}
name: {{ $external.name }}
key: {{ default "java-cacerts" $external.key }}
{{- else if or .Values.stackstate.java.trustStore .Values.stackstate.java.trustStoreBase64Encoded }}
name: {{ include "stackstate.common.secret.fullname" . }}
key: javaTrustStore
{{- end }}
{{- end -}}

{{- define "stackstate.trustStore.java.password" -}}
{{- $external := default (dict) .Values.stackstate.java.trustStoreFromExternalSecret -}}
{{- if and $external.name $external.passwordKey }}
name: {{ $external.name }}
key: {{ $external.passwordKey }}
{{- else if .Values.stackstate.java.trustStorePassword }}
name: {{ include "stackstate.common.secret.fullname" . }}
key: javaTrustStorePassword
{{- end }}
{{- end -}}

{{- define "stackstate.trustStore.ldap" -}}
{{- $ssl := default (dict) (default (dict) .Values.stackstate.authentication.ldap).ssl -}}
{{- $external := default (dict) $ssl.trustStoreFromExternalSecret -}}
{{- if $external.name }}
name: {{ $external.name }}
key: {{ default "ldap-cacerts" $external.key }}
{{- else if or $ssl.trustStore $ssl.trustStoreBase64Encoded }}
name: {{ include "stackstate.common.secret.fullname" . }}
key: ldapTrustStore
{{- end }}
{{- end -}}

{{- define "stackstate.trustStore.ldapCertificates" -}}
{{- $ssl := default (dict) (default (dict) .Values.stackstate.authentication.ldap).ssl -}}
{{- $external := default (dict) $ssl.trustCertificatesFromExternalSecret -}}
{{- if $external.name }}
name: {{ $external.name }}
key: {{ default "ldap-certificates.pem" $external.key }}
{{- else if or $ssl.trustCertificates $ssl.trustCertificatesBase64Encoded }}
name: {{ include "stackstate.common.secret.fullname" . }}
key: ldapTrustCertificates
{{- end }}
{{- end -}}
