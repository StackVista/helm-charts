{{/*
Standalone fallback for the platform-managed registry Secret. The parent
overrides this interface so the collector needs no dependency on common.
*/}}
{{- define "opentelemetry-collector.platformPullSecret.fullname" -}}
suse-observability-pull-secret
{{- end -}}
