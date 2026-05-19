{{- /*
common.labelize takes a dict or map and generates labels.

Values will be quoted. Keys will not.

Example output:

  first: "Matt"
  last: "Butcher"

*/ -}}
{{- define "common.labelize" -}}
{{- range $key, $value := . }}
{{ $key }}: {{ $value | quote }}
{{- end -}}
{{- end -}}

{{- /*
common.labels.standard prints the standard Helm labels.

The standard labels are frequently used in metadata.
*/ -}}
{{- define "common.labels.standard" -}}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/name: {{ template "common.name" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
helm.sh/chart: {{ template "common.chartref" . }}
{{- end -}}

{{- /*
common.labels.selector prints the selector Helm labels.

The selector labels are frequently used in specific pod, or component targeting.
*/ -}}
{{- define "common.labels.selector" -}}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/name: {{ template "common.name" . }}
{{- end -}}

{{- /*
common.labels.custom prints the custom common Helm labels.

Allow users of helm charts to specify custom common labels via .Values.commonLabels
*/ -}}
{{- define "common.labels.custom" -}}
{{- include "common.labelize" .Values.commonLabels }}
{{- end -}}
