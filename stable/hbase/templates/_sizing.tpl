{{/*
Sizing-related helper templates for HBase chart.
These helpers determine configuration based on sizing profiles when available,
falling back to values.yaml defaults.
*/}}

{{/*
Get HBase deployment mode with sizing profile evaluation.
Auto-detects if global.suseObservability values are set and uses sizing profile
configuration when available, otherwise falls back to values.yaml.

Usage: {{ include "hbase.deployment.mode" . }}
Returns: "Distributed" or "Mono"
*/}}
{{- define "hbase.deployment.mode" -}}
{{- $useSizingProfile := false -}}
{{- if and .Values.global .Values.global.suseObservability -}}
  {{- $hasProfile := and .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile (ne .Values.global.suseObservability.sizing.profile "") -}}
  {{- $hasLicense := and .Values.global.suseObservability.license (ne .Values.global.suseObservability.license "") -}}
  {{- $hasBaseUrl := and .Values.global.suseObservability.baseUrl (ne .Values.global.suseObservability.baseUrl "") -}}
  {{- if or $hasProfile $hasLicense $hasBaseUrl -}}
    {{- $useSizingProfile = true -}}
  {{- end -}}
{{- end -}}
{{- if $useSizingProfile -}}
{{- $profileMode := include "common.sizing.hbase.deploymentMode" . | trim -}}
{{- if $profileMode -}}
{{- $profileMode -}}
{{- else -}}
{{- .Values.deployment.mode | default "Distributed" -}}
{{- end -}}
{{- else -}}
{{- .Values.deployment.mode | default "Distributed" -}}
{{- end -}}
{{- end -}}
