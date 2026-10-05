{{/*
Resource identities retain existing naming expressions. Keep uname separate for
selectors, affinity, cluster identity and legacy runtime node-prefix matching.
*/}}
{{- define "elasticsearch.statefulset.fullname" -}}
{{ include "elasticsearch.uname" . }}
{{- end -}}

{{- define "elasticsearch.service.fullname" -}}
{{- if eq .Values.nodeGroup "master" -}}
{{ include "elasticsearch.masterService" . }}
{{- else -}}
{{ include "elasticsearch.uname" . }}
{{- end -}}
{{- end -}}

{{- define "elasticsearch.headless.service.fullname" -}}
{{- if eq .Values.nodeGroup "master" -}}
{{ include "elasticsearch.masterService" . }}-headless
{{- else -}}
{{ include "elasticsearch.uname" . }}-headless
{{- end -}}
{{- end -}}

{{- define "elasticsearch.config.configmap.fullname" -}}
{{ include "elasticsearch.uname" . }}-config
{{- end -}}

{{- define "elasticsearch.credentials.secret.fullname" -}}
{{ include "elasticsearch.uname" . }}-credentials
{{- end -}}

{{- define "elasticsearch.certificates.secret.fullname" -}}
{{ include "elasticsearch.uname" . }}-certs
{{- end -}}

{{- define "elasticsearch.serviceaccount.fullname" -}}
{{ include "elasticsearch.uname" . }}
{{- end -}}

{{- define "elasticsearch.role.fullname" -}}
{{ include "elasticsearch.uname" . }}
{{- end -}}

{{- define "elasticsearch.rolebinding.fullname" -}}
{{ include "elasticsearch.uname" . }}
{{- end -}}

{{- define "elasticsearch.podsecuritypolicy.fullname" -}}
{{ default (include "elasticsearch.uname" .) .Values.podSecurityPolicy.name }}
{{- end -}}

{{- define "elasticsearch.pull.secret.fullname" -}}
{{ include "elasticsearch.uname" . }}-pull-secret
{{- end -}}

{{- define "elasticsearch.pdb.fullname" -}}
{{ include "elasticsearch.uname" . }}-pdb
{{- end -}}

{{- define "elasticsearch.ingress.fullname" -}}
{{ include "elasticsearch.uname" . }}
{{- end -}}

{{- define "elasticsearch.test.pod.fullname" -}}
{{ .Release.Name }}-{{ randAlpha 5 | lower }}-test
{{- end -}}
