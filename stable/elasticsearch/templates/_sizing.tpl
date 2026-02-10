{{/*
Elasticsearch Sizing Integration
This file provides wrappers for sizing-based configuration overrides.
*/}}

{{/*
Get Elasticsearch esJavaOpts from sizing profile if applicable.
*/}}
{{- define "elasticsearch.esJavaOpts" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $esJavaOpts := include "common.sizing.elasticsearch.esJavaOpts" . | trim -}}
{{- if $esJavaOpts -}}
{{- $esJavaOpts -}}
{{- else -}}
{{- .Values.esJavaOpts -}}
{{- end -}}
{{- else -}}
{{- .Values.esJavaOpts -}}
{{- end -}}
{{- end -}}
