{{/*
KAFKAUP-OPERATOR SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get kafkaup-operator resources based on sizing profile
Usage: {{ include "common.sizing.kafkaup-operator.resources" . }}
*/}}
{{- define "common.sizing.kafkaup-operator.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
requests:
  cpu: 10m
  memory: 32Mi
limits:
  cpu: 10m
  memory: 64Mi
{{- end }}
{{- end }}
{{- end }}
