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
  ephemeral-storage: "1Mi"
limits:
  cpu: "1500m"
  memory: "2250Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "10-nonha" -}}
requests:
  memory: "2250Mi"
  cpu: "500m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "1500m"
  memory: "2250Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "20-nonha" -}}
requests:
  memory: "2500Mi"
  cpu: "1000m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "2000m"
  memory: "2500Mi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "50-nonha" -}}
requests:
  memory: "3Gi"
  cpu: "2000m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "4000m"
  memory: "3Gi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "100-nonha" -}}
requests:
  memory: "4500Mi"
  cpu: "2000m"
  ephemeral-storage: "1Mi"
limits:
  cpu: "4000m"
  memory: "4500Mi"
  ephemeral-storage: "1Gi"
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
  memory: "1Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "1000m"
  memory: "1Gi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "2000m"
  memory: "1024Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "2500m"
  memory: "1024Mi"
  ephemeral-storage: "1Gi"
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
  memory: "3Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "4000m"
  memory: "3Gi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "250-ha" }}
requests:
  cpu: "3000m"
  memory: 4Gi
  ephemeral-storage: "1Mi"
limits:
  cpu: "6000m"
  memory: 4Gi
  ephemeral-storage: "1Gi"
{{- else if eq $profile "500-ha" }}
requests:
  cpu: "4000m"
  memory: 6Gi
  ephemeral-storage: "1Mi"
limits:
  cpu: "8000m"
  memory: 6Gi
  ephemeral-storage: "1Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "6000m"
  memory: 10Gi
  ephemeral-storage: "1Mi"
limits:
  cpu: "8000m"
  memory: 12Gi
  ephemeral-storage: "1Gi"
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
  memory: "4Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "1200m"
  memory: "4Gi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "4000m"
  memory: "3Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "6000m"
  memory: "5Gi"
  ephemeral-storage: "1Gi"
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
  memory: "1Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "400m"
  memory: "1Gi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "2000m"
  memory: "1024Mi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "4000m"
  memory: "2048Mi"
  ephemeral-storage: "1Gi"
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
  memory: "1Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "500m"
  memory: "1Gi"
  ephemeral-storage: "1Gi"
{{- else if eq $profile "4000-ha" }}
requests:
  cpu: "50m"
  memory: "1Gi"
  ephemeral-storage: "1Mi"
limits:
  cpu: "500m"
  memory: "1Gi"
  ephemeral-storage: "1Gi"
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
  ephemeral-storage: "1Gi"
requests:
  memory: "512Mi"
  cpu: "50m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "10-nonha" }}
limits:
  cpu: "100m"
  memory: "512Mi"
  ephemeral-storage: "1Gi"
requests:
  memory: "512Mi"
  cpu: "50m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "20-nonha" }}
limits:
  cpu: "100m"
  memory: "600Mi"
  ephemeral-storage: "1Gi"
requests:
  memory: "600Mi"
  cpu: "50m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "50-nonha" }}
limits:
  cpu: "100m"
  memory: "750Mi"
  ephemeral-storage: "1Gi"
requests:
  memory: "750Mi"
  cpu: "50m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "100-nonha" }}
limits:
  memory: "1Gi"
  cpu: "200m"
  ephemeral-storage: "1Gi"
requests:
  memory: "1Gi"
  cpu: "100m"
  ephemeral-storage: "1Mi"
{{- else if or (eq $profile "150-ha") (eq $profile "250-ha") }}
limits:
  memory: "1Gi"
  cpu: "1000m"
  ephemeral-storage: "1Gi"
requests:
  memory: "1Gi"
  cpu: "500m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "500-ha" }}
limits:
  memory: "1Gi"
  cpu: "1000m"
  ephemeral-storage: "1Gi"
requests:
  memory: "1Gi"
  cpu: "500m"
  ephemeral-storage: "1Mi"
{{- else if eq $profile "4000-ha" }}
limits:
  cpu: "6000m"
  memory: "3Gi"
  ephemeral-storage: "1Gi"
requests:
  cpu: "4000m"
  memory: "3Gi"
  ephemeral-storage: "1Mi"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get hbase tephra replicaCount
Usage: {{ include "common.sizing.hbase.tephra.replicaCount" . }}
Returns: 1 for non-HA profiles, 2 for HA profiles, empty if no profile set
*/}}
{{- define "common.sizing.hbase.tephra.replicaCount" -}}
{{- $profileMap := dict "trial" "1" "10-nonha" "1" "20-nonha" "1" "50-nonha" "1" "100-nonha" "1" "150-ha" "2" "250-ha" "2" "500-ha" "2" "4000-ha" "2" -}}
{{- include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) -}}
{{- end }}

