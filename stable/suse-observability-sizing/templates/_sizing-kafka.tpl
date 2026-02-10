{{/*
KAFKA SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get kafka replicaCount
Usage: {{ include "common.sizing.kafka.replicaCount" . }}
Returns: 1 for non-HA profiles, 3 for HA profiles, empty if no profile set
*/}}
{{- define "common.sizing.kafka.replicaCount" -}}
{{- $profileMap := dict "trial" "1" "10-nonha" "1" "20-nonha" "1" "50-nonha" "1" "100-nonha" "1" "150-ha" "3" "250-ha" "3" "500-ha" "3" "4000-ha" "3" -}}
{{- include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) -}}
{{- end }}

{{/*
Get effective kafka replicaCount with backwards-compatibility support.
Usage: {{ include "common.sizing.kafka.effectiveReplicaCount" . }}
Returns: Resolved replica count (callers pipe to | int as needed)
*/}}
{{- define "common.sizing.kafka.effectiveReplicaCount" -}}
{{- $sizingReplicaCount := include "common.sizing.kafka.replicaCount" . | trim -}}
{{- include "common.sizing.effectiveReplicaCount" (dict "sizingReplicaCount" $sizingReplicaCount "chartDefault" "3" "valuesReplicaCount" .Values.replicaCount) -}}
{{- end }}

{{/*
Get kafka numRecoveryThreadsPerDataDir
Usage: {{ include "common.sizing.kafka.numRecoveryThreadsPerDataDir" . }}
*/}}
{{- define "common.sizing.kafka.numRecoveryThreadsPerDataDir" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "500-ha" }}6
{{- else if eq $profile "4000-ha" }}8
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka deleteTopicEnable
Usage: {{ include "common.sizing.kafka.deleteTopicEnable" . }}
*/}}
{{- define "common.sizing.kafka.deleteTopicEnable" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}true
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka extraEnv for REPLICA_FETCH_MAX_BYTES
Usage: {{ include "common.sizing.kafka.extraEnv.open" . }}
*/}}
{{- define "common.sizing.kafka.extraEnv.open" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
KAFKA_CFG_REPLICA_FETCH_MAX_BYTES: "4194304"
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
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}1
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka resources
Usage: {{ include "common.sizing.kafka.resources" . }}
*/}}
{{- define "common.sizing.kafka.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") }}
requests:
  cpu: "800m"
  memory: "2048Mi"
limits:
  cpu: "1600m"
  memory: "2048Mi"
{{- else if eq $profile "20-nonha" }}
requests:
  cpu: "1000m"
  memory: "2048Mi"
limits:
  cpu: "2000m"
  memory: "2048Mi"
{{- else if eq $profile "50-nonha" }}
requests:
  cpu: "1500m"
  memory: "2048Mi"
limits:
  cpu: "3000m"
  memory: "2048Mi"
{{- else if eq $profile "100-nonha" }}
requests:
  cpu: "2000m"
  memory: "3000Mi"
limits:
  cpu: "4000m"
  memory: "3000Mi"
{{- else if eq $profile "150-ha" }}
requests:
  cpu: "1"
  memory: "3Gi"
limits:
  cpu: "2"
  memory: "4Gi"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "2"
  memory: "3Gi"
limits:
  cpu: "4"
  memory: "4Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "3"
  memory: "3Gi"
limits:
  cpu: "6"
  memory: "4Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "4000m"
  memory: "6Gi"
limits:
  cpu: "5000m"
  memory: "8Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka metrics.jmx.resources
Usage: {{ include "common.sizing.kafka.metrics.jmx.resources" . }}
*/}}
{{- define "common.sizing.kafka.metrics.jmx.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
requests:
  cpu: "500m"
  memory: "300Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "1000m"
  memory: "300Mi"
  ephemeral-storage: "1Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafka persistence.size
Usage: {{ include "common.sizing.kafka.persistence.size" . }}
*/}}
{{- define "common.sizing.kafka.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") }}60Gi
{{- else if or (eq $profile "500-ha") (eq $profile "4000-ha") }}400Gi
{{- end }}
{{- end }}
{{- end }}
