{{/*
Adds optional ENV VAR with API_KEY value, the env can be disabled and then only Ingestion API Keys are allowed
*/}}
{{- define "stackstate.deployment.optional.apiKey.env" }}
{{- if not (.Values.global.onlyIngestionApiKey) }}
- name: API_KEY
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" . }}-receiver
      key: sts-receiver-api-key
{{- end }}
{{- end -}}
