{{/*
ELASTICSEARCH SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get elasticsearch storage size based on sizing profile
Usage: {{ include "common.sizing.elasticsearch.storage" . }}
*/}}
{{- define "common.sizing.elasticsearch.storage" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}20Gi
{{- else if eq $profile "10-nonha" }}50Gi
{{- else if eq $profile "20-nonha" }}50Gi
{{- else if eq $profile "50-nonha" }}50Gi
{{- else if eq $profile "100-nonha" }}100Gi
{{- else if eq $profile "150-ha" }}200Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get elasticsearch resources based on sizing profile
Usage: {{ include "common.sizing.elasticsearch.resources" . }}
*/}}
{{- define "common.sizing.elasticsearch.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") }}
requests:
  cpu: 500m
  memory: 2500Mi
limits:
  cpu: 1000m
  memory: 2500Mi
{{- else if eq $profile "20-nonha" }}
requests:
  cpu: 500m
  memory: 3000Mi
limits:
  cpu: 1000m
  memory: 3000Mi
{{- else if eq $profile "50-nonha" }}
requests:
  cpu: 1000m
  memory: 4Gi
limits:
  cpu: 2000m
  memory: 4Gi
{{- else if eq $profile "100-nonha" }}
requests:
  cpu: 750m
  memory: 7Gi
limits:
  cpu: 1500m
  memory: 7Gi
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "4"
  memory: 4Gi
limits:
  cpu: "6"
  memory: 4Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get elasticsearch replicas based on sizing profile
Usage: {{ include "common.sizing.elasticsearch.replicas" . }}
*/}}
{{- define "common.sizing.elasticsearch.replicas" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}1
{{- else }}3
{{- end }}
{{- end }}
{{- end }}

{{/*
Get Elasticsearch anti-affinity strategy
Usage: {{ include "common.sizing.elasticsearch.antiAffinityConfig" . }}
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityConfig" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") -}}
hard
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get Elasticsearch anti-affinity topology key
Usage: {{ include "common.sizing.elasticsearch.antiAffinityTopologyKeyConfig" . }}
*/}}
{{- define "common.sizing.elasticsearch.antiAffinityTopologyKeyConfig" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") -}}
kubernetes.io/hostname
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get Elasticsearch esJavaOpts
Usage: {{ include "common.sizing.elasticsearch.esJavaOpts" . }}
*/}}
{{- define "common.sizing.elasticsearch.esJavaOpts" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") -}}
-Xmx1500m -Xms1500m -Des.allow_insecure_settings=true
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Get prometheus-elasticsearch-exporter resources based on sizing profile
Usage: {{ include "common.sizing.elasticsearch.prometheus-exporter.resources" . }}
*/}}
{{- define "common.sizing.elasticsearch.prometheus-exporter.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") }}
requests:
  cpu: "50m"
  memory: "50Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "50m"
  memory: "50Mi"
  ephemeral-storage: "1Gi"
{{- else }}
requests:
  cpu: "100m"
  memory: "100Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "100m"
  memory: "100Mi"
  ephemeral-storage: "1Gi"
{{- end }}
{{- end }}
{{- end }}
