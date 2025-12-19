{{/*
Victoria Metrics Sizing Integration
This file provides wrappers for sizing-based configuration overrides.
*/}}

{{/*
Get Victoria Metrics server resources from sizing profile if applicable.
Automatically detects victoria-metrics-0 or victoria-metrics-1 based on the fullname.
*/}}
{{- define "victoria-metrics.server.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $sizingTemplate := "" -}}
{{- if contains "victoria-metrics-0" (include "victoria-metrics.server.fullname" .) -}}
{{- $sizingTemplate = "common.sizing.victoria-metrics-0.server.resources" -}}
{{- else if contains "victoria-metrics-1" (include "victoria-metrics.server.fullname" .) -}}
{{- $sizingTemplate = "common.sizing.victoria-metrics-1.server.resources" -}}
{{- end -}}
{{- if $sizingTemplate -}}
{{- $serverResources := include $sizingTemplate . | trim -}}
{{- if $serverResources -}}
{{- $serverResources -}}
{{- else -}}
{{- toYaml .Values.server.resources -}}
{{- end -}}
{{- else -}}
{{- toYaml .Values.server.resources -}}
{{- end -}}
{{- else -}}
{{- toYaml .Values.server.resources -}}
{{- end -}}
{{- end -}}
