{{/*
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
{{- if eq $profile "trial" -}}
requests:
  cpu: 500m
  memory: 1500Mi
limits:
  cpu: 1
  memory: 1750Mi
{{- else if eq $profile "10-nonha" -}}
requests:
  cpu: "500m"
  memory: 1500Mi
limits:
  cpu: "1000m"
  memory: 1750Mi
{{- else if eq $profile "20-nonha" -}}
requests:
  cpu: "500m"
  memory: 2Gi
limits:
  cpu: "1000m"
  memory: 2500Mi
{{- else if eq $profile "50-nonha" -}}
requests:
  cpu: "1000m"
  memory: 3500Mi
limits:
  cpu: "2000m"
  memory: 3500Mi
{{- else if eq $profile "100-nonha" -}}
requests:
  cpu: "1500m"
  memory: 8Gi
limits:
  cpu: "3000m"
  memory: 8Gi
{{- else if eq $profile "150-ha" -}}
requests:
  cpu: 2
  memory: 9Gi
limits:
  cpu: 4
  memory: 9Gi
{{- else if eq $profile "250-ha" -}}
requests:
  cpu: 3
  memory: 10Gi
limits:
  cpu: 6
  memory: 10Gi
{{- else if eq $profile "500-ha" -}}
requests:
  cpu: 3
  memory: 10Gi
limits:
  cpu: 6
  memory: 10Gi
{{- else if eq $profile "4000-ha" -}}
requests:
  cpu: 7
  memory: 16Gi
limits:
  cpu: 8000m
  memory: 18Gi
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get victoria-metrics-0 vmbackup resources
Usage: {{ include "common.sizing.victoria-metrics-0.vmbackup.resources" . }}
*/}}
{{- define "common.sizing.victoria-metrics-0.vmbackup.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "10-nonha") (eq $profile "20-nonha") }}
requests:
  memory: 256Mi
limits:
  memory: 256Mi
{{- else if or (eq $profile "50-nonha") (eq $profile "100-nonha") (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
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
{{- else if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}false
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
{{- if eq $profile "150-ha" }}
requests:
  cpu: 2
  memory: 9Gi
limits:
  cpu: 4
  memory: 9Gi
{{- else if eq $profile "250-ha" }}
requests:
  cpu: 3
  memory: 10Gi
limits:
  cpu: 6
  memory: 10Gi
{{- else if eq $profile "500-ha" }}
requests:
  cpu: 3
  memory: 10Gi
limits:
  cpu: 6
  memory: 10Gi
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: 7
  memory: 16Gi
limits:
  cpu: 8000m
  memory: 18Gi
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
Get victoria-metrics storage size based on sizing profile
Usage: {{ include "common.sizing.victoria-metrics.storage" . }}
*/}}
{{- define "common.sizing.victoria-metrics.storage" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}10Gi
{{- else if or (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}50Gi
{{- else if eq $profile "4000-ha" }}400Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get victoria-metrics retention period based on sizing profile
Usage: {{ include "common.sizing.victoria-metrics.retention" . }}
*/}}
{{- define "common.sizing.victoria-metrics.retention" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}3d
{{- end }}
{{- end }}
{{- end }}

{{/*
Get Victoria Metrics affinity configurations
Usage: {{ include "common.sizing.victoria-metrics-0.affinityConfig" . }}
Usage: {{ include "common.sizing.victoria-metrics-1.affinityConfig" . }}
*/}}
{{- define "common.sizing.victoria-metrics-0.affinityConfig" -}}
{{- end -}}
{{- define "common.sizing.victoria-metrics-1.affinityConfig" -}}
{{- end -}}
