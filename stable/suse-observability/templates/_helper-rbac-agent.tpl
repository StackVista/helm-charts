{{- define "stackstate.rbacAgent.url" -}}
{{ template "stackstate.router.endpoint" . }}/receiver/stsAgent
{{- end -}}

{{- define "stackstate.rbacAgent.roleName" -}}
{{ template "common.fullname.short" . }}-rbac-agent
{{- end -}}
