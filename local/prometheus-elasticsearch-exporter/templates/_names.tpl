{{/*
The stateless Deployment and monitoring resources use the canonical product
name. Service DNS, RBAC, accounts and certificate Secrets retain their identities.
Generic name/fullname remain separate for labels, selectors, rule-group names
and customer tpl values.
*/}}
{{- define "elasticsearch-exporter.deployment.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-prometheus-elasticsearch-exporter
{{- end -}}

{{- define "elasticsearch-exporter.service.fullname" -}}
{{ include "elasticsearch-exporter.fullname" . }}
{{- end -}}

{{- define "elasticsearch-exporter.serviceaccount.fullname" -}}
{{ include "elasticsearch-exporter.fullname" . }}
{{- end -}}

{{- define "elasticsearch-exporter.role.fullname" -}}
{{ include "elasticsearch-exporter.fullname" . }}
{{- end -}}

{{- define "elasticsearch-exporter.rolebinding.fullname" -}}
{{ include "elasticsearch-exporter.fullname" . }}
{{- end -}}

{{- define "elasticsearch-exporter.podsecuritypolicy.fullname" -}}
{{ include "elasticsearch-exporter.fullname" . }}
{{- end -}}

{{- define "elasticsearch-exporter.certificates.secret.fullname" -}}
{{ include "elasticsearch-exporter.fullname" . }}-cert
{{- end -}}

{{- define "elasticsearch-exporter.servicemonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-prometheus-elasticsearch-exporter
{{- end -}}

{{- define "elasticsearch-exporter.podmonitor.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-prometheus-elasticsearch-exporter
{{- end -}}

{{- define "elasticsearch-exporter.prometheusrule.fullname" -}}
{{ include "suse-observability.resourcePrefix" . }}-prometheus-elasticsearch-exporter
{{- end -}}
