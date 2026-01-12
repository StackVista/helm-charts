{{/*
CLICKHOUSE SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get clickhouse resources based on sizing profile
Usage: {{ include "common.sizing.clickhouse.resources" . }}
*/}}
{{- define "common.sizing.clickhouse.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- /* No profile-specific clickhouse resources for 500-ha */ -}}
{{- end }}
{{- end }}

{{/*
Get clickhouse persistence size based on sizing profile
Usage: {{ include "common.sizing.clickhouse.persistence.size" . }}
*/}}
{{- define "common.sizing.clickhouse.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}5Gi
{{- else if or (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}50Gi
{{- else if eq $profile "150-ha" }}100Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get clickhouse replicaCount based on sizing profile.
Returns 1 for all profiles except 4000-ha (which uses the default of 3 from values.yaml).
Usage: {{ include "common.sizing.clickhouse.replicaCount" . }}
*/}}
{{- define "common.sizing.clickhouse.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if ne $profile "4000-ha" }}1
{{- end }}
{{- end }}
{{- end }}

{{/*
Get clickhouse affinity
Usage: {{ include "common.sizing.clickhouse.affinity" . }}
*/}}
{{- define "common.sizing.clickhouse.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "clickhouse") "context" .) }}
{{- end }}
