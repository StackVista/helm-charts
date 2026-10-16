{{/*
Use canonical monitor names within an installation's namespace. Preserve legacy
names in a shared monitoring namespace to avoid introducing name collisions.
Resolve namespace with the same value precedence as common.servicemonitor.
*/}}
{{- define "hbase.servicemonitor.resourcePrefix" -}}
{{- $common := .Values.common | default dict -}}
{{- $monitor := merge (deepCopy (.Values.servicemonitor | default dict)) ($common.servicemonitor | default dict) -}}
{{- if ne (default .Release.Namespace $monitor.namespace) .Release.Namespace -}}
{{ include "common.fullname.short" . }}
{{- else -}}
{{ include "suse-observability.resourcePrefix" . }}
{{- end -}}
{{- end -}}

{{- define "hbase.hbase.master.servicemonitor.fullname" -}}
{{ include "hbase.servicemonitor.resourcePrefix" . }}-hbase-master
{{- end -}}

{{- define "hbase.hbase.regionserver.servicemonitor.fullname" -}}
{{ include "hbase.servicemonitor.resourcePrefix" . }}-hbase-rs
{{- end -}}

{{- define "hbase.hdfs.namenode.servicemonitor.fullname" -}}
{{ include "hbase.servicemonitor.resourcePrefix" . }}-hdfs-nn
{{- end -}}

{{- define "hbase.hdfs.secondarynamenode.servicemonitor.fullname" -}}
{{ include "hbase.servicemonitor.resourcePrefix" . }}-hdfs-snn
{{- end -}}

{{- define "hbase.hdfs.datanode.servicemonitor.fullname" -}}
{{ include "hbase.servicemonitor.resourcePrefix" . }}-hdfs-dn
{{- end -}}

{{- define "hbase.tephra.servicemonitor.fullname" -}}
{{ include "hbase.servicemonitor.resourcePrefix" . }}-tephra
{{- end -}}

{{/* Probe scripts are independent of persistent controller and Service names. */}}
{{- define "hbase.hbase.regionserver.scripts.configmap.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-hbase-rs-scripts
{{- end -}}

{{- define "hbase.hdfs.datanode.scripts.configmap.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-hdfs-dn-scripts
{{- end -}}

{{/*
PDB names are independent of HBase's persistent workload and Service identities.
Preserve existing selectors, disruption limits and creation conditions.
*/}}
{{- define "hbase.hbase.master.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-hbase-master
{{- end -}}

{{- define "hbase.hbase.regionserver.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-hbase-rs
{{- end -}}

{{- define "hbase.hdfs.namenode.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-hdfs-nn
{{- end -}}

{{- define "hbase.hdfs.secondarynamenode.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-hdfs-snn
{{- end -}}

{{- define "hbase.hdfs.datanode.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-hdfs-dn
{{- end -}}

{{- define "hbase.tephra.poddisruptionbudget.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-tephra
{{- end -}}
