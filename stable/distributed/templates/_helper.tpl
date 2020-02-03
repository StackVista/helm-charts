{{/*
Common extra environment variables for all processes inherited through `stackstate.components.all.extraEnv`
*/}}
{{- define "distributed.common.envvars" -}}
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
      name: {{ template "common.fullname" $ }}-common
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Correlate extra environment variables for correlate pods inherited through `stackstate.components.correlate.extraEnv`
*/}}
{{- define "distributed.correlate.envvars" -}}
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
      name: {{ template "common.fullname" $ }}-correlate
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
K2ES extra environment variables for k2es pods inherited through `stackstate.components.k2es.extraEnv`
*/}}
{{- define "distributed.k2es.envvars" -}}
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
      name: {{ template "common.fullname" $ }}-k2es
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Receiver extra environment variables for receiver pods inherited through `stackstate.components.receiver.extraEnv`
*/}}
{{- define "distributed.receiver.envvars" -}}
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
      name: {{ template "common.fullname" $ }}-receiver
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Router extra environment variables for ui pods inherited through `stackstate.components.router.extraEnv`
*/}}
{{- define "distributed.router.envvars" -}}
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
      name: {{ template "common.fullname" $ }}-router
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Server extra environment variables for server pods inherited through `stackstate.components.server.extraEnv`
*/}}
{{- define "distributed.server.envvars" -}}
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
      name: {{ template "common.fullname" $ }}-server
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
UI extra environment variables for ui pods inherited through `stackstate.components.ui.extraEnv`
*/}}
{{- define "distributed.ui.envvars" -}}
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
      name: {{ template "common.fullname" $ }}-ui
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Common secret checksum annotations
*/}}
{{- define "distributed.common.secret.checksum" -}}
{{- if .Values.stackstate.components.all.extraEnv.secret }}
checksum/common-env: {{ include (print $.Template.BasePath "/secret-common.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
Correlate secret checksum annotations
*/}}
{{- define "distributed.correlate.secret.checksum" -}}
{{- if .Values.stackstate.components.correlate.extraEnv.secret }}
checksum/correlate-env: {{ include (print $.Template.BasePath "/secret-correlate.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
K2ES secret checksum annotations
*/}}
{{- define "distributed.k2es.secret.checksum" -}}
{{- if .Values.stackstate.components.k2es.extraEnv.secret }}
checksum/k2es-env: {{ include (print $.Template.BasePath "/secret-k2es.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
Receiver secret checksum annotations
*/}}
{{- define "distributed.receiver.secret.checksum" -}}
checksum/receiver-env: {{ include (print $.Template.BasePath "/secret-receiver.yaml") . | sha256sum }}
{{- end -}}

{{/*
Server secret checksum annotations
*/}}
{{- define "distributed.server.secret.checksum" -}}
checksum/server-env: {{ include (print $.Template.BasePath "/secret-server.yaml") . | sha256sum }}
{{- end -}}

{{/*
UI secret checksum annotations
*/}}
{{- define "distributed.ui.secret.checksum" -}}
{{- if .Values.stackstate.components.ui.extraEnv.secret }}
checksum/ui-env: {{ include (print $.Template.BasePath "/secret-ui.yaml") . | sha256sum }}
{{- end }}
{{- end -}}

{{/*
Router configmap checksum annotations
*/}}
{{- define "distributed.router.configmap.checksum" -}}
checksum/router-configmap: {{ include (print $.Template.BasePath "/configmap-router.yaml") . | sha256sum }}
{{- end -}}

{{/*
Ingress paths / routes
*/}}
{{- define "distributed.ingress.rules" -}}
{{- if .Values.ingress.hosts }}
  {{- range .Values.ingress.hosts }}
- host: {{ .host | quote }}
  http:
    paths:
      - backend:
          serviceName: {{ include "common.fullname" $ }}-server
          servicePort: 7071
        path: /admin/?(.*)
      - backend:
          serviceName: {{ include "common.fullname" $ }}-server
          servicePort: 7070
        path: /?(api/.*)
      - backend:
          serviceName: {{ include "common.fullname" $ }}-server
          servicePort: 7070
        path: /?(loginCallback)
      - backend:
          serviceName: {{ include "common.fullname" $ }}-server
          servicePort: 7070
        path: /?(loginInfo)
      - backend:
          serviceName: {{ include "common.fullname" $ }}-server
          servicePort: 7070
        path: /?(logout)
      - backend:
          serviceName: {{ include "common.fullname" $ }}-receiver
          servicePort: 7077
        path: /receiver/?(.*)
      - backend:
          serviceName: {{ include "common.fullname" $ }}-ui
          servicePort: 8080
        path: /?(.*)
  {{- end }}
{{- else }}
- http:
    paths:
      - backend:
          serviceName: {{ include "common.fullname" . }}-server
          servicePort: 7071
        path: /admin/?(.*)
      - backend:
          serviceName: {{ include "common.fullname" . }}-server
          servicePort: 7070
        path: /?(api/.*)
      - backend:
          serviceName: {{ include "common.fullname" . }}-server
          servicePort: 7070
        path: /?(loginCallback)
      - backend:
          serviceName: {{ include "common.fullname" . }}-server
          servicePort: 7070
        path: /?(loginInfo)
      - backend:
          serviceName: {{ include "common.fullname" . }}-server
          servicePort: 7070
        path: /?(logout)
      - backend:
          serviceName: {{ include "common.fullname" . }}-receiver
          servicePort: 7077
        path: /receiver/?(.*)
      - backend:
          serviceName: {{ include "common.fullname" . }}-ui
          servicePort: 8080
        path: /?(.*)
{{- end }}
{{- end -}}
{{- define "distributed.servicemonitor.extraLabels" -}}
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
