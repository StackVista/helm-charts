
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

{{- define "redirector-labels" -}}
app.kubernetes.io/component: redirector
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: redirector
{{- end -}}

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
