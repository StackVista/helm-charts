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
