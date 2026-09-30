{{/*
Standalone fallback for the platform-managed registry Secret. The parent
overrides this interface so the RBAC agent needs no dependency on common.
*/}}
{{- define "kubernetes-rbac-agent.platformPullSecret.fullname" -}}
suse-observability-pull-secret
{{- end -}}
