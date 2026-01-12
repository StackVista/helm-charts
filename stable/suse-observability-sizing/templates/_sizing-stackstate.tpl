{{/*
STACKSTATE COMPONENTS TEMPLATES
=============================================================================
*/}}

{{/*
Get stackstate server split mode
Usage: {{ include "common.sizing.stackstate.server.split" . }}
*/}}
{{- define "common.sizing.stackstate.server.split" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}false
{{- else }}true
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate agent limit config
Usage: {{ include "common.sizing.stackstate.agentLimit" . }}
*/}}
{{- define "common.sizing.stackstate.agentLimit" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}"10"
{{- else if eq $profile "10-nonha" }}"10"
{{- else if eq $profile "20-nonha" }}"20"
{{- else if eq $profile "50-nonha" }}"50"
{{- else if eq $profile "100-nonha" }}"100"
{{- else if eq $profile "150-ha" }}"150"
{{- else if eq $profile "250-ha" }}"250"
{{- else if eq $profile "500-ha" }}"500"
{{- else if eq $profile "4000-ha" }}"4000"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate server resources (for non-split mode)
Usage: {{ include "common.sizing.stackstate.server.resources" . }}
*/}}
{{- define "common.sizing.stackstate.server.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}
limits:
  ephemeral-storage: 5Gi
  cpu: 3000m
  memory: 5Gi
requests:
  cpu: 1500m
  memory: 5Gi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "10-nonha" }}
limits:
  ephemeral-storage: 5Gi
  cpu: 3000m
  memory: 5Gi
requests:
  cpu: 1500m
  memory: 5Gi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "20-nonha" }}
limits:
  ephemeral-storage: 5Gi
  cpu: 4000m
  memory: 6Gi
requests:
  cpu: 2000m
  memory: 6Gi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "50-nonha" }}
limits:
  ephemeral-storage: 5Gi
  cpu: 5000m
  memory: 6500Mi
requests:
  cpu: 2500m
  memory: 6500Mi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "100-nonha" }}
limits:
  ephemeral-storage: 5Gi
  cpu: 8000m
  memory: 8Gi
requests:
  cpu: 4000m
  memory: 8Gi
  ephemeral-storage: "1Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate API component resources
