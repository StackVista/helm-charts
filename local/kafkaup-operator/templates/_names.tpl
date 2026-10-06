{{/*
Preserve the existing release-dependent identities. Keep selectors and Kafka
selection settings separate from these resource names.
*/}}
{{- define "kafkaup-operator.deployment.fullname" -}}
{{ include "common.fullname.short" . }}-kafkaup
{{- end -}}

{{- define "kafkaup-operator.serviceaccount.fullname" -}}
{{ include "common.fullname.short" . }}-kafkaup
{{- end -}}

{{- define "kafkaup-operator.role.fullname" -}}
{{ include "common.fullname.short" . }}-kafkaup
{{- end -}}

{{- define "kafkaup-operator.rolebinding.fullname" -}}
{{ include "common.fullname.short" . }}-kafkaup-binding
{{- end -}}

{{- define "kafkaup-operator.configmap.fullname" -}}
{{ include "common.fullname.short" . }}-kafkaup-config
{{- end -}}
