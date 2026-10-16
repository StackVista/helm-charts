{{/*
The stateless Deployment uses the canonical product name. Monitoring resources
use canonical names in the release namespace and legacy names elsewhere.
Service DNS, RBAC, accounts and certificate Secrets retain their identities.
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
{{ include "elasticsearch-exporter.monitoring.fullname" (dict "context" . "namespace" .Values.serviceMonitor.namespace) }}
{{- end -}}

{{- define "elasticsearch-exporter.podmonitor.fullname" -}}
{{ include "elasticsearch-exporter.monitoring.fullname" (dict "context" . "namespace" .Values.podMonitor.namespace) }}
{{- end -}}

{{- define "elasticsearch-exporter.prometheusrule.fullname" -}}
{{ include "elasticsearch-exporter.monitoring.fullname" (dict "context" . "namespace" .Values.prometheusRule.namespace) }}
{{- end -}}

{{/*
Preserve legacy names outside the release namespace so installations sharing a
monitoring namespace keep distinct identities. Each object has its own namespace.
*/}}
{{- define "elasticsearch-exporter.monitoring.fullname" -}}
{{- if ne (default .context.Release.Namespace .namespace) .context.Release.Namespace -}}
{{ include "elasticsearch-exporter.fullname" .context }}
{{- else -}}
{{ include "suse-observability.resourcePrefix" .context }}-prometheus-elasticsearch-exporter
{{- end -}}
{{- end -}}
