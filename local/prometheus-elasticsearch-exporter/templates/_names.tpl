{{/*
Standalone fallback for the platform-managed registry Secret. The parent
overrides this interface so the exporter needs no dependency on common.
*/}}
{{- define "elasticsearch-exporter.platformPullSecret.fullname" -}}
suse-observability-pull-secret
{{- end -}}
