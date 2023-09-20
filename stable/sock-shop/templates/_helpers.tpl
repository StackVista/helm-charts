{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "sock-shop.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "sock-shop.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Application common labels
*/}}
{{- define "sock-shop.common.labels" -}}
app.kubernetes.io/name: {{ include "sock-shop.name" . }}
helm.sh/chart: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Component common labels
*/}}
{{- define "component.common.labels" -}}
name: {{ . }}
app.kubernetes.io/component: {{ . }}
{{- end -}}

{{/*
Componment annotations
*/}}
{{- define "component.custom.annotations" -}}
{{- $ctx := . -}}
{{- if $ctx.annotations -}}
annotations:
{{- range $key, $value := $ctx.annotations }}
  {{ $key }}: {{ $value | quote }}
{{- end -}}
{{- end -}}
{{- end -}}


{{/*
A shortcut for PriorityClass name
*/}}
{{- define "priority-class-name" -}}
{{- printf "%s-%d" .Release.Name (.Values.priority | int) -}}
{{- end -}}
