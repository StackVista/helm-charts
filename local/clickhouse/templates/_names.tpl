{{/*
Backup resources retain their canonical names, independently of fullnameOverride.
The backup Service itself continues to use clickhouse.backup.service.name.
*/}}
{{- define "clickhouse.backup.scripts.fullname" -}}
{{ include "stackstate.backup.clickhouse.backup.service" . }}-scripts
{{- end -}}

{{- define "clickhouse.backup.cronjob.fullname" -}}
{{ include "stackstate.clickhouse.fullname" . }}-{{ .job_name }}
{{- end -}}

{{- define "clickhouse.backup.podName" -}}
{{ include "stackstate.clickhouse.fullname" . }}-shard0-0
{{- end -}}
