{{- define "sts.values.receiverKey" -}}
{{- if .Values.receiverApiKey -}}
{{ .Values.receiverApiKey }}
{{- else -}}
{{ randAlphaNum 32 | quote }}
{{- end -}}
{{- end -}}

{{/* Checks whether an `adminPassword` is set, if not it will generate a new password and set this for printing. */}}
{{/* If the password is set, this function will validate it is correctly `bcrypt` hashed, if not, it will hash it for the outputted values. */}}
{{- define "sts.values.getOrGenerateAdminPassword" -}}
{{- $pwd := .Values.adminPassword -}}
{{- if $pwd -}}
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
