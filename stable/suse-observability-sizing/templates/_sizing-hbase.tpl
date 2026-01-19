{{/*
HBASE SIZING TEMPLATES
=============================================================================
*/}}

{{/*
Get hbase deployment mode based on sizing profile
Usage: {{ include "common.sizing.hbase.deployment.mode" . }}
*/}}
{{- define "common.sizing.hbase.deployment.mode" -}}
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
{{- else if eq $profile "20-nonha" }}50Gi
{{- else if eq $profile "50-nonha" }}100Gi
{{- else if eq $profile "100-nonha" }}100Gi
{{- else if eq $profile "150-ha" }}250Gi
{{- else if eq $profile "250-ha" }}250Gi
{{- else if eq $profile "500-ha" }}250Gi
{{- else if eq $profile "4000-ha" }}1000Gi
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
{{- else if eq $profile "10-nonha" -}}
requests:
  memory: "2250Mi"
  cpu: "500m"
limits:
  cpu: "1500m"
  memory: "2250Mi"
{{- else if eq $profile "20-nonha" -}}
requests:
  memory: "2500Mi"
  cpu: "1000m"
limits:
  cpu: "2000m"
  memory: "2500Mi"
{{- else if eq $profile "50-nonha" -}}
requests:
  memory: "3Gi"
  cpu: "2000m"
limits:
  cpu: "4000m"
  memory: "3Gi"
{{- else if eq $profile "100-nonha" -}}
requests:
  memory: "4500Mi"
  cpu: "2000m"
limits:
  cpu: "4000m"
  memory: "4500Mi"
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
{{- if eq $profile "250-ha" }}
requests:
  cpu: "50m"
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
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}
requests:
  cpu: "500m"
limits:
  cpu: "1000m"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "2000m"
  memory: "1024Mi"
limits:
  cpu: "2500m"
  memory: "1024Mi"
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
{{- if eq $profile "150-ha" }}
requests:
  cpu: "2000m"
limits:
  cpu: "4000m"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "3000m"
  memory: 4Gi
limits:
  cpu: "6000m"
  memory: 4Gi
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "4000m"
  memory: 6Gi
limits:
  cpu: "8000m"
  memory: 6Gi
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "6000m"
  memory: 10Gi
limits:
  cpu: "8000m"
  memory: 12Gi
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
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}
requests:
  cpu: "600m"
limits:
  cpu: "1200m"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "4000m"
  memory: "3Gi"
limits:
  cpu: "6000m"
  memory: "5Gi"
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
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}
requests:
  cpu: "200m"
limits:
  cpu: "400m"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "2000m"
  memory: "1024Mi"
limits:
  cpu: "4000m"
  memory: "2048Mi"
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
{{- if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}
requests:
  cpu: "10m"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "50m"
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
{{- else if eq $profile "10-nonha" }}
limits:
  cpu: "100m"
  memory: "512Mi"
requests:
  memory: "512Mi"
  cpu: "50m"
{{- else if eq $profile "20-nonha" }}
limits:
  cpu: "100m"
  memory: "600Mi"
requests:
  memory: "600Mi"
  cpu: "50m"
{{- else if eq $profile "50-nonha" }}
limits:
  cpu: "100m"
  memory: "750Mi"
requests:
  memory: "750Mi"
  cpu: "50m"
{{- else if eq $profile "100-nonha" }}
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
  memory: "1Gi"
  cpu: "1000m"
requests:
  memory: "1Gi"
  cpu: "500m"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: "6000m"
requests:
  cpu: "4000m"
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
Get hbase experimental.execLivenessProbe.enabled
Usage: {{ include "common.sizing.hbase.experimental.execLivenessProbe.enabled" . }}
*/}}
{{- define "common.sizing.hbase.experimental.execLivenessProbe.enabled" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}true
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase regionserver replicaCount
Usage: {{ include "common.sizing.hbase.regionserver.replicaCount" . }}
*/}}
{{- define "common.sizing.hbase.regionserver.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}5
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase hdfs datanode persistence size
Usage: {{ include "common.sizing.hbase.hdfs.datanode.persistence.size" . }}
*/}}
{{- define "common.sizing.hbase.hdfs.datanode.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}1000Gi
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase tephra persistence size (for Mono mode)
Usage: {{ include "common.sizing.hbase.tephra.persistence.size" . }}

NOTE: This template intentionally returns empty string for all profiles to use
the HBase chart's default value (1Gi), matching the behavior of the old profile YAML files.
*/}}
{{- define "common.sizing.hbase.tephra.persistence.size" -}}
{{- /* Intentionally empty - all profiles use HBase default (1Gi) */ -}}
{{- end }}

{{/*
Get hbase tephra archive max size MB
Usage: {{ include "common.sizing.hbase.tephra.archiveMaxSizeMb" . }}

NOTE: This helper intentionally returns empty for all profiles to use the
computed value (1% of disk space) in the hbase chart, matching the old profile behavior.
Only add explicit values here if a profile needs a value different from the computed default.
*/}}
{{- define "common.sizing.hbase.tephra.archiveMaxSizeMb" -}}
{{- /* Intentionally empty - let hbase chart compute from disk space */ -}}
{{- end }}

{{/*
Get hbase master replicaCount
Usage: {{ include "common.sizing.hbase.master.replicaCount" . }}
*/}}
{{- define "common.sizing.hbase.master.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- /* All profiles use 1 master - only needed for distributed mode */ -}}
1
{{- end }}
{{- end }}

{{/*
Get hdfs datanode replicaCount
Usage: {{ include "common.sizing.hdfs.datanode.replicaCount" . }}
*/}}
{{- define "common.sizing.hdfs.datanode.replicaCount" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- /* All profiles use 1 datanode - only needed for distributed mode */ -}}
1
{{- end }}
{{- end }}


{{/*
=============================================================================
*/}}
