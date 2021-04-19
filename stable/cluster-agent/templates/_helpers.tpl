{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "cluster-agent.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "cluster-agent.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cluster-agent.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "cluster-agent.labels" -}}
app.kubernetes.io/name: {{ include "cluster-agent.name" . }}
helm.sh/chart: {{ include "cluster-agent.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Cluster agent checksum annotations
*/}}
{{- define "cluster-agent.checksum-configs" }}
checksum/secret: {{ include (print $.Template.BasePath "/secret.yaml") . | sha256sum }}
{{- end }}

{{/*
StackState URL function
*/}}
{{- define "cluster-agent.stackstate.url" -}}
{{ tpl .Values.stackstate.url . | quote }}
{{- end }}

{{- define "cluster-agent.configmap.override.checksum" -}}
{{- if .Values.clusterAgent.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/cluster-agent-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}

{{- define "cluster-agent.agent.configmap.override.checksum" -}}
{{- if .Values.agent.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/agent-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}

{{- define "cluster-agent.clusterChecksRunner.configmap.override.checksum" -}}
{{- if .Values.clusterChecksRunner.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/agent-clusterchecks-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}
