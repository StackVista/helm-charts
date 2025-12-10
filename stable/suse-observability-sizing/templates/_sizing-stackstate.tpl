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
{{- else if or (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
limits:
  ephemeral-storage: 5Gi
  cpu: 6000m
  memory: 7Gi
requests:
  cpu: 3000m
  memory: 7Gi
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
limits:
  cpu: "3000m"
  memory: "4Gi"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "2000m"
  memory: "4Gi"
limits:
  cpu: "4000m"
  memory: "4Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "3000m"
  memory: "5Gi"
limits:
  cpu: "6000m"
  memory: "5Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "4000m"
  memory: "6Gi"
limits:
  cpu: "8000m"
  memory: "6Gi"
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
limits:
  cpu: "6000m"
  memory: 5Gi
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "4000m"
  memory: 7Gi
limits:
  cpu: "8000m"
  memory: 7Gi
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "6000m"
  memory: 10Gi
limits:
  cpu: "12000m"
  memory: 10Gi
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "8000m"
  memory: 14Gi
limits:
  cpu: "16000m"
  memory: 14Gi
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
requests:
  memory: 5Gi
  cpu: "1000m"
{{- else if eq $profile "250-ha" }}
limits:
  memory: 7Gi
  cpu: "3000m"
requests:
  memory: 7Gi
  cpu: "1500m"
{{- else if eq $profile "500-ha" }}
limits:
  memory: 10Gi
  cpu: "4000m"
requests:
  memory: 10Gi
  cpu: "2000m"
{{- else if eq $profile "4000-ha" }}
limits:
  memory: 14Gi
  cpu: "6000m"
requests:
  memory: 14Gi
  cpu: "3000m"
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
{{- if eq $profile "150-ha" }}"1200Mi"
{{- else if eq $profile "250-ha" }}"1700Mi"
{{- else if eq $profile "500-ha" }}"2500Mi"
{{- else if eq $profile "4000-ha" }}"3500Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate initializer component resources
Usage: {{ include "common.sizing.stackstate.initializer.resources" . }}
*/}}
{{- define "common.sizing.stackstate.initializer.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "50m"
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
  cpu: 750m
  memory: "3500Mi"
limits:
  cpu: 1500m
  memory: "3500Mi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: 1000m
  memory: "4Gi"
limits:
  cpu: 2000m
  memory: "4Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: 1500m
  memory: "5Gi"
limits:
  cpu: 3000m
  memory: "5Gi"
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
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "120m"
limits:
  cpu: "240m"
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
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "300m"
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
{{- if eq $profile "150-ha" }}
limits:
  cpu: "4"
  memory: "4000Mi"
requests:
  cpu: "2"
  memory: "4000Mi"
{{- else if eq $profile "250-ha" }}
limits:
  cpu: "6"
  memory: "5500Mi"
requests:
  cpu: "3"
  memory: "5500Mi"
{{- else if eq $profile "500-ha" }}
limits:
  cpu: "8"
  memory: "7500Mi"
requests:
  cpu: "4"
  memory: "7500Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: "12"
  memory: "10Gi"
requests:
  cpu: "6"
  memory: "10Gi"
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
requests:
  cpu: "2000m"
  memory: "6500Mi"
{{- else if eq $profile "250-ha" }}
limits:
  cpu: "6000m"
  memory: "9Gi"
requests:
  cpu: "3000m"
  memory: "9Gi"
{{- else if eq $profile "500-ha" }}
limits:
  cpu: "8000m"
  memory: "12500Mi"
requests:
  cpu: "4000m"
  memory: "12500Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: "12000m"
  memory: "17500Mi"
requests:
  cpu: "6000m"
  memory: "17500Mi"
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
{{- else if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  memory: "512Mi"
  cpu: "50m"
limits:
  memory: "512Mi"
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
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}3
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
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}3
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
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
requests:
  memory: "1250Mi"
  cpu: "500m"
limits:
  cpu: "1000m"
  memory: "1250Mi"
{{- else if eq $profile "150-ha" }}
limits:
  cpu: 6
  memory: 3500Mi
requests:
  cpu: 3
  memory: 3500Mi
{{- else if eq $profile "250-ha" }}
limits:
  cpu: 8
  memory: 5Gi
requests:
  cpu: 4
  memory: 5Gi
{{- else if eq $profile "500-ha" }}
limits:
  cpu: 12
  memory: 7Gi
requests:
  cpu: 6
  memory: 7Gi
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: 16
  memory: 10Gi
requests:
  cpu: 8
  memory: 10Gi
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
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
requests:
  memory: "1000Mi"
  cpu: "1000m"
limits:
  memory: "1000Mi"
  cpu: "2000m"
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
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}3
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
  memory: "7500Mi"
  cpu: "6000m"
limits:
  memory: "7500Mi"
  cpu: "12000m"
{{- else if eq $profile "500-ha" }}
requests:
  memory: "10Gi"
  cpu: "8000m"
limits:
  memory: "10Gi"
  cpu: "16000m"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "14Gi"
  cpu: "12000m"
limits:
  memory: "14Gi"
  cpu: "24000m"
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
  memory: "3500Mi"
  cpu: "1500m"
limits:
  memory: "3500Mi"
  cpu: "3000m"
{{- else if eq $profile "500-ha" }}
requests:
  memory: "5Gi"
  cpu: "2000m"
limits:
  memory: "5Gi"
  cpu: "4000m"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "7Gi"
  cpu: "3000m"
limits:
  memory: "7Gi"
  cpu: "6000m"
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
  memory: "4Gi"
  cpu: "1500m"
limits:
  memory: "4Gi"
  cpu: "3000m"
{{- else if eq $profile "500-ha" }}
requests:
  memory: "5500Mi"
  cpu: "2000m"
limits:
  memory: "5500Mi"
  cpu: "4000m"
{{- else if eq $profile "4000-ha" }}
requests:
  memory: "8Gi"
  cpu: "3000m"
limits:
  memory: "8Gi"
  cpu: "6000m"
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
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
limits:
  memory: "640Mi"
requests:
  memory: "384Mi"
{{- else if eq $profile "150-ha" }}
limits:
  cpu: 3000m
  memory: "1500Mi"
requests:
  cpu: 1500m
  memory: "1500Mi"
{{- else if eq $profile "250-ha" }}
limits:
  cpu: 4000m
  memory: "2Gi"
requests:
  cpu: 2000m
  memory: "2Gi"
{{- else if eq $profile "500-ha" }}
limits:
  cpu: 6000m
  memory: "3Gi"
requests:
  cpu: 3000m
  memory: "3Gi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: 8000m
  memory: "4Gi"
requests:
  cpu: 4000m
  memory: "4Gi"
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
=============================================================================
*/}}
