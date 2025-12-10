HBASE TEMPLATES
=============================================================================
*/}}

{{/*
Get hbase deployment mode based on sizing profile
Usage: {{ include "common.sizing.hbase.deploymentMode" . }}
*/}}
{{- define "common.sizing.hbase.deploymentMode" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}Mono
{{- else}}Distributed
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase stackgraph persistence size (for Mono mode)
Usage: {{ include "common.sizing.hbase.stackgraph.persistence.size" . }}
*/}}
{{- define "common.sizing.hbase.stackgraph.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}20Gi
{{- else if eq $profile "10-nonha" }}50Gi
{{- else if eq $profile "20-nonha" }}100Gi
{{- else if eq $profile "50-nonha" }}150Gi
{{- else if eq $profile "100-nonha" }}200Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase stackgraph resources (for Mono mode)
Usage: {{ include "common.sizing.hbase.stackgraph.resources" . }}
*/}}
{{- define "common.sizing.hbase.stackgraph.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}
requests:
  memory: "2250Mi"
  cpu: "500m"
limits:
  cpu: "1500m"
  memory: "2250Mi"
{{- else if or (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
requests:
  memory: "2250Mi"
  cpu: "1000m"
limits:
  cpu: "3000m"
  memory: "2250Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase console resources
Usage: {{ include "common.sizing.hbase.console.resources" . }}
*/}}
{{- define "common.sizing.hbase.console.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "20m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase master resources
Usage: {{ include "common.sizing.hbase.master.resources" . }}
*/}}
{{- define "common.sizing.hbase.master.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "500m"
limits:
  cpu: "1000m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase regionserver resources
Usage: {{ include "common.sizing.hbase.regionserver.resources" . }}
*/}}
{{- define "common.sizing.hbase.regionserver.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") }}
requests:
  cpu: "2000m"
limits:
  cpu: "4000m"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "3000m"
limits:
  cpu: "6000m"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "4000m"
limits:
  cpu: "8000m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase hdfs datanode resources
Usage: {{ include "common.sizing.hbase.hdfs.datanode.resources" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.datanode.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "600m"
limits:
  cpu: "1200m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase hdfs namenode resources
Usage: {{ include "common.sizing.hbase.hdfs.namenode.resources" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.namenode.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "200m"
limits:
  cpu: "400m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase hdfs secondarynamenode enabled flag
Usage: {{ include "common.sizing.hbase.hdfs.secondarynamenode.enabled" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.secondarynamenode.enabled" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}true
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase hdfs secondarynamenode resources
Usage: {{ include "common.sizing.hbase.hdfs.secondarynamenode.resources" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.secondarynamenode.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") (eq $profile "4000-ha") }}
requests:
  cpu: "10m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase tephra resources
Usage: {{ include "common.sizing.hbase.tephra.resources" . }}
*/}}
{{- define "common.sizing.hbase.tephra.resources" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}
limits:
  cpu: "100m"
  memory: "512Mi"
requests:
  memory: "512Mi"
  cpu: "50m"
{{- else if or (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
limits:
  memory: "1Gi"
  cpu: "200m"
requests:
  memory: "1Gi"
  cpu: "100m"
{{- else if or (eq $profile "150-ha") (eq $profile "250-ha") }}
limits:
  memory: "1Gi"
  cpu: "1000m"
requests:
  memory: "1Gi"
  cpu: "500m"
{{- else if eq $profile "500-ha" }}
limits:
  memory: "1500Mi"
  cpu: "1500m"
requests:
  memory: "1500Mi"
  cpu: "750m"
{{- else if eq $profile "4000-ha" }}
limits:
  memory: "2Gi"
  cpu: "2000m"
requests:
  memory: "2Gi"
  cpu: "1000m"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase tephra replicaCount
Usage: {{ include "common.sizing.hbase.tephra.replicaCount" . }}
*/}}
{{- define "common.sizing.hbase.tephra.replicaCount" -}}
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
