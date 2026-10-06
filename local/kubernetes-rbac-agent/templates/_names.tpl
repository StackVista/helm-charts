{{/*
Resource names retain the existing release/namespace truncation and suffixes.
Keep app.name and global.name separate for labels and selectors.
*/}}
{{- define "kubernetes-rbac-agent.deployment.fullname" -}}
{{ include "kubernetes-rbac-agent.app.name" . }}
{{- end -}}

{{- define "kubernetes-rbac-agent.serviceaccount.fullname" -}}
{{ include "kubernetes-rbac-agent.app.name" . }}
{{- end -}}

{{- define "kubernetes-rbac-agent.clusterrole.fullname" -}}
{{ include "kubernetes-rbac-agent.global.name" . }}
{{- end -}}

{{- define "kubernetes-rbac-agent.clusterrolebinding.fullname" -}}
{{ include "kubernetes-rbac-agent.global.name" . }}
{{- end -}}

{{- define "kubernetes-rbac-agent.api-key.secret.fullname" -}}
{{ include "kubernetes-rbac-agent.app.name" . }}-api-key
{{- end -}}

{{- define "kubernetes-rbac-agent.pull.secret.fullname" -}}
{{ include "kubernetes-rbac-agent.app.name" . }}-pull-secret
{{- end -}}

{{- define "kubernetes-rbac-agent.url.configmap.fullname" -}}
{{ include "kubernetes-rbac-agent.app.name" . }}-url
{{- end -}}

{{- define "kubernetes-rbac-agent.clusterName.configmap.fullname" -}}
{{ include "kubernetes-rbac-agent.app.name" . }}-cluster-name
{{- end -}}

{{- define "kubernetes-rbac-agent.customCertificates.configmap.fullname" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}-custom-certificates
{{- end -}}
