{{- define "sts.values.receiverKey" -}}
{{- if .Values.receiverApiKey -}}
{{ .Values.receiverApiKey }}
{{- else -}}
{{ randAlphaNum 32 | quote }}
{{- end -}}
{{- end -}}

{{- define "sts.values.getOrGenerateAdminPassword" -}}
{{- $pwd := .Values.adminPassword -}}
{{- if $pwd -}}
{{/*2[abxy]?\$(0[4-9]|[12][0-9]|3[01])\$[./0-9a-zA-Z]{53}$ */}}
{{- if regexMatch "^\\$2[abxy]{0,1}\\$(0[4-9]|[12][0-9]|3[01])\\$[./0-9a-zA-Z]{53}$" $pwd -}}
{{ $pwd }}
{{- else -}}
{{ $pwd | bcrypt | quote }}
{{- end -}}
{{- else -}}
{{- $pwd = randAlphaNum 16 -}}
{{- $ignored := set .Values "generatedAdminPassword" $pwd -}}
{{ $pwd | bcrypt | quote }}
{{- end -}}
{{- end -}}

{{- define "sts.values.getOrGenerateAdminApiPassword" -}}
{{- $pwd := .Values.adminApiPassword -}}
{{- if $pwd -}}
{{/*2[abxy]?\$(0[4-9]|[12][0-9]|3[01])\$[./0-9a-zA-Z]{53}$ */}}
{{- if regexMatch "^\\$2[abxy]{0,1}\\$(0[4-9]|[12][0-9]|3[01])\\$[./0-9a-zA-Z]{53}$" $pwd -}}
{{ $pwd }}
{{- else -}}
{{ $pwd | bcrypt | quote }}
{{- end -}}
{{- else -}}
{{- $pwd = randAlphaNum 16 -}}
{{- $ignored := set .Values "generatedAdminApiPassword" $pwd -}}
{{ $pwd | bcrypt | quote }}
{{- end -}}
{{- end -}}
