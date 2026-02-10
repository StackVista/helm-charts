
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

{{/*
Return the proper Docker Image Registry Secret Names
*/}}
{{- define "kafkaup.image.pullSecret.name" -}}
  {{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.pullSecret .Values.global.suseObservability.pullSecret.username .Values.global.suseObservability.pullSecret.password -}}
imagePullSecrets:
- name: suse-observability-pull-secret
  {{- else -}}
    {{- include "common.image.pullSecret.name" (dict "images" (list .Values.image) "context" $) -}}
  {{- end -}}
{{- end -}}
