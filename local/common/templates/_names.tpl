{{/*
Resource names shared by SUSE Observability and its subcharts.
These do not change the generic common.name or common.fullname helpers.
*/}}
{{- define "suse-observability.resourcePrefix" -}}
suse-observability
{{- end -}}

{{- define "suse-observability.pullSecret.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-pull-secret
{{- end -}}

{{/*
The platform produces this endpoint ConfigMap; the collector consumes it.
Its identity is independent of the collector's fullnameOverride and workload
configuration ConfigMaps. Standalone collectors retain the same external reference.
*/}}
{{- define "stackstate.otelCollector.endpoints.configmap.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-otel-collector
{{- end -}}

{{- define "stackstate.clickhouse.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-clickhouse
{{- end -}}

{{/*
The backup configuration uses this fixed name independently of ClickHouse's fullnameOverride.
*/}}
{{- define "stackstate.backup.clickhouse.backup.service" -}}
{{ include "stackstate.clickhouse.fullname" . }}-backup
{{- end -}}

{{/*
HBase backup environment objects are produced by the platform and consumed by
HBase in both Mono and Distributed mode, even when backups are disabled.
Preserve Base=backup, global overrides/prefixes/suffixes and truncation exactly;
using the canonical prefix here would break existing StatefulSet references.
The helpers live in common so standalone HBase resolves the same legacy names.
*/}}
{{- define "stackstate.backup.hbase.configmap.fullname" -}}
{{ template "common.fullname.global" (merge (dict "Base" "backup") .) }}-sts-hbase-backup
{{- end -}}

{{- define "stackstate.backup.hbase.secret.fullname" -}}
{{ template "common.fullname.global" (merge (dict "Base" "backup") .) }}-sts-hbase-backup
{{- end -}}
