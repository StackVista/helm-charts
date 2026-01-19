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
Usage: {{ include "common.sizing.clickhouse.replicaCount" . }}
Returns: 1 for most profiles, 3 for 4000-ha, empty if no profile set
*/}}
{{- define "common.sizing.clickhouse.replicaCount" -}}
{{- $profileMap := dict "trial" "1" "10-nonha" "1" "20-nonha" "1" "50-nonha" "1" "100-nonha" "1" "150-ha" "1" "250-ha" "1" "500-ha" "1" "4000-ha" "3" -}}
{{- include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) -}}
{{- end }}

{{/*
Get clickhouse affinity
Usage: {{ include "common.sizing.clickhouse.affinity" . }}
*/}}
{{- define "common.sizing.clickhouse.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "clickhouse") "context" .) }}
{{- end }}