Usage: {{ include "common.sizing.stackstate.api.resources" . }}
*/}}
{{- define "common.sizing.stackstate.api.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
requests:
  cpu: "1500m"
  memory: "4Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "3000m"
  memory: "4Gi"
  ephemeral-storage: "2Gi"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "2000m"
  memory: "5Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "4000m"
  memory: "5Gi"
  ephemeral-storage: "2Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "3000m"
  memory: "6Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "6000m"
  memory: "6Gi"
  ephemeral-storage: "2Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "7000m"
  memory: "10Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "9000m"
  memory: "12Gi"
  ephemeral-storage: "2Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate checks component resources
Usage: {{ include "common.sizing.stackstate.checks.resources" . }}
*/}}
{{- define "common.sizing.stackstate.checks.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
requests:
  cpu: "3000m"
  memory: 5Gi
  ephemeral-storage: "1Mi"
limits:
  cpu: "6000m"
  memory: 5Gi
  ephemeral-storage: "1Gi"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "3000m"
  memory: 5Gi
  ephemeral-storage: "1Mi"
limits:
  cpu: "6000m"
  memory: 5Gi
  ephemeral-storage: "1Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "3000m"
  memory: 5Gi
  ephemeral-storage: "1Mi"
limits:
  cpu: "6000m"
  memory: 5Gi
  ephemeral-storage: "1Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "6000m"
  memory: 5Gi
  ephemeral-storage: "1Mi"
limits:
  cpu: "7000m"
  memory: 5Gi
  ephemeral-storage: "1Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate healthSync component resources
Usage: {{ include "common.sizing.stackstate.healthSync.resources" . }}
*/}}
{{- define "common.sizing.stackstate.healthSync.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
limits:
  memory: 5Gi
  cpu: "2000m"
  ephemeral-storage: "1Gi"
requests:
  memory: 5Gi
  cpu: "1000m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "250-ha" }}
limits:
  memory: 6Gi
  cpu: "3000m"
  ephemeral-storage: "1Gi"
requests:
  memory: 6Gi
  cpu: "1500m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "500-ha" }}
limits:
  memory: 8Gi
  cpu: "8000m"
  ephemeral-storage: "1Gi"
requests:
  memory: 8Gi
  cpu: "4000m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  memory: 14Gi
  cpu: "7000m"
  ephemeral-storage: "1Gi"
requests:
  memory: 12Gi
  cpu: "5000m"
  ephemeral-storage: "1Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate healthSync baseMemoryConsumption
Usage: {{ include "common.sizing.stackstate.healthSync.baseMemoryConsumption" . }}
*/}}
{{- define "common.sizing.stackstate.healthSync.baseMemoryConsumption" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}1200Mi
{{- else if eq $profile "4000-ha" }}1800Mi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate sync baseMemoryConsumption
Usage: {{ include "common.sizing.stackstate.sync.baseMemoryConsumption" . }}
*/}}
{{- define "common.sizing.stackstate.sync.baseMemoryConsumption" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}400Mi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate server baseMemoryConsumption
Usage: {{ include "common.sizing.stackstate.server.baseMemoryConsumption" . }}
*/}}
{{- define "common.sizing.stackstate.server.baseMemoryConsumption" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{/* No profile-specific overrides for server baseMemoryConsumption at this time */}}
{{- end }}
{{- end }}

{{/*
Get stackstate server javaHeapMemoryFraction
Usage: {{ include "common.sizing.stackstate.server.javaHeapMemoryFraction" . }}
*/}}
{{- define "common.sizing.stackstate.server.javaHeapMemoryFraction" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{/* No profile-specific overrides for server javaHeapMemoryFraction at this time */}}
{{- end }}
{{- end }}

{{/*
Get stackstate api baseMemoryConsumption
Usage: {{ include "common.sizing.stackstate.api.baseMemoryConsumption" . }}
*/}}
{{- define "common.sizing.stackstate.api.baseMemoryConsumption" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- /* No profile-specific API baseMemoryConsumption - use default 500Mi */ -}}
{{- end }}
{{- end }}

{{/*
Get stackstate api javaHeapMemoryFraction
Usage: {{ include "common.sizing.stackstate.api.javaHeapMemoryFraction" . }}
*/}}
{{- define "common.sizing.stackstate.api.javaHeapMemoryFraction" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{/* All profiles use default javaHeapMemoryFraction of 45% */}}
{{- end }}
{{- end }}

{{/*
Get stackstate state baseMemoryConsumption
Usage: {{ include "common.sizing.stackstate.state.baseMemoryConsumption" . }}
*/}}
{{- define "common.sizing.stackstate.state.baseMemoryConsumption" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{/* No profile-specific overrides for state baseMemoryConsumption at this time */}}
{{- end }}
{{- end }}

{{/*
Get stackstate state javaHeapMemoryFraction
Usage: {{ include "common.sizing.stackstate.state.javaHeapMemoryFraction" . }}
*/}}
{{- define "common.sizing.stackstate.state.javaHeapMemoryFraction" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{/* No profile-specific overrides for state javaHeapMemoryFraction at this time */}}
{{- end }}
{{- end }}

{{/*
Get stackstate initializer component resources
Usage: {{ include "common.sizing.stackstate.initializer.resources" . }}
*/}}
{{- define "common.sizing.stackstate.initializer.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}
requests:
  cpu: "50m"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "1000m"
  memory: "1Gi"
limits:
  cpu: "1500m"
  memory: "1Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate notification component resources
Usage: {{ include "common.sizing.stackstate.notification.resources" . }}
*/}}
{{- define "common.sizing.stackstate.notification.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
requests:
  cpu: 500m
  memory: "3000Mi"
limits:
  cpu: 1000m
  memory: "3000Mi"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: 1000m
  memory: "3000Mi"
limits:
  cpu: 2000m
  memory: "3000Mi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: 1500m
  memory: "3000Mi"
limits:
  cpu: 3000m
  memory: "3000Mi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: 2000m
  memory: "3200Mi"
limits:
  cpu: 3000m
  memory: "4500Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate router component resources
Usage: {{ include "common.sizing.stackstate.router.resources" . }}
*/}}
{{- define "common.sizing.stackstate.router.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") }}
requests:
  cpu: "120m"
  memory: "128Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "240m"
  memory: "128Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "120m"
  memory: "128Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "240m"
  memory: "128Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "2000m"
  memory: "256Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "2500m"
  memory: "512Mi"
  ephemeral-storage: "1Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate slicing component resources
Usage: {{ include "common.sizing.stackstate.slicing.resources" . }}
*/}}
{{- define "common.sizing.stackstate.slicing.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "500-ha") }}
requests:
  cpu: "300m"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "300m"
  memory: 1500Mi
limits:
  cpu: "600m"
  memory: 1500Mi
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "1000m"
  memory: "4Gi"
limits:
  cpu: "1500m"
  memory: "4Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate state component resources
Usage: {{ include "common.sizing.stackstate.state.resources" . }}
*/}}
{{- define "common.sizing.stackstate.state.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") }}
limits:
  cpu: "4"
  memory: "4000Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "2"
  memory: "4000Mi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "500-ha" }}
limits:
  cpu: "4"
  memory: "4000Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "2"
  memory: "4000Mi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: "5"
  memory: "6000Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "3"
  memory: "4000Mi"
  ephemeral-storage: "1Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate sync component tmpToPVC volumeSize
Usage: {{ include "common.sizing.stackstate.sync.tmpToPVC.volumeSize" . }}
*/}}
{{- define "common.sizing.stackstate.sync.tmpToPVC.volumeSize" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}10Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate sync component resources
Usage: {{ include "common.sizing.stackstate.sync.resources" . }}
*/}}
{{- define "common.sizing.stackstate.sync.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
limits:
  cpu: "4000m"
  memory: "6500Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "2000m"
  memory: "6500Mi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "250-ha" }}
limits:
  cpu: "6000m"
  memory: 8Gi
  ephemeral-storage: "1Gi"
requests:
  cpu: "3000m"
  memory: 8Gi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "500-ha" }}
