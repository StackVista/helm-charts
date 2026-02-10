{{/*
Sizing-related helper templates for HBase chart.
These helpers determine configuration based on sizing profiles when available,
falling back to values.yaml defaults.
*/}}

{{/*
Get HBase deployment mode with sizing profile evaluation.

Usage: {{ include "hbase.deployment.mode" . }}
Returns: "Distributed" or "Mono"
*/}}
{{- define "hbase.deployment.mode" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profileMode := include "common.sizing.hbase.deployment.mode" . | trim -}}
{{- if $profileMode -}}
{{- $profileMode -}}
{{- else -}}
{{- .Values.deployment.mode | default "Distributed" -}}
{{- end -}}
{{- else -}}
{{- .Values.deployment.mode | default "Distributed" -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase stackgraph persistence size with sizing profile evaluation.

Usage: {{ include "hbase.stackgraph.persistence.size" . }}
Returns: Storage size (e.g., "50Gi")
*/}}
{{- define "hbase.stackgraph.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profileSize := include "common.sizing.hbase.stackgraph.persistence.size" . | trim -}}
{{- if $profileSize -}}
{{- $profileSize -}}
{{- else -}}
{{- .Values.stackgraph.persistence.size -}}
{{- end -}}
{{- else -}}
{{- .Values.stackgraph.persistence.size -}}
{{- end -}}
{{- end -}}

{{/*
Get HBase tephra persistence size with sizing profile evaluation.

Usage: {{ include "hbase.tephra.persistence.size" . }}
Returns: Storage size (e.g., "50Gi")
*/}}
{{- define "hbase.tephra.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profileSize := include "common.sizing.hbase.tephra.persistence.size" . | trim -}}
{{- if $profileSize -}}
{{- $profileSize -}}
{{- else -}}
{{- .Values.tephra.persistence.size -}}
{{- end -}}
{{- else -}}
{{- .Values.tephra.persistence.size -}}
{{- end -}}
{{- end -}}

{{/*
Get HDFS datanode persistence size with sizing profile evaluation.

Usage: {{ include "hbase.hdfs.datanode.persistence.size" . }}
Returns: Storage size (e.g., "250Gi")
*/}}
{{- define "hbase.hdfs.datanode.persistence.size" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profileSize := include "common.sizing.hbase.hdfs.datanode.persistence.size" . | trim -}}
{{- if $profileSize -}}
{{- $profileSize -}}
{{- else -}}
{{- .Values.hdfs.datanode.persistence.size -}}
{{- end -}}
{{- else -}}
{{- .Values.hdfs.datanode.persistence.size -}}
{{- end -}}
{{- end -}}
