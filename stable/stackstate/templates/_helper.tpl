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
Api extra environment variables for api pods inherited through `stackstate.components.api.extraEnv`
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
Checks extra environment variables for state pods inherited through `stackstate.components.checks.extraEnv`
*/}}
{{- define "stackstate.checks.envvars" -}}
{{- if .Values.stackstate.components.checks.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.checks.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.checks.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.checks.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-checks
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Api extra environment variables for initializer pods inherited through `stackstate.components.api.extraEnv`
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
Sync extra environment variables for sync pods inherited through `stackstate.components.sync.extraEnv`
*/}}
{{- define "stackstate.sync.envvars" -}}
{{- if .Values.stackstate.components.sync.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.sync.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.sync.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.sync.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-sync
      key: {{ $key }}
  {{- end }}
{{- end }}
{{- end -}}

{{/*
Environment variables to enable authentication with safe defaults
*/}}
{{- define "stackstate.authentication.envvars" -}}
{{- if or .Values.stackstate.authentication.adminPassword .Values.stackstate.components.server.extraEnv.secret.CONFIG_FORCE_stackstate_api_authentication_authServer_stackstateAuthServer_defaultPassword }}
- name: CONFIG_FORCE_stackstate_adminApi_authentication_authServer_stackstateAuthServer_defaultPassword
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-common
      key: adminApiPassword
{{- end }}
- name: CONFIG_FORCE_stackstate_instanceApi_authentication_authServer_stackstateAuthServer_defaultPassword
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-common
      key: adminApiPassword
- name: CONFIG_FORCE_stackstate_api_authentication_authServer_stackstateAuthServer_defaultPassword
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-common
      key: adminPassword
{{- if hasKey .Values.stackstate.authentication.ldap "bind" }}
- name: LDAP_BIND_DN
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" . }}-common
      key: ldapBindDn
- name: LDAP_BIND_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" . }}-common
      key: ldapBindPassword
{{- end }}
{{- if .Values.stackstate.java.trustStorePassword }}
- name: JAVA_TRUSTSTORE_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" . }}-common
      key: javaTrustStorePassword
{{- end }}
{{- end -}}

{{/*
Mount secrets for custom certificates
*/}}
{{- define "stackstate.mountsecrets" -}}
{{- $mountSecrets := dict }}
{{- if hasKey .Values.stackstate.authentication.ldap "ssl" }}
  {{- if .Values.stackstate.authentication.ldap.ssl.trustCertificates }}
    {{- $_ := set $mountSecrets "ldapTrustCertificates" "ldap-certificates.pem" }}
  {{- end }}
  {{- if .Values.stackstate.authentication.ldap.ssl.trustStore }}
    {{- $_ := set $mountSecrets "ldapTrustStore" "ldap-cacerts" }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.java.trustStore }}
    {{- $_ := set $mountSecrets "javaTrustStore" "java-cacerts" }}
{{- end }}
{{ $mountSecrets | toYaml }}
{{- end -}}

{{/*
State extra environment variables for state pods inherited through `stackstate.components.state.extraEnv`
*/}}
{{- define "stackstate.state.envvars" -}}
{{- if .Values.stackstate.components.state.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.state.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.state.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.state.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-state
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
ViewHealth extra environment variables for viewHealth pods inherited through `stackstate.components.state.extraEnv`
*/}}
{{- define "stackstate.viewHealth.envvars" -}}
{{- if .Values.stackstate.components.viewHealth.extraEnv.open }}
  {{- range $key, $value := .Values.stackstate.components.viewHealth.extraEnv.open  }}
- name: {{ $key }}
  value: {{ $value | quote }}
  {{- end }}
{{- end }}
{{- if .Values.stackstate.components.viewHealth.extraEnv.secret }}
  {{- range $key, $value := .Values.stackstate.components.viewHealth.extraEnv.secret  }}
- name: {{ $key }}
  valueFrom:
    secretKeyRef:
      name: {{ template "common.fullname.short" $ }}-viewHealth
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
Checks secret checksum annotations
*/}}
{{- define "stackstate.checks.secret.checksum" -}}
checksum/checks-env: {{ include (print $.Template.BasePath "/secret-checks.yaml") . | sha256sum }}
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
Sync secret checksum annotations
*/}}
{{- define "stackstate.sync.secret.checksum" -}}
checksum/sync-env: {{ include (print $.Template.BasePath "/secret-sync.yaml") . | sha256sum }}
{{- end -}}

{{/*
State secret checksum annotations
*/}}
{{- define "stackstate.state.secret.checksum" -}}
checksum/state-env: {{ include (print $.Template.BasePath "/secret-state.yaml") . | sha256sum }}
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
ViewHealth secret checksum annotations
*/}}
{{- define "stackstate.viewHealth.secret.checksum" -}}
{{- if .Values.stackstate.components.viewHealth.extraEnv.secret }}
checksum/viewHealth-env: {{ include (print $.Template.BasePath "/secret-viewHealth.yaml") . | sha256sum }}
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
Checks configmap checksum annotations
*/}}
{{- define "stackstate.checks.configmap.checksum" -}}
checksum/checks-configmap: {{ include (print $.Template.BasePath "/configmap-checks.yaml") . | sha256sum }}
{{- end -}}

{{/*
Initializer configmap checksum annotations
*/}}
{{- define "stackstate.initializer.configmap.checksum" -}}
checksum/initializer-configmap: {{ include (print $.Template.BasePath "/configmap-initializer.yaml") . | sha256sum }}
{{- end -}}

{{/*
Sync configmap checksum annotations
*/}}
{{- define "stackstate.sync.configmap.checksum" -}}
checksum/sync-configmap: {{ include (print $.Template.BasePath "/configmap-sync.yaml") . | sha256sum }}
{{- end -}}

{{/*
State configmap checksum annotations
*/}}
{{- define "stackstate.state.configmap.checksum" -}}
checksum/state-configmap: {{ include (print $.Template.BasePath "/configmap-state.yaml") . | sha256sum }}
{{- end -}}

{{/*
ViewHealth configmap checksum annotations
*/}}
{{- define "stackstate.viewHealth.configmap.checksum" -}}
checksum/viewHealth-configmap: {{ include (print $.Template.BasePath "/configmap-viewHealth.yaml") . | sha256sum }}
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

{{- define "stackstate.image.pullSecret.name" -}}
{{- if .Values.stackstate.components.all.image.pullSecretName }}
imagePullSecrets:
- name: '{{ .Values.stackstate.components.all.image.pullSecretName }}'
{{- else if .Values.stackstate.components.all.image.pullSecretUsername }}
imagePullSecrets:
- name: '{{ template "common.fullname.short" . }}-pull-secret'
{{- end }}
{{- end -}}
