{{- define "kafkaup.image.pullSecret.name" -}}
{{- if .Values.image.pullSecretName }}
imagePullSecrets:
- name: '{{ .Values.image.pullSecretName }}'
{{- else if .Values.image.pullSecretUsername }}
imagePullSecrets:
- name: '{{ template "common.fullname.short" . }}-pull-secret'
{{- end }}
{{- end -}}

{{- define "kafkaup-operator.checksum-configs" }}
checksum/configmap: {{ include (print $.Template.BasePath "/kafkaup-configmap.yaml") . | sha256sum }}
{{- end }}
