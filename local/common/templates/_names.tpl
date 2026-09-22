{{/*
Canonical names shared by SUSE Observability and its subcharts.
These do not change the generic common.name or common.fullname helpers.
*/}}
{{- define "suse-observability.resourcePrefix" -}}
suse-observability
{{- end -}}

{{- define "suse-observability.pullSecret.name" -}}
{{ include "suse-observability.resourcePrefix" . }}-pull-secret
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