{{/*
Get effective hbase master replicaCount with backwards-compatibility support.
Usage: {{ include "common.sizing.hbase.master.effectiveReplicaCount" . }}
Returns: Resolved replica count (callers pipe to | int as needed)
*/}}
{{- define "common.sizing.hbase.master.effectiveReplicaCount" -}}
{{- $sizingReplicaCount := include "common.sizing.hbase.master.replicaCount" . | trim -}}
{{- include "common.sizing.effectiveReplicaCount" (dict "sizingReplicaCount" $sizingReplicaCount "chartDefault" "2" "valuesReplicaCount" .Values.hbase.master.replicaCount) -}}
{{- end }}

{{/*
Get effective hbase regionserver replicaCount with backwards-compatibility support.
Usage: {{ include "common.sizing.hbase.regionserver.effectiveReplicaCount" . }}
Returns: Resolved replica count (callers pipe to | int as needed)
*/}}
{{- define "common.sizing.hbase.regionserver.effectiveReplicaCount" -}}
{{- $sizingReplicaCount := include "common.sizing.hbase.regionserver.replicaCount" . | trim -}}
{{- include "common.sizing.effectiveReplicaCount" (dict "sizingReplicaCount" $sizingReplicaCount "chartDefault" "3" "valuesReplicaCount" .Values.hbase.regionserver.replicaCount) -}}
{{- end }}

{{/*
Get effective hbase tephra replicaCount with backwards-compatibility support.
Usage: {{ include "common.sizing.hbase.tephra.effectiveReplicaCount" . }}
Returns: Resolved replica count (callers pipe to | int as needed)
*/}}
{{- define "common.sizing.hbase.tephra.effectiveReplicaCount" -}}
{{- $sizingReplicaCount := include "common.sizing.hbase.tephra.replicaCount" . | trim -}}
{{- include "common.sizing.effectiveReplicaCount" (dict "sizingReplicaCount" $sizingReplicaCount "chartDefault" "2" "valuesReplicaCount" .Values.tephra.replicaCount) -}}
{{- end }}

{{/*
Get effective hdfs datanode replicaCount with backwards-compatibility support.
Usage: {{ include "common.sizing.hdfs.datanode.effectiveReplicaCount" . }}
Returns: Resolved replica count (callers pipe to | int as needed)
*/}}
{{- define "common.sizing.hdfs.datanode.effectiveReplicaCount" -}}
{{- $sizingReplicaCount := include "common.sizing.hdfs.datanode.replicaCount" . | trim -}}
{{- include "common.sizing.effectiveReplicaCount" (dict "sizingReplicaCount" $sizingReplicaCount "chartDefault" "3" "valuesReplicaCount" .Values.hdfs.datanode.replicaCount) -}}
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
Returns: 1 for non-HA profiles, 3 for standard HA profiles, 5 for 4000-ha, empty if no profile set
*/}}
{{- define "common.sizing.hbase.regionserver.replicaCount" -}}
{{- $profileMap := dict "trial" "1" "10-nonha" "1" "20-nonha" "1" "50-nonha" "1" "100-nonha" "1" "150-ha" "3" "250-ha" "3" "500-ha" "3" "4000-ha" "5" -}}
{{- include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) -}}
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
Returns: 1 for non-HA profiles, 2 for HA profiles, empty if no profile set
*/}}
{{- define "common.sizing.hbase.master.replicaCount" -}}
{{- $profileMap := dict "trial" "1" "10-nonha" "1" "20-nonha" "1" "50-nonha" "1" "100-nonha" "1" "150-ha" "2" "250-ha" "2" "500-ha" "2" "4000-ha" "2" -}}
{{- include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) -}}
{{- end }}

{{/*
Get hdfs datanode replicaCount
Usage: {{ include "common.sizing.hdfs.datanode.replicaCount" . }}
Returns: 1 for non-HA profiles, 3 for HA profiles, empty if no profile set
*/}}
{{- define "common.sizing.hdfs.datanode.replicaCount" -}}
{{- $profileMap := dict "trial" "1" "10-nonha" "1" "20-nonha" "1" "50-nonha" "1" "100-nonha" "1" "150-ha" "3" "250-ha" "3" "500-ha" "3" "4000-ha" "3" -}}
{{- include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) -}}
{{- end }}

{{/*
Get hdfs secondarynamenode replicaCount
Usage: {{ include "common.sizing.hdfs.secondarynamenode.replicaCount" . }}
Returns: 0 for non-HA profiles, 1 for HA profiles, empty if no profile set
*/}}
{{- define "common.sizing.hdfs.secondarynamenode.replicaCount" -}}
{{- $profileMap := dict "trial" "0" "10-nonha" "0" "20-nonha" "0" "50-nonha" "0" "100-nonha" "0" "150-ha" "1" "250-ha" "1" "500-ha" "1" "4000-ha" "1" -}}
{{- include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) -}}
{{- end }}


{{/*
=============================================================================
*/}}
