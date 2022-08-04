{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "stackstate-agent.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "stackstate-agent.fullname" -}}
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
{{- define "stackstate-agent.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "stackstate-agent.labels" -}}
app.kubernetes.io/name: {{ include "stackstate-agent.name" . }}
helm.sh/chart: {{ include "stackstate-agent.chart" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Cluster agent checksum annotations
*/}}
{{- define "stackstate-agent.checksum-configs" }}
checksum/secret: {{ include (print $.Template.BasePath "/secret.yaml") . | sha256sum }}
{{- end }}

{{/*
StackState URL function
*/}}
{{- define "stackstate-agent.stackstate.url" -}}
{{ tpl .Values.stackstate.url . | quote }}
{{- end }}

{{- define "stackstate-agent.configmap.override.checksum" -}}
{{- if .Values.stackstateAgent.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/stackstate-agent-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}

{{- define "stackstate-agent.agent.configmap.override.checksum" -}}
{{- if .Values.agent.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/agent-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}

{{- define "stackstate-agent.clusterChecks.configmap.override.checksum" -}}
{{- if .Values.clusterChecks.config.override }}
checksum/override-configmap: {{ include (print $.Template.BasePath "/agent-clusterchecks-configmap.yaml") . | sha256sum }}
{{- end }}
{{- end }}


{{/*
Return the image registry
*/}}
{{- define "stackstate-agent.imageRegistry" -}}
  {{- if .Values.global }}
    {{- .Values.global.imageRegistry | default .Values.all.image.registry -}}
  {{- else -}}
    {{- .Values.all.image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Renders a value that contains a template.
Usage:
{{ include "stackstate-agent.tplvalue.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "stackstate-agent.tplvalue.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names evaluating values as templates
{{ include "stackstate-agent.image.pullSecret.name" ( dict "images" (list .Values.path.to.the.image1, .Values.path.to.the.image2) "context" $) }}
*/}}
{{- define "stackstate-agent.image.pullSecret.name" -}}
  {{- $pullSecrets := list }}
  {{- $context := .context }}

  {{- if $context.Values.global }}
    {{- range $context.Values.global.imagePullSecrets -}}
      {{/* Is plain array of strings, compatible with all bitnami charts */}}
      {{- $pullSecrets = append $pullSecrets (include "stackstate-agent.tplvalue.render" (dict "value" . "context" $context)) -}}
    {{- end -}}
  {{- end -}}
  {{- range $context.Values.imagePullSecrets -}}
    {{- $pullSecrets = append $pullSecrets (include "stackstate-agent.tplvalue.render" (dict "value" .name "context" $context)) -}}
  {{- end -}}
  {{- range .images -}}
    {{- if .pullSecretName -}}
      {{- $pullSecrets = append $pullSecrets (include "stackstate-agent.tplvalue.render" (dict "value" .pullSecretName "context" $context)) -}}
    {{- else if (or .pullSecretUsername .pullSecretDockerConfigJson) -}}
      {{- $pullSecrets = append $pullSecrets ((list (include "stackstate-agent.fullname" $context ) "pull-secret") | join "-")  -}}
    {{- end -}}
  {{- end -}}

  {{- if (not (empty $pullSecrets)) }}
imagePullSecrets:
    {{- range $pullSecrets | uniq }}
  - name: {{ . }}
    {{- end }}
  {{- end }}
{{- end -}}

{{/*
Check whether the kubernetes-state-metrics configuration is overridden. If so, return 'true' else return nothing (which is false).
{{ include "stackstate-agent.kube-state-metrics.overridden" $ }}
*/}}
{{- define "stackstate-agent.kube-state-metrics.overridden" -}}
{{- if .Values.stackstateAgent.config.override }}
  {{- range $i, $val := .Values.stackstateAgent.config.override }}
    {{- if and (eq $val.name "conf.yaml") (eq $val.path "/etc/stackstate-agent/conf.d/kubernetes_state.d") }}
true
    {{- end }}
  {{- end }}
{{- end }}
{{- end -}}

{{- define "stackstate-agent.agent.kube-state-metrics.overridden" -}}
{{- if .Values.agent.config.override }}
  {{- range $i, $val := .Values.agent.config.override }}
    {{- if and (eq $val.name "auto_conf.yaml") (eq $val.path "/etc/stackstate-agent/conf.d/kubernetes_state.d") }}
true
    {{- end }}
  {{- end }}
{{- end }}
{{- end -}}
