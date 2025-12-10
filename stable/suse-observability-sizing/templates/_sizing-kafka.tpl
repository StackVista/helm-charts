{{/*
=============================================================================
KAFKA SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get kafka resources based on sizing profile
Usage: {{ include "common.sizing.kafka.resources" . }}
*/}}
{{- define "common.sizing.kafka.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}
requests:
  cpu: "250m"
  memory: "750Mi"
limits:
  cpu: "500m"
  memory: "1Gi"
{{- else if eq $profile "10-nonha" }}
requests:
  cpu: "250m"
  memory: "750Mi"
limits:
  cpu: "500m"
  memory: "1Gi"
{{- else if eq $profile "20-nonha" }}
requests:
  cpu: "500m"
  memory: "1500Mi"
limits:
  cpu: "1"
  memory: "2Gi"
{{- else if eq $profile "50-nonha" }}
requests:
  cpu: "500m"
  memory: "1500Mi"
limits:
  cpu: "1"
  memory: "2Gi"
{{- else if eq $profile "100-nonha" }}
requests:
  cpu: "500m"
  memory: "1500Mi"
limits:
  cpu: "1"
  memory: "2Gi"
{{- else if eq $profile "150-ha" }}
requests:
  cpu: "1"
  memory: "3Gi"
limits:
  cpu: "2"
  memory: "4Gi"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "1"
  memory: "3Gi"
limits:
  cpu: "2"
  memory: "4Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "1500m"
  memory: "4Gi"
limits:
  cpu: "3"
  memory: "5Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "2"
  memory: "5Gi"
limits:
  cpu: "4"
  memory: "6Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka persistence size based on sizing profile
Usage: {{ include "common.sizing.kafka.persistence.size" . }}
*/}}
{{- define "common.sizing.kafka.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") (eq $profile "150-ha") }}50Gi
{{- else if eq $profile "250-ha" }}100Gi
{{- else if eq $profile "500-ha" }}200Gi
{{- else if eq $profile "4000-ha" }}500Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka replicaCount
Usage: {{ include "common.sizing.kafka.replicaCount" . }}
*/}}
{{- define "common.sizing.kafka.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}1
{{- else }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka transactionStateLogReplicationFactor
Usage: {{ include "common.sizing.kafka.transactionStateLogReplicationFactor" . }}
*/}}
{{- define "common.sizing.kafka.transactionStateLogReplicationFactor" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}1
{{- else }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka defaultReplicationFactor
Usage: {{ include "common.sizing.kafka.defaultReplicationFactor" . }}
*/}}
{{- define "common.sizing.kafka.defaultReplicationFactor" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}1
{{- else }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka offsetsTopicReplicationFactor
Usage: {{ include "common.sizing.kafka.offsetsTopicReplicationFactor" . }}
*/}}
{{- define "common.sizing.kafka.offsetsTopicReplicationFactor" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}1
{{- else }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka affinity
Usage: {{ include "common.sizing.kafka.affinity" . }}
*/}}
{{- define "common.sizing.kafka.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "kafka") "context" .) }}
{{- end }}
