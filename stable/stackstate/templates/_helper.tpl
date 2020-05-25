{{/*
Return the image registry
*/}}
{{- define "stackstate.image.registry" -}}
  {{- if .Values.global }}
    {{- default .Values.stackstate.components.all.image.registry .Values.global.imageRegistry -}}
  {{- else -}}
    {{- .Values.stackstate.components.all.image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Return the image registry for the kafka-topic-create container
*/}}
{{- define "stackstate.kafkaTopicCreate.image.registry" -}}
  {{- if .Values.global }}
    {{- default .Values.stackstate.components.kafkaTopicCreate.image.registry .Values.global.imageRegistry -}}
  {{- else -}}
    {{- .Values.stackstate.components.kafkaTopicCreate.image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Return the image registry for the nginx-prometheus-operator container
*/}}
{{- define "stackstate.nginxPrometheusExporter.image.registry" -}}
  {{- if .Values.global }}
    {{- default .Values.stackstate.components.nginxPrometheusExporter.image.registry .Values.global.imageRegistry -}}
  {{- else -}}
    {{- .Values.stackstate.components.nginxPrometheusExporter.image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Return the image registry for the router container
*/}}
{{- define "stackstate.router.image.registry" -}}
  {{- if .Values.global }}
    {{- default .Values.stackstate.components.router.image.registry .Values.global.imageRegistry -}}
  {{- else -}}
    {{- .Values.stackstate.components.router.image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Return the image registry for the wait containers
*/}}
{{- define "stackstate.wait.image.registry" -}}
  {{- if .Values.global }}
    {{- default .Values.stackstate.components.wait.image.registry .Values.global.imageRegistry -}}
  {{- else -}}
    {{- .Values.stackstate.components.wait.image.registry -}}
  {{- end -}}
{{- end -}}

