{{/*
Backup connection interface, overridable by a parent chart.
Consumers must resolve this helper with fromYaml into $configuration and use
$configuration.endpoint and $configuration.secretName. Direct reads of
.Values.backup.overrideS3Endpoint or .Values.backup.awsSecrets outside this
helper bypass the parent's configuration. The template guard test checks this.
*/}}
{{- define "victoria-metrics.backup.connection" -}}
endpoint: {{ tpl .Values.backup.overrideS3Endpoint . | quote }}
secretName: {{ tpl .Values.backup.awsSecrets . | quote }}
{{- end -}}
