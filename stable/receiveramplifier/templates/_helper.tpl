{{- define "receiveramplifier.image.pullSecret.name" -}}
{{- if .Values.receiveramplifier.image.pullSecretName -}}
imagePullSecrets:
- name: '{{ .Values.receiveramplifier.image.pullSecretName }}'
{{- else if .Values.receiveramplifier.image.pullSecretUsername -}}
imagePullSecrets:
- name: '{{ template "common.fullname.short" . }}-pull-secret'
{{- end -}}
{{- end -}}
