{{/*
Standalone fallback for the platform-managed registry Secret. The parent
overrides this interface so Elasticsearch needs no dependency on common.
*/}}
{{- define "elasticsearch.platformPullSecret.fullname" -}}
suse-observability-pull-secret
{{- end -}}
