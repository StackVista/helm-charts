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


{{/*
Secret for email.
*/}}
{{- define "stackstate.secret.name.email" -}}
{{ include "stackstate.secret.externalOrInternal" (merge (dict "externalSecret" .Values.stackstate.email.server.auth.fromExternalSecret "internalSecretName" "email") .) | quote }}
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
name: {{ template "common.fullname.short" . }}-common
key: javaTrustStore
{{- end }}
{{- end -}}

{{- define "stackstate.trustStore.java.password" -}}
{{- $external := default (dict) .Values.stackstate.java.trustStoreFromExternalSecret -}}
{{- if and $external.name $external.passwordKey }}
name: {{ $external.name }}
key: {{ $external.passwordKey }}
{{- else if .Values.stackstate.java.trustStorePassword }}
name: {{ template "common.fullname.short" . }}-common
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
name: {{ template "common.fullname.short" . }}-common
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
name: {{ template "common.fullname.short" . }}-common
key: ldapTrustCertificates
{{- end }}
{{- end -}}
