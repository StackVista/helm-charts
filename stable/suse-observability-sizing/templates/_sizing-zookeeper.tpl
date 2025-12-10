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
Get zookeeper affinity
Usage: {{ include "common.sizing.zookeeper.affinity" . }}
*/}}
{{- define "common.sizing.zookeeper.affinity" -}}
{{- include "common.sizing.podAntiAffinity" (dict "labels" (dict "app.kubernetes.io/component" "zookeeper") "context" .) }}
{{- end }}