limits:
  cpu: "8000m"
  memory: 8Gi
  ephemeral-storage: "1Gi"
requests:
  cpu: "4000m"
  memory: 8Gi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: "14"
  memory: "16000Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "12"
  memory: "14000Mi"
  ephemeral-storage: "1Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate e2es component resources
Usage: {{ include "common.sizing.stackstate.e2es.resources" . }}
*/}}
{{- define "common.sizing.stackstate.e2es.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
requests:
  memory: "512Mi"
  cpu: "50m"
limits:
  memory: "512Mi"
{{- else if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}
requests:
  memory: "512Mi"
  cpu: "50m"
limits:
  memory: "512Mi"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "768Mi"
  cpu: "2000m"
limits:
  memory: "1024Mi"
  cpu: "3000m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate correlate component replicaCount
Usage: {{ include "common.sizing.stackstate.correlate.replicaCount" . }}
*/}}
{{- define "common.sizing.stackstate.correlate.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "500-ha") (eq $profile "4000-ha") }}10
{{- else if eq $profile "150-ha" }}3
{{- else if eq $profile "250-ha" }}5
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate correlate component resources
Usage: {{ include "common.sizing.stackstate.correlate.resources" . }}
*/}}
{{- define "common.sizing.stackstate.correlate.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") }}
requests:
  memory: "1250Mi"
  cpu: "500m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "1000m"
  memory: "1250Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "20-nonha" }}
requests:
  memory: "1750Mi"
  cpu: "1500m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "3000m"
  memory: "1750Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "50-nonha" }}
requests:
  memory: "2000Mi"
  cpu: "2000m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "4000m"
  memory: "2000Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "100-nonha" }}
requests:
  memory: "4000Mi"
  cpu: "5000m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "10000m"
  memory: "4000Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "150-ha" }}
limits:
  cpu: 6
  memory: 3500Mi
  ephemeral-storage: "1Gi"
requests:
  cpu: 3
  memory: 3500Mi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "250-ha" }}
limits:
  cpu: 4000m
  memory: 3Gi
  ephemeral-storage: "1Gi"
requests:
  cpu: 2000m
  memory: 3Gi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "500-ha" }}
limits:
  cpu: 4000m
  memory: 3000Mi
  ephemeral-storage: "1Gi"
requests:
  cpu: 2000m
  memory: 3000Mi
  ephemeral-storage: "1Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: 6000m
  memory: 6000Mi
  ephemeral-storage: "1Gi"
requests:
  cpu: 5000m
  memory: 4000Mi
  ephemeral-storage: "1Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate receiver split enabled flag