{{/*
Common extra environment variables for all processes inherited through `stackstate.components.all.extraEnv`
*/}}
{{- define "stackstate.common.envvars" -}}
{{- if .Values.stackstate.components.all.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.all.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.all.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.all.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-common
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Correlate extra environment variables for correlate pods inherited through `stackstate.components.correlate.extraEnv`
*/}}
{{- define "stackstate.correlate.envvars" -}}
{{- if .Values.stackstate.components.correlate.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.correlate.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.correlate.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.correlate.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-correlate
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
K2ES extra environment variables for k2es pods inherited through `stackstate.components.k2es.extraEnv`
*/}}
{{- define "stackstate.k2es.envvars" -}}
{{- if .Values.stackstate.components.k2es.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.k2es.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.k2es.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.k2es.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-k2es
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Receiver extra environment variables for receiver pods inherited through `stackstate.components.receiver.extraEnv`
*/}}
{{- define "stackstate.receiver.envvars" -}}
{{- if .Values.stackstate.components.receiver.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.receiver.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.receiver.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.receiver.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-receiver
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Router extra environment variables for ui pods inherited through `stackstate.components.router.extraEnv`
*/}}
{{- define "stackstate.router.envvars" -}}
{{- if .Values.stackstate.components.router.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.router.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.router.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.router.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-router
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Server extra environment variables for server pods inherited through `stackstate.components.server.extraEnv`
*/}}
{{- define "stackstate.server.envvars" -}}
{{- if .Values.stackstate.components.server.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.server.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.server.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.server.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-server
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Api extra environment variables for server pods inherited through `stackstate.components.api.extraEnv`
*/}}
{{- define "stackstate.api.envvars" -}}
{{- if .Values.stackstate.components.api.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.api.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.api.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.api.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-api
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Api extra environment variables for server pods inherited through `stackstate.components.api.extraEnv`
*/}}
{{- define "stackstate.initializer.envvars" -}}
{{- if .Values.stackstate.components.initializer.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.initializer.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.initializer.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.initializer.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-initializer
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Processor extra environment variables for server pods inherited through `stackstate.components.api.extraEnv`
*/}}
{{- define "stackstate.processor.envvars" -}}
{{- if .Values.stackstate.components.processor.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.processor.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.processor.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.processor.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-processor
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
UI extra environment variables for ui pods inherited through `stackstate.components.ui.extraEnv`
*/}}
{{- define "stackstate.ui.envvars" -}}
{{- if .Values.stackstate.components.ui.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.ui.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.ui.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.ui.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-ui
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Common secret checksum annotations
*/}}
{{- define "stackstate.common.secret.checksum" -}}
{{- if .Values.stackstate.components.all.extraEnv.secret }}
checksum/common-env: {{ include (print $.Template.BasePath "/secret-common.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
Correlate secret checksum annotations
*/}}
{{- define "stackstate.correlate.secret.checksum" -}}
{{- if .Values.stackstate.components.correlate.extraEnv.secret }}
checksum/correlate-env: {{ include (print $.Template.BasePath "/secret-correlate.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
K2ES secret checksum annotations
*/}}
{{- define "stackstate.k2es.secret.checksum" -}}
{{- if .Values.stackstate.components.k2es.extraEnv.secret }}
checksum/k2es-env: {{ include (print $.Template.BasePath "/secret-k2es.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
Receiver secret checksum annotations
*/}}
{{- define "stackstate.receiver.secret.checksum" -}}
checksum/receiver-env: {{ include (print $.Template.BasePath "/secret-receiver.yaml") . | sha256sum }}
{{- end -}}

{{/*
Api secret checksum annotations
*/}}
{{- define "stackstate.api.secret.checksum" -}}
checksum/api-env: {{ include (print $.Template.BasePath "/secret-api.yaml") . | sha256sum }}
{{- end -}}

{{/*
License secret checksum annotations
*/}}
{{- define "stackstate.license.secret.checksum" -}}
checksum/license-env: {{ include (print $.Template.BasePath "/secret-license-key.yaml") . | sha256sum }}
{{- end -}}

{{/*
Initializer secret checksum annotations
*/}}
{{- define "stackstate.initializer.secret.checksum" -}}
checksum/initializer-env: {{ include (print $.Template.BasePath "/secret-initializer.yaml") . | sha256sum }}
{{- end -}}

{{/*
Processor secret checksum annotations
*/}}
{{- define "stackstate.processor.secret.checksum" -}}
checksum/processor-env: {{ include (print $.Template.BasePath "/secret-processor.yaml") . | sha256sum }}
{{- end -}}

{{/*
Server secret checksum annotations
*/}}
{{- define "stackstate.server.secret.checksum" -}}
checksum/server-env: {{ include (print $.Template.BasePath "/secret-server.yaml") . | sha256sum }}
{{- end -}}

{{/*
UI secret checksum annotations
*/}}
{{- define "stackstate.ui.secret.checksum" -}}
{{- if .Values.stackstate.components.ui.extraEnv.secret }}
checksum/ui-env: {{ include (print $.Template.BasePath "/secret-ui.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
Router configmap checksum annotations
*/}}
{{- define "stackstate.router.configmap.checksum" -}}
checksum/router-configmap: {{ include (print $.Template.BasePath "/configmap-router.yaml") . | sha256sum }}
{{- end -}}

{{/*
Server configmap checksum annotations
*/}}
{{- define "stackstate.server.configmap.checksum" -}}
checksum/server-configmap: {{ include (print $.Template.BasePath "/configmap-server.yaml") . | sha256sum }}
{{- end -}}

{{/*
Api configmap checksum annotations
*/}}
{{- define "stackstate.api.configmap.checksum" -}}
checksum/api-configmap: {{ include (print $.Template.BasePath "/configmap-api.yaml") . | sha256sum }}
{{- end -}}

{{/*
Initializer configmap checksum annotations
*/}}
{{- define "stackstate.initializer.configmap.checksum" -}}
checksum/initializer-configmap: {{ include (print $.Template.BasePath "/configmap-initializer.yaml") . | sha256sum }}
{{- end -}}

{{/*
Processor configmap checksum annotations
*/}}
{{- define "stackstate.processor.configmap.checksum" -}}
checksum/processor-configmap: {{ include (print $.Template.BasePath "/configmap-processor.yaml") . | sha256sum }}
{{- end -}}

{{/*
Ingress paths / routes
*/}}
{{- define "stackstate.ingress.rules" -}}
{{- if .Values.ingress.hosts }}
  {{- range .Values.ingress.hosts }}
- host: {{ .host | quote }}
  http:
    paths:
      - backend:
          serviceName: {{ include "common.fullname.short" $ }}-router
          servicePort: 8080
        path: /
  {{- end }}
{{- else }}
- http:
    paths:
      - backend:
          serviceName: {{ include "common.fullname.short" . }}-router
          servicePort: 8080
        path: /
{{- end }}
{{- end -}}
{{- define "stackstate.servicemonitor.extraLabels" -}}
podTargetLabels:
  - __meta_kubernetes_pod_label_app_kubernetes_io_name
  - __meta_kubernetes_pod_label_app_kubernetes_io_component
endpoints:
  - interval: "20s"
    path: /metrics
    port: metrics
    scheme: http
    relabelings:
    - sourceLabels: [__meta_kubernetes_pod_label_app_kubernetes_io_name]
      regex: (.+)
      targetLabel: app_name
      action: replace
    - sourceLabels: [__meta_kubernetes_pod_label_app_kubernetes_io_component]
      regex: (.+)
      targetLabel: app_component
      action: replace
{{- end -}}
