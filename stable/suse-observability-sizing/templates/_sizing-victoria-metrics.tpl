VICTORIA METRICS TEMPLATES
=============================================================================
*/}}

{{/*
Get victoria-metrics-0 server resources
Usage: {{ include "common.sizing.victoria-metrics-0.server.resources" . }}
*/}}
{{- define "common.sizing.victoria-metrics-0.server.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") }}
requests:
  cpu: 2
  memory: 9Gi
limits:
  cpu: 4
  memory: 9Gi
{{- else if eq $profile "500-ha" }}
requests:
  cpu: 3
  memory: 12Gi
limits:
  cpu: 6
  memory: 12Gi
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: 4
  memory: 16Gi
limits:
  cpu: 8
  memory: 16Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get victoria-metrics-0 vmbackup resources
Usage: {{ include "common.sizing.victoria-metrics-0.vmbackup.resources" . }}
*/}}
{{- define "common.sizing.victoria-metrics-0.vmbackup.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  memory: 512Mi
limits:
  memory: 512Mi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get victoria-metrics-1 enabled flag
Usage: {{ include "common.sizing.victoria-metrics-1.enabled" . }}
*/}}
{{- define "common.sizing.victoria-metrics-1.enabled" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}true
{{- end }}
{{- end }}
{{- end }}

{{/*
Get victoria-metrics-1 server resources
Usage: {{ include "common.sizing.victoria-metrics-1.server.resources" . }}
*/}}
{{- define "common.sizing.victoria-metrics-1.server.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") }}
requests:
  cpu: 2
  memory: 9Gi
limits:
  cpu: 4
  memory: 9Gi
{{- else if eq $profile "500-ha" }}
requests:
  cpu: 3
  memory: 12Gi
limits:
  cpu: 6
  memory: 12Gi
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: 4
  memory: 16Gi
limits:
  cpu: 8
  memory: 16Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get victoria-metrics-1 vmbackup resources
Usage: {{ include "common.sizing.victoria-metrics-1.vmbackup.resources" . }}
*/}}
{{- define "common.sizing.victoria-metrics-1.vmbackup.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  memory: 512Mi
limits:
  memory: 512Mi
{{- end }}
{{- end }}
{{- end }}


{{/*
=============================================================================
*/}}
