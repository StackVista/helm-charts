{{/*
KAFKAUP-OPERATOR SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get kafkaup-operator resources based on sizing profile
Usage: {{ include "common.sizing.kafkaup-operator.resources" . }}
*/}}
{{- define "common.sizing.kafkaup-operator.resources" -}}
requests:
  cpu: 25m
  memory: 32Mi
limits:
  cpu: 50m
  memory: 64Mi
{{- end }}
