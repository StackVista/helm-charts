{{/*
Resource identities for the collector. These retain the existing fullname rules,
overrides and suffixes. Resource helpers are independent even when names match.
Changing the StatefulSet name, serviceName or claim templates requires a separate
storage/upgrade plan. User-provided volumeClaimTemplates remain untouched.
*/}}
{{- define "opentelemetry-collector.deployment.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.statefulset.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.daemonset.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}-agent
{{- end -}}

{{- define "opentelemetry-collector.service.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.grpc.service.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}-grpc
{{- end -}}

{{- define "opentelemetry-collector.deployment.configmap.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.daemonset.configmap.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}-agent
{{- end -}}

{{- define "opentelemetry-collector.statefulset.configmap.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}-statefulset
{{- end -}}

{{- define "opentelemetry-collector.horizontalpodautoscaler.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.poddisruptionbudget.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.podmonitor.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}-agent
{{- end -}}

{{- define "opentelemetry-collector.servicemonitor.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.prometheusrule.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.networkpolicy.fullname" -}}
{{ include "opentelemetry-collector.fullname" . }}
{{- end -}}

{{- define "opentelemetry-collector.serviceaccount.fullname" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "opentelemetry-collector.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "opentelemetry-collector.clusterrole.fullname" -}}
{{- default (include "opentelemetry-collector.fullname" .) .Values.clusterRole.name -}}
{{- end -}}

{{- define "opentelemetry-collector.clusterrolebinding.fullname" -}}
{{- default (include "opentelemetry-collector.fullname" .) .Values.clusterRole.clusterRoleBinding.name -}}
{{- end -}}

{{/* Ingress and route helpers receive the root context and optional entry name. */}}
{{- define "opentelemetry-collector.ingress.fullname" -}}
{{- include "opentelemetry-collector.fullname" .context -}}{{- with .name -}}-{{ . }}{{- end -}}
{{- end -}}

{{- define "opentelemetry-collector.httproute.fullname" -}}
{{- include "opentelemetry-collector.fullname" .context -}}{{- with .name -}}-{{ . }}{{- end -}}
{{- end -}}

{{- define "opentelemetry-collector.grpcroute.fullname" -}}
{{- include "opentelemetry-collector.fullname" .context -}}{{- with .name -}}-{{ . }}{{- end -}}
{{- end -}}

{{- define "opentelemetry-collector.autoscaling.target.fullname" -}}
{{- if eq .Values.mode "statefulset" -}}
{{- include "opentelemetry-collector.statefulset.fullname" . -}}
{{- else -}}
{{- include "opentelemetry-collector.deployment.fullname" . -}}
{{- end -}}
{{- end -}}

{{/* Preserve the broad legacy prefix used to exclude the collector's own logs. */}}
{{- define "opentelemetry-collector.logsCollection.podNamePrefix" -}}
{{- if eq .Values.mode "daemonset" -}}
{{- include "opentelemetry-collector.daemonset.fullname" . | trimSuffix "-agent" -}}
{{- else if eq .Values.mode "statefulset" -}}
{{- include "opentelemetry-collector.statefulset.fullname" . -}}
{{- else -}}
{{- include "opentelemetry-collector.deployment.fullname" . -}}
{{- end -}}
{{- end -}}
