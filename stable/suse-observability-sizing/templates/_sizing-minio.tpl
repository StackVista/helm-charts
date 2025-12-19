{{/*
=============================================================================
MINIO SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get minio resources based on sizing profile
Usage: {{ include "common.sizing.minio.resources" . }}
*/}}
{{- define "common.sizing.minio.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
requests:
  cpu: 250m
  memory: 256Mi
limits:
  cpu: 250m
  memory: 256Mi
{{- end }}
{{- end }}
{{- end }}
