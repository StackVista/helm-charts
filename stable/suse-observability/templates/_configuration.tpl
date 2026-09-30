{{/* Overrides of the configuration interfaces defined by bundled subcharts. */}}

{{/*
The platform owns the bundled RBAC agent's ConfigMap connection, evaluated in the
subchart's context. Use this installation's router and release name, ignoring the
subchart's url and clusterName settings. Explicit global url/clusterName.fromSecret
settings take precedence in the consuming templates.
*/}}
{{- define "kubernetes-rbac-agent.connection" -}}
url:
  value: {{ include "stackstate.rbacAgent.url" . | quote }}
clusterName:
  value: {{ .Release.Name | quote }}
{{- end -}}

{{/*
The platform owns backup connections; storage backends and external credentials
are configured through the platform's S3Proxy settings. These helpers are
evaluated in the respective subchart contexts.
*/}}
{{- define "victoria-metrics.backup.connection" -}}
endpoint: {{ printf "http://%s" (include "stackstate.s3proxy.endpoint" .) | quote }}
secretName: {{ include "stackstate.s3proxy.secretName" . | quote }}
{{- end -}}

{{- define "clickhouse.backup.connection" -}}
endpoint: {{ include "stackstate.s3proxy.endpoint" . | quote }}
secretName: {{ include "stackstate.s3proxy.secretName" . | quote }}
{{- end -}}

{{/* Bundled anomaly detection always targets this installation's router. */}}
{{- define "anomaly-detection.stackstate.instance" -}}
{{- include "stackstate.router.endpoint" . -}}
{{- end -}}
