
{{- define "kafkaup-operator.checksum-configs" }}
checksum/configmap: {{ include (print $.Template.BasePath "/kafkaup-configmap.yaml") . | sha256sum }}
{{- end }}
