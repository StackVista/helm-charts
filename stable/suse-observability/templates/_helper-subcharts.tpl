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
