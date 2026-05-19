{{/*
Elasticsearch Sizing Integration
This file provides wrappers for sizing-based configuration overrides.
*/}}

{{/*
Get Elasticsearch esJavaOpts from sizing profile if applicable.
*/}}
{{- define "elasticsearch.esJavaOpts" -}}
{{- $sizingOpts := include "common.sizing.elasticsearch.esJavaOpts" . | trim -}}
{{- default $sizingOpts .Values.esJavaOpts -}}
{{- end -}}
