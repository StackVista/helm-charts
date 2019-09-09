{{- /*
common.hook defines a hook.

This is to be used in a 'metadata.annotations' section.

This should be called as 'template "common.metadata.hook" "post-install"'

Any valid hook may be passed in. Separate multiple hooks with a ",".
*/ -}}
{{- define "common.hook" -}}
"helm.sh/hook": {{ printf "%s" . | quote }}
{{- end -}}

{{- define "common.annotations.gitlab" -}}
{{- if .Values.gitlab.app }}
app.gitlab.com/app: {{ .Values.gitlab.app | quote }}
{{- end }}
{{- if .Values.gitlab.env }}
app.gitlab.com/env: {{ .Values.gitlab.env | quote }}
{{- end }}
{{- end -}}


{{- define "common.annotations" -}}
{{- range $key, $value := . }}
{{ $key | quote }}: {{ $value | quote }}
{{- end -}}
{{- end -}}
