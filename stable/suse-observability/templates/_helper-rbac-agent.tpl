{{- define "stackstate.rbacAgent.url" -}}
{{ template "stackstate.router.endpoint" . }}/receiver/stsAgent
{{- end -}}

{{/*
Legacy application authorization subject used by the API/server staticSubjects
configuration. Kubernetes resources use dedicated helpers in _names.tpl.
Preserve this identity during extraction. A future Kubernetes Role rename must
explicitly review its relationship with this application authorization mapping.
*/}}
{{- define "stackstate.rbacAgent.roleName" -}}
{{ template "common.fullname.short" . }}-rbac-agent
{{- end -}}
