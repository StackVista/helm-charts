{{/*
Backup connection interface, overridable by a parent chart.
Consumers must resolve this helper with fromYaml into $configuration and use
$configuration.endpoint and $configuration.secretName. Direct reads of
.Values.backup.s3 outside this helper bypass the parent's configuration.
The template guard test checks this. The endpoint excludes the URL scheme;
the backup ConfigMap adds http://.
*/}}
{{- define "clickhouse.backup.connection" -}}
endpoint: {{ tpl .Values.backup.s3.endpoint . | quote }}
secretName: {{ tpl .Values.backup.s3.secretName . | quote }}
{{- end -}}
