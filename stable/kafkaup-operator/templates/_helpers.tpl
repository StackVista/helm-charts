
{{/*
Common labels for all resources. Existing .Values.commonLabels takes precedence over global.commonLabels.
*/}}
{{- define "kafkaup-operator.commonLabels" -}}
{{- if or .Values.commonLabels .Values.global.commonLabels -}}
{{- $globalLabels := .Values.global.commonLabels | default dict -}}
{{- $commonLabels := .Values.commonLabels | default dict -}}
{{- $mergedLabels := merge $commonLabels $globalLabels -}}
{{- range $key, $value := $mergedLabels }}
{{ $key }}: {{ $value | quote }}
{{- end }}
{{- end }}
{{- end -}}

{{- define "kafkaup-operator.checksum-configs" }}
checksum/configmap: {{ include (print $.Template.BasePath "/kafkaup-configmap.yaml") . | sha256sum }}
{{- end }}
