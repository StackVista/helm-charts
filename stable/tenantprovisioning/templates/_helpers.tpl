{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "fullname" -}}
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
{{- define "chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "tenantprovisioning-labels" -}}
app.kubernetes.io/component: tenantprovisioning
app.kubernetes.io/name: tenantprovisioning
{{- end -}}

{{/*
Common labels
*/}}
{{- define "labels" -}}
app.kubernetes.io/name: {{ include "name" . }}
helm.sh/chart: {{ include "chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Calculate baseUrl
*/}}
{{- define "baseUrl" -}}
{{- if .Values.baseUrl -}}
{{ .Values.baseUrl }}
{{- else if .Values.ingress.host -}}
{{- if .Values.ingress.tls.enabled -}}
https://{{- .Values.ingress.host -}}
{{- else if .Values.ingress.host -}}
http://{{- .Values.ingress.host -}}
{{- else -}}
http://{{ include "fullname" . }}.{{ .Release.Namespace }}.svc
{{- end -}}
{{- end -}}
{{- end -}}


{{- define "image-registry-global" -}}
  {{- if .Values.global }}
    {{- .Values.global.imageRegistry | default "quay.io" -}}
  {{- else -}}
    quay.io
  {{- end -}}
{{- end -}}

{{- define "image-registry" -}}
  {{- if .Values.image.registry -}}
    {{- .Values.image.registry  -}}
  {{- else -}}
    {{- include "image-registry-global" . }}
  {{- end -}}
{{- end -}}

{{- define "image-pull-secrets" -}}
  {{- $pullSecrets := list }}
  {{- range .Values.global.imagePullSecrets -}}
    {{- $pullSecrets = append $pullSecrets .  -}}
  {{- end -}}
  {{- if (not (empty $pullSecrets)) -}}
imagePullSecrets:
    {{- range $pullSecrets | uniq }}
  - name: {{ . }}
    {{- end }}
  {{- end -}}
{{- end -}}

{{- define "imageTag" -}}
{{- $tag := .Values.image.tag | quote -}}
{{- $tag | replace "\"" "" -}}
{{- end -}}
