{{- define "stackstate.rbacAgent.url" -}}
http://{{ printf "%s-suse-observability" .Release.Name }}-router:8080/receiver/stsAgent
{{- end -}}

{{- define "stackstate.rbacAgent.roleName" -}}
{{ template "common.fullname.short" . }}-rbac-agent
{{- end -}}
