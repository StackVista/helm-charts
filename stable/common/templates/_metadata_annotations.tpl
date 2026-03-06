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

{{/*
Generate a checksum annotation for a given value (map/object).
Takes a dict with:
  - value: the object to checksum
  - name: (optional) the annotation name, defaults to "checksum/object"
Usage:
  {{- include "common.annotations.checksumObject" (dict "value" .Values.global.s3proxy.credentials "name" "checksum/s3proxy-credentials") }}
*/}}
{{- define "common.annotations.checksumObject" -}}
{{ default "checksum/object" .name }}: {{ toJson .value | sha256sum | quote }}
{{- end -}}