Usage: {{ include "common.sizing.stackstate.receiver.split.enabled" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.split.enabled" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}false
{{- else }}true
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate receiver resources (non-split mode)
Usage: {{ include "common.sizing.stackstate.receiver.resources" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") }}
requests:
  memory: "1000Mi"
  cpu: "1000m"
limits:
  memory: "1000Mi"
  cpu: "2000m"
{{- else if eq $profile "20-nonha" }}
requests:
  memory: "1750Mi"
  cpu: "1000m"
limits:
  memory: "1750Mi"
  cpu: "2000m"
{{- else if eq $profile "50-nonha" }}
requests:
  memory: "2250Mi"
  cpu: "2000m"
limits:
  memory: "2250Mi"
  cpu: "4000m"
{{- else if eq $profile "100-nonha" }}
requests:
  memory: "4Gi"
  cpu: "5500m"
limits:
  memory: "4Gi"
  cpu: "11000m"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "6Gi"
  cpu: "6000m"
limits:
  memory: "6Gi"
  cpu: "12000m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate receiver retention
Usage: {{ include "common.sizing.stackstate.receiver.retention" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.retention" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "4000-ha") }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate receiver split base resources
Usage: {{ include "common.sizing.stackstate.receiver.split.base.resources" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.split.base.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
requests:
  memory: "5500Mi"
  cpu: "4000m"
limits:
  memory: "5500Mi"
  cpu: "8000m"
{{- else if eq $profile "250-ha" }}
requests:
  memory: "6500Mi"
  cpu: "4000m"
limits:
  memory: "6500Mi"
  cpu: "8000m"
{{- else if eq $profile "500-ha" }}
requests:
  memory: "7Gi"
  cpu: "6000m"
limits:
  memory: "7Gi"
  cpu: "12000m"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "6Gi"
  cpu: "8"
limits:
  memory: "7Gi"
  cpu: "9000m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate receiver split processAgent resources
Usage: {{ include "common.sizing.stackstate.receiver.split.processAgent.resources" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.split.processAgent.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
requests:
  memory: "2500Mi"
  cpu: "1000m"
limits:
  memory: "2500Mi"
  cpu: "2000m"
{{- else if eq $profile "250-ha" }}
requests:
  memory: "3Gi"
  cpu: "1000m"
limits:
  memory: "3Gi"
  cpu: "2000m"
{{- else if eq $profile "500-ha" }}
requests:
  memory: "3Gi"
  cpu: "1000m"
limits:
  memory: "3Gi"
  cpu: "2000m"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "4Gi"
  cpu: "6"
limits:
  memory: "6Gi"
  cpu: "8000m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate receiver split logs resources
Usage: {{ include "common.sizing.stackstate.receiver.split.logs.resources" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.split.logs.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "150-ha" }}
requests:
  memory: "3Gi"
  cpu: "1000m"
limits:
  memory: "3Gi"
  cpu: "2000m"
{{- else if eq $profile "250-ha" }}
requests:
  memory: "3Gi"
  cpu: "1000m"
limits:
  memory: "3Gi"
  cpu: "2000m"
{{- else if eq $profile "500-ha" }}
requests:
  memory: "3Gi"
  cpu: "1000m"
limits:
  memory: "3Gi"
  cpu: "2000m"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "4Gi"
  cpu: "2"
limits:
  memory: "6Gi"
  cpu: "3"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate vmagent resources
Usage: {{ include "common.sizing.stackstate.vmagent.resources" . }}
*/}}
{{- define "common.sizing.stackstate.vmagent.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") }}
limits:
  cpu: "200m"
  memory: "640Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "200m"
  memory: "384Mi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "20-nonha" }}
limits:
  memory: "768Mi"
  cpu: "600m"
  ephemeral-storage: "1Gi"
requests:
  memory: "600Mi"
  cpu: "300m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "50-nonha" }}
limits:
  memory: "750Mi"
  cpu: "1000m"
  ephemeral-storage: "1Gi"
requests:
  memory: "750Mi"
  cpu: "500m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "100-nonha" }}
limits:
  cpu: "2500m"
  memory: "1250Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "1250m"
  memory: "1250Mi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "150-ha" }}
limits:
  cpu: 3000m
  memory: "1500Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: 1500m
  memory: "1500Mi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "250-ha" }}
limits:
  cpu: 4
  memory: "2Gi"
  ephemeral-storage: "1Gi"
requests:
  cpu: 2
  memory: "2Gi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "500-ha" }}
limits:
  cpu: 5000m
  memory: "2Gi"
  ephemeral-storage: "1Gi"
requests:
  cpu: 2500m
  memory: "2Gi"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: 5000m
  memory: "5000Mi"
  ephemeral-storage: "1Gi"
requests:
  cpu: 4000m
  memory: "5000Mi"
  ephemeral-storage: "1Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate UI replicaCount
Usage: {{ include "common.sizing.stackstate.ui.replicaCount" . }}
*/}}
{{- define "common.sizing.stackstate.ui.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}1
{{- else }}2
{{- end }}
{{- end }}
{{- end }}

{{/*
Get kafkaTopicCreate extraEnv based on sizing profile
Usage: {{ include "common.sizing.stackstate.kafkaTopicCreate.extraEnv.open" . }}
*/}}
{{- define "common.sizing.stackstate.kafkaTopicCreate.extraEnv.open" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
KAFKA_PARTITIONS_sts_correlated_connections: "40"
KAFKA_PARTITIONS_sts_correlate_endpoints: "40"
KAFKA_PARTITIONS_sts_correlate_http_trace_observations: "40"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get components.all extraEnv based on sizing profile
Usage: {{ include "common.sizing.stackstate.all.extraEnv.open" . }}
*/}}
{{- define "common.sizing.stackstate.all.extraEnv.open" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}
CONFIG_FORCE_stackstate_topologyQueryService_maxStackElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_topologyQueryService_maxLoadedElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_agents_agentLimit: "10"
CONFIG_FORCE_stackgraph_retentionWindowMs: "259200000"
CONFIG_FORCE_stackstate_traces_retentionDays: "3"
{{- else if eq $profile "150-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "150"
{{- else if eq $profile "250-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "250"
{{- else if eq $profile "500-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "500"
{{- else if eq $profile "4000-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "4000"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get receiver env based on sizing profile
Usage: {{ include "common.sizing.stackstate.receiver.env" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
CONFIG_FORCE_akka_http_host__connection__pool_max__open__requests: "384"
CONFIG_FORCE_stackstate_receiver_countBufferSize: "3000000"
CONFIG_FORCE_zstd__decompress__dispatcher_fork__join__executor_parallelism__factor: "4.0"
CONFIG_FORCE_zstd__decompress__dispatcher_fork__join__executor_parallelism__max: "64"
CONFIG_FORCE_akka_actor_default__dispatcher_fork__join__executor_parallelism__min: "16"
CONFIG_FORCE_akka_actor_default__dispatcher_fork__join__executor_parallelism__factor: "4.0"
CONFIG_FORCE_akka_actor_default__dispatcher_fork__join__executor_parallelism__max: "64"
CONFIG_FORCE_akka_actor_default__blocking__io__dispatcher_thread__pool__executor_fixed__pool__size: "64"
CONFIG_FORCE_stackstate_receiver_kafkaProducerConfig_max_request_size: "4194304"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get correlate env based on sizing profile
Usage: {{ include "common.sizing.stackstate.correlate.env" . }}
*/}}
{{- define "common.sizing.stackstate.correlate.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "250-ha" }}
CONFIG_FORCE_stackstate_correlate_correlateConnections_workers: "2"
CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers: "2"
CONFIG_FORCE_stackstate_correlate_aggregation_workers: "2"
{{- else if or (eq $profile "500-ha") (eq $profile "4000-ha") }}
CONFIG_FORCE_stackstate_correlate_correlateConnections_workers: "4"
CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers: "4"
CONFIG_FORCE_stackstate_correlate_aggregation_workers: "4"
{{- if eq $profile "4000-ha" }}
CONFIG_FORCE_stackstate_correlate_correlateConnections_kafka_producer_request.timeout.ms: "25000"
CONFIG_FORCE_stackstate_correlate_correlateConnections_kafka_producer_delivery.timeout.ms: "30000"
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Get healthSync env based on sizing profile
Usage: {{ include "common.sizing.stackstate.healthSync.env" . }}
*/}}
{{- define "common.sizing.stackstate.healthSync.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
CONFIG_FORCE_stackstate_healthSync_initializationTimeout: "30 minutes"
CONFIG_FORCE_stackstate_healthSync_maxIdentifiersProcessingDelay: "30 minutes"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get UI component resources
Usage: {{ include "common.sizing.stackstate.ui.resources" . }}
*/}}
{{- define "common.sizing.stackstate.ui.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
requests:
  cpu: 100m
limits:
  cpu: 500m
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate e2es retention
Usage: {{ include "common.sizing.stackstate.e2es.retention" . }}
*/}}
{{- define "common.sizing.stackstate.e2es.retention" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate correlate extraEnv for Kafka topic partitions
Usage: {{ include "common.sizing.stackstate.correlate.kafkaPartitions.extraEnv.open" . }}
*/}}
{{- define "common.sizing.stackstate.correlate.kafkaPartitions.extraEnv.open" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
KAFKA_PARTITIONS_sts_correlated_connections: "40"
KAFKA_PARTITIONS_sts_correlate_endpoints: "40"
KAFKA_PARTITIONS_sts_correlate_http_trace_observations: "40"
{{- end }}
{{- end }}
{{- end }}
