{{/*
=============================================================================
HBase Affinity Configuration
These templates provide complete affinity configurations for HBase components.
=============================================================================
*/}}

{{/*
Get complete HBase master affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles. Empty for non-HA profiles without nodeAffinity.

Usage in values.yaml:
hbase:
  hbase:
    master:
      affinity: {{ include "common.sizing.hbase.master.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.master.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.master.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete HBase regionserver affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles.

Usage in values.yaml:
hbase:
  hbase:
    regionserver:
      affinity: {{ include "common.sizing.hbase.regionserver.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.regionserver.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.regionserver.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete HBase tephra affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles.

Usage in values.yaml:
hbase:
  tephra:
    affinity: {{ include "common.sizing.hbase.tephra.affinityConfig" . | nindent 6 }}
*/}}
{{- define "common.sizing.hbase.tephra.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.tephra.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete HBase HDFS datanode affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles.

Usage in values.yaml:
hbase:
  hdfs:
    datanode:
      affinity: {{ include "common.sizing.hbase.hdfs.datanode.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.hdfs.datanode.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.hdfs.datanode.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete HBase HDFS namenode affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles.

Usage in values.yaml:
hbase:
  hdfs:
    namenode:
      affinity: {{ include "common.sizing.hbase.hdfs.namenode.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.hdfs.namenode.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.hdfs.namenode.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete HBase HDFS secondarynamenode affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles.

Usage in values.yaml:
hbase:
  hdfs:
    secondarynamenode:
      affinity: {{ include "common.sizing.hbase.hdfs.secondarynamenode.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.hdfs.secondarynamenode.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.hdfs.secondarynamenode.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete HBase console affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles.

Usage in values.yaml:
hbase:
  console:
    affinity: {{ include "common.sizing.hbase.console.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.console.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.console.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Get complete HBase stackgraph affinity configuration for HA profiles.
Returns a complete affinity block with nodeAffinity (from global config) and
podAntiAffinity for HA profiles.

Usage in values.yaml:
hbase:
  stackgraph:
    affinity: {{ include "common.sizing.hbase.stackgraph.affinityConfig" . | nindent 8 }}
*/}}
{{- define "common.sizing.hbase.stackgraph.affinityConfig" -}}
{{- $result := dict -}}
{{- $nodeAffinity := include "common.sizing.global.nodeAffinity" . | trim -}}
{{- if $nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" ($nodeAffinity | fromYaml) -}}
{{- end -}}
{{- $podAntiAffinity := include "common.sizing.hbase.stackgraph.affinity" . | trim -}}
{{- if $podAntiAffinity -}}
  {{- $parsed := $podAntiAffinity | fromYaml -}}
  {{- if $parsed.podAntiAffinity -}}
    {{- $_ := set $result "podAntiAffinity" $parsed.podAntiAffinity -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}
