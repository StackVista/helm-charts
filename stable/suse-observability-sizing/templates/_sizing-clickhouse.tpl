Get clickhouse resources based on sizing profile
Usage: {{ include "common.sizing.clickhouse.resources" . }}
*/}}
{{- define "common.sizing.clickhouse.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") (eq $profile "150-ha") (eq $profile "250-ha") }}
requests:
  cpu: "2"
  memory: "6Gi"
limits:
  cpu: "4"
  memory: "6Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "2"
  memory: "8Gi"
limits:
  cpu: "4"
  memory: "8Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "4"
  memory: "12Gi"
limits:
  cpu: "8"
  memory: "12Gi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get clickhouse persistence size based on sizing profile
Usage: {{ include "common.sizing.clickhouse.persistence.size" . }}
*/}}
{{- define "common.sizing.clickhouse.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") }}20Gi
{{- else if or (eq $profile "20-nonha") (eq $profile "50-nonha") }}50Gi
{{- else if eq $profile "100-nonha" }}100Gi
{{- else if eq $profile "150-ha" }}100Gi
{{- else if eq $profile "250-ha" }}200Gi
{{- else if eq $profile "500-ha" }}500Gi
{{- else if eq $profile "4000-ha" }}1000Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get zookeeper resources based on sizing profile
Usage: {{ include "common.sizing.zookeeper.resources" . }}
*/}}
{{- define "common.sizing.zookeeper.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
requests:
  cpu: 100m
limits:
  cpu: 250m
{{- else if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: 200m
limits:
  cpu: 500m
{{- end }}
{{- end }}
{{- end }}

{{/*
Get zookeeper persistence size based on sizing profile
Usage: {{ include "common.sizing.zookeeper.persistence.size" . }}
*/}}
{{- define "common.sizing.zookeeper.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}8Gi
{{- end }}
{{- end }}

{{/*
=============================================================================

{{/*
Get clickhouse replicaCount
Usage: {{ include "common.sizing.clickhouse.replicaCount" . }}
*/}}
{{- define "common.sizing.clickhouse.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}1
{{- end }}
{{- end }}

{{/*
Get clickhouse affinity
Usage: {{ include "common.sizing.clickhouse.affinity" . }}
*/}}
{{- define "common.sizing.clickhouse.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "clickhouse") "context" .) }}
{{- end }}
