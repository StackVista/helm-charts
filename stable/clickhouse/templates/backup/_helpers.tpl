{{- define "clickhouse.ensureTrailingSlashIfNotEmpty" -}}
  {{- if . -}}
    {{- printf "%s/" (. | trimSuffix "/") -}}
  {{- else -}}
    {{- "" -}}
  {{- end -}}
{{- end -}}

{{- define "clickhouse.backup.service.name" -}}
{{ include "common.names.fullname" . }}-backup
{{- end -}}
