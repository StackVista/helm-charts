{{/*
=============================================================================
HBase Affinity Configuration
These templates provide complete affinity configurations for HBase components.
=============================================================================
*/}}

{{/*
Get complete HBase master affinity configuration for HA profiles.
Returns a complete affinity block with podAntiAffinity for HA profiles, empty for non-HA.

Usage in values.yaml:
hbase:
  hbase:
    master:
      affinity: {{ include "common.sizing.hbase.master.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.master.affinityConfig" -}}
{{- include "common.sizing.hbase.master.affinity" . -}}
{{- end -}}

{{/*
Get complete HBase regionserver affinity configuration for HA profiles.

Usage in values.yaml:
hbase:
  hbase:
    regionserver:
      affinity: {{ include "common.sizing.hbase.regionserver.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.regionserver.affinityConfig" -}}
{{- include "common.sizing.hbase.regionserver.affinity" . -}}
{{- end -}}

{{/*
Get complete HBase tephra affinity configuration for HA profiles.

Usage in values.yaml:
hbase:
  tephra:
    affinity: {{ include "common.sizing.hbase.tephra.affinityConfig" . | nindent 6 }}
*/}}
{{- define "common.sizing.hbase.tephra.affinityConfig" -}}
{{- include "common.sizing.hbase.tephra.affinity" . -}}
{{- end -}}

{{/*
Get complete HBase HDFS datanode affinity configuration for HA profiles.

Usage in values.yaml:
hbase:
  hdfs:
    datanode:
      affinity: {{ include "common.sizing.hbase.hdfs.datanode.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.hdfs.datanode.affinityConfig" -}}
{{- include "common.sizing.hbase.hdfs.datanode.affinity" . -}}
{{- end -}}

{{/*
Get complete HBase HDFS namenode affinity configuration for HA profiles.

Usage in values.yaml:
hbase:
  hdfs:
    namenode:
      affinity: {{ include "common.sizing.hbase.hdfs.namenode.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.hdfs.namenode.affinityConfig" -}}
{{- include "common.sizing.hbase.hdfs.namenode.affinity" . -}}
{{- end -}}

{{/*
Get complete HBase HDFS secondarynamenode affinity configuration for HA profiles.

Usage in values.yaml:
hbase:
  hdfs:
    secondarynamenode:
      affinity: {{ include "common.sizing.hbase.hdfs.secondarynamenode.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.hdfs.secondarynamenode.affinityConfig" -}}
{{- include "common.sizing.hbase.hdfs.secondarynamenode.affinity" . -}}
{{- end -}}
