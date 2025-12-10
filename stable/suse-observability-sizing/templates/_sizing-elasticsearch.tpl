ELASTICSEARCH TEMPLATES
=============================================================================
*/}}

{{/*
Get elasticsearch storage size based on sizing profile
Usage: {{ include "common.sizing.elasticsearch.storage" . }}
*/}}
{{- define "common.sizing.elasticsearch.storage" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}20Gi
{{- else if eq $profile "10-nonha" }}50Gi
{{- else if eq $profile "20-nonha" }}100Gi
{{- else if eq $profile "50-nonha" }}150Gi
{{- else if eq $profile "100-nonha" }}200Gi
{{- else if eq $profile "150-ha" }}200Gi
{{- else if eq $profile "250-ha" }}300Gi
{{- else if eq $profile "500-ha" }}500Gi
{{- else if eq $profile "4000-ha" }}1000Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get elasticsearch resources based on sizing profile
Usage: {{ include "common.sizing.elasticsearch.resources" . }}
*/}}
{{- define "common.sizing.elasticsearch.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
requests:
  cpu: "4"
  memory: 4Gi
limits:
  cpu: "8"
  memory: 4Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get elasticsearch replicas based on sizing profile
Usage: {{ include "common.sizing.elasticsearch.replicas" . }}
*/}}
{{- define "common.sizing.elasticsearch.replicas" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}1
{{- else }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
=============================================================================
*/}}
