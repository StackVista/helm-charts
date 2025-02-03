{{- define "stackstate.rbacAgent.url" -}}
http://{{ template "common.fullname.short" . }}-router:8080/receiver/stsAgent
{{- end -}}

{{- define "stackstate.rbacAgent.roleName" -}}
{{ template "common.fullname.short" . }}-rbac-agent
{{- end -}}
