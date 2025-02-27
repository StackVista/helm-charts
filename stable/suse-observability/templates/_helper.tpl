{{/*
Return the image registry
*/}}
{{- define "stackstate.image.registry" -}}
  {{- if ((.ServiceConfig).image).imageRegistry -}}
    {{- .ServiceConfig.image.imageRegistry -}}
  {{- else -}}
  {{ include "common.image.registry" ( dict "image" .Values.stackstate.components.all.image "context" $) }}
  {{- end -}}
{{- end -}}

{{/*
Return the image registry for the kafka-topic-create container
*/}}
{{- define "stackstate.kafkaTopicCreate.image.registry" -}}
{{ include "common.image.registry" ( dict "image" .Values.stackstate.components.kafkaTopicCreate.image "context" $) }}
{{- end -}}

{{/*
Return the image registry for the router-mode container
*/}}
{{- define "stackstate.router.mode.image.registry" -}}
{{ include "common.image.registry" ( dict "image" .Values.stackstate.components.router.mode.image "context" $) }}
{{- end -}}

{{/*
Return the image registry for the nginx-prometheus-operator container
*/}}
{{- define "stackstate.nginxPrometheusExporter.image.registry" -}}
{{ include "common.image.registry" ( dict "image" .Values.stackstate.components.nginxPrometheusExporter.image "context" $) }}
{{- end -}}

{{/*
Return the image registry for the router container
*/}}
{{- define "stackstate.router.image.registry" -}}
{{ include "common.image.registry" ( dict "image" .Values.stackstate.components.router.image "context" $) }}
{{- end -}}

{{/*
Return the image registry for the stackpacks containers
*/}}
{{- define "stackstate.stackpacks.image.registry" -}}
{{ include "common.image.registry" ( dict "image" .Values.stackstate.stackpacks.image "context" $) }}
{{- end -}}

{{/*
Return the image registry for the container-tools containers
*/}}
{{- define "stackstate.containerTools.image.registry" -}}
{{ include "common.image.registry" ( dict "image" .Values.stackstate.components.containerTools.image "context" $) }}
{{- end -}}

{{/*
Return the image registry for the wait containers
*/}}
{{- define "stackstate.wait.image.registry" -}}
{{ include "common.image.registry" ( dict "image" .Values.stackstate.components.wait.image "context" $) }}
{{- end -}}

{{/*
Return the image registry for the vmrestore containers
*/}}
{{- define "stackstate.vmrestore.image.registry" -}}
{{ include "common.image.registry" ( dict "image" (index .Values "victoria-metrics" "restore" "image") "context" $) }}
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
Environment variables containing the properly sanitized StackState Base URLs
*/}}
{{- define "stackstate.baseurls.envvars" }}
- name: STACKSTATE_BASE_URL
  value: {{ .Values.stackstate.baseUrl | default .Values.stackstate.receiver.baseUrl | trimSuffix "/" | required "stackstate.baseUrl is required" | quote }}
- name: RECEIVER_BASE_URL
  value: {{ printf "%s/%s" ( .Values.stackstate.baseUrl | default .Values.stackstate.receiver.baseUrl | trimSuffix "/" | required "stackstate.baseUrl is required" ) "receiver" | quote }}
{{- end -}}

{{/*
Environment variables containing the properly sanitized StackState Base URLs
*/}}
{{- define "stackstate.metricstore.envvar" }}
- name: METRICSTORE_URI
  value: "{{ include "stackstate.metrics.query.url" . }}"
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
Authorization Service secret checksum annotations
*/}}
{{- define "stackstate.authorizationSync.secret.checksum" -}}
checksum/authorizationSync-env: {{ include (print $.Template.BasePath "/secret-authorizationSync.yaml") . | sha256sum }}
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
E2ES secret checksum annotations
*/}}
{{- define "stackstate.e2es.secret.checksum" -}}
{{- if .Values.stackstate.components.e2es.extraEnv.secret }}
checksum/e2es-env: {{ include (print $.Template.BasePath "/secret-e2es.yaml") . | sha256sum }}
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
License secret checksum annotations
*/}}
{{- define "stackstate.apiKey.secret.checksum" -}}
checksum/api-key-env: {{ include (print $.Template.BasePath "/secret-api-key.yaml") . | sha256sum }}
{{- end -}}

{{/*
License secret checksum annotations
*/}}
{{- define "stackstate.auth.secret.checksum" -}}
checksum/auth-env: {{ include (print $.Template.BasePath "/secret-auth.yaml") . | sha256sum }}
{{- end -}}

{{/*
Email secret checksum annotations
*/}}
{{- define "stackstate.email.secret.checksum" -}}
checksum/auth-env: {{ include (print $.Template.BasePath "/secret-email.yaml") . | sha256sum }}
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
Slicing secret checksum annotations
*/}}
{{- define "stackstate.slicing.secret.checksum" -}}
checksum/slicing-env: {{ include (print $.Template.BasePath "/secret-slicing.yaml") . | sha256sum }}
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
checksum/ui-env: {{ include (print $.Template.BasePath "/secret-ui.yaml") . | sha256sum }}
{{- end -}}

{{/*
HealthSync secret checksum annotations
*/}}
{{- define "stackstate.healthSync.secret.checksum" -}}
checksum/healthSync-env: {{ include (print $.Template.BasePath "/secret-healthSync.yaml") . | sha256sum }}
{{- end -}}

{{/*
Notification secret checksum annotations
*/}}
{{- define "stackstate.notification.secret.checksum" -}}
checksum/notification-env: {{ include (print $.Template.BasePath "/secret-notification.yaml") . | sha256sum }}
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
Correlate configmap checksum annotations
*/}}
{{- define "stackstate.correlate.configmap.checksum" -}}
checksum/correlate-configmap: {{ include (print $.Template.BasePath "/configmap-correlate.yaml") . | sha256sum }}
{{- end -}}

{{/*
E2ES configmap checksum annotations
*/}}
{{- define "stackstate.e2es.configmap.checksum" -}}
checksum/e2es-configmap: {{ include (print $.Template.BasePath "/configmap-e2es.yaml") . | sha256sum }}
{{- end -}}

{{/*
Initializer configmap checksum annotations
*/}}
{{- define "stackstate.initializer.configmap.checksum" -}}
checksum/initializer-configmap: {{ include (print $.Template.BasePath "/configmap-initializer.yaml") . | sha256sum }}
{{- end -}}

{{/*
Receiver configmap checksum annotations
*/}}
{{- define "stackstate.receiver.configmap.checksum" -}}
checksum/receiver-configmap: {{ include (print $.Template.BasePath "/configmap-receiver.yaml") . | sha256sum }}
{{- end -}}

{{/*
Sync configmap checksum annotations
*/}}
{{- define "stackstate.sync.configmap.checksum" -}}
checksum/sync-configmap: {{ include (print $.Template.BasePath "/configmap-sync.yaml") . | sha256sum }}
{{- end -}}

{{/*
Slicing configmap checksum annotations
*/}}
{{- define "stackstate.slicing.configmap.checksum" -}}
checksum/slicing-configmap: {{ include (print $.Template.BasePath "/configmap-slicing.yaml") . | sha256sum }}
{{- end -}}

{{/*
State configmap checksum annotations
*/}}
{{- define "stackstate.state.configmap.checksum" -}}
checksum/state-configmap: {{ include (print $.Template.BasePath "/configmap-state.yaml") . | sha256sum }}
{{- end -}}

{{/*
HealthSync configmap checksum annotations
*/}}
{{- define "stackstate.healthSync.configmap.checksum" -}}
checksum/healthSync-configmap: {{ include (print $.Template.BasePath "/configmap-healthSync.yaml") . | sha256sum }}
{{- end -}}

{{/*
AuthorizationSync configmap checksum annotations
*/}}
{{- define "stackstate.authorizationSync.configmap.checksum" -}}
checksum/authorizationSync-configmap: {{ include (print $.Template.BasePath "/configmap-authorizationSync.yaml") . | sha256sum }}
{{- end -}}

{{/*
Notification configmap checksum annotations
*/}}
{{- define "stackstate.notification.configmap.checksum" -}}
checksum/notification-configmap: {{ include (print $.Template.BasePath "/configmap-notification.yaml") . | sha256sum }}
{{- end -}}


{{/*
Vmagent configmap checksum annotations
*/}}
{{- define "stackstate.vmagent.configmap.checksum" -}}
checksum/vmagent-configmap: {{ include (print $.Template.BasePath "/configmap-vmagent.yaml") . | sha256sum }}
{{- end -}}


{{/*
Ingress paths / routes
*/}}
{{- define "stackstate.ingress.rules" -}}
{{- $ctx := .global }}
{{- if .global.Values.ingress.hosts }}
  {{- range .global.Values.ingress.hosts }}
- host: {{ print $.hostPrefix (tpl .host $ctx) | quote }}
  http:
    paths:
      - path: {{ $.global.Values.ingress.path | quote }}
    {{- if $ctx.Capabilities.APIVersions.Has "networking.k8s.io/v1/Ingress" }}
        pathType: Prefix
        backend:
          service:
            name: {{ $.serviceName }}
            port:
              number: {{ $.port }}
    {{- else }}
        backend:
          serviceName: {{ $.serviceName }}
          servicePort: {{ $.port }}
    {{- end }}
  {{- end }}
{{- else }}
- http:
    paths:
      - path: {{ $.global.Values.ingress.path | quote }}
    {{- if $ctx.Capabilities.APIVersions.Has "networking.k8s.io/v1/Ingress" }}
        pathType: Prefix
        backend:
          service:
            name: {{ $.serviceName }}
            port:
              number: {{ $.port }}
    {{- else }}
        backend:
          serviceName: {{ $.serviceName }}
          servicePort: {{ $.port }}
    {{- end }}
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

{{/*
Renders a value that contains a template.
Usage:
{{ include "stackstate.tplvalue.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "stackstate.tplvalue.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}

{{/*
Return the proper Docker Image Registry Secret Names evaluating values as templates
{{ include "stackstate.image.pullSecret.name" ( dict "images" (list .Values.path.to.the.image1, .Values.path.to.the.image2) "context" $) }}
*/}}
{{- define "stackstate.image.pullSecret.name" -}}
  {{- $pullSecrets := list }}
  {{- $context := .context }}

  {{- if $context.Values.global }}
    {{- range $context.Values.global.imagePullSecrets -}}
      {{- $pullSecrets = append $pullSecrets (include "stackstate.tplvalue.render" (dict "value" . "context" $context)) -}}
    {{- end -}}
  {{- end -}}
  {{- range $context.Values.imagePullSecrets -}}
    {{- $pullSecrets = append $pullSecrets (include "stackstate.tplvalue.render" (dict "value" .name "context" $context)) -}}
  {{- end -}}

  {{- if (not (empty $pullSecrets)) }}
imagePullSecrets:
    {{- range $pullSecrets | uniq }}
  - name: {{ . }}
    {{- end }}
  {{- end }}
{{- end -}}


{{- define "stackstate.service.spec.poddisruptionbudget" -}}
metadata:
  name: {{ template "common.fullname.short" . }}-{{ .PdbName }}
  labels:
    app.kubernetes.io/component: {{ .PdbName }}
spec:
  selector:
    matchLabels:
      app.kubernetes.io/component: {{ .PdbName }}
{{- with .PdbBudget }}
  {{ toYaml . | nindent 2 }}
{{- end }}
{{- end -}}

{{- define "stackstate.service.poddisruptionbudget" -}}
{{- $commonPdb := fromYaml (include "common.poddisruptionbudget" .) -}}
{{- $sserviceSpecPdb := fromYaml (include "stackstate.service.spec.poddisruptionbudget" .) -}}
{{- toYaml (merge $sserviceSpecPdb $commonPdb) -}}
{{- end -}}

{{/*
Logic validate the total shares of Es disk
*/}}
{{- define "stackstate.elastic.storage.total" -}}
{{- $receiverDs := .Values.stackstate.components.receiver.esDiskSpaceShare | int -}}
{{- $eventsDs := .Values.stackstate.components.e2es.esDiskSpaceShare | int -}}
{{- $total := (add $receiverDs $eventsDs) | int -}}
{{- if ne $total 100 }}
{{- fail "The share of ElasticSearch disk on receiver.esDiskSpaceShare, e2es.esDiskSpaceShare should be 100." }}
{{- end }}
{{- end -}}

{{- define "stackstate.storage.to.megabytes" -}}
{{- if hasSuffix "Ti" . -}}
    {{- $ti := trimSuffix "Ti" . | int -}}
    {{- mul $ti 1100000 -}}
{{- else if hasSuffix "T" . -}}
    {{- $t := trimSuffix "T" . | int -}}
    {{- mul $t 1000000 -}}
{{- else if hasSuffix "Gi" . -}}
    {{- $gi := trimSuffix "Gi" . | int -}}
    {{- mulf $gi 1073.74 -}}
{{- else if hasSuffix "G" . -}}
    {{- $g := trimSuffix "G" . | int -}}
    {{- mul $g 1000 -}}
{{- else if hasSuffix "Mi" . -}}
    {{- $mi := trimSuffix "Mi" . | int -}}
    {{- mulf $mi 1.048576 -}}
{{- else if hasSuffix "M" . -}}
    {{- trimSuffix "M" . | int -}}
{{- else -}}
    {{- if regexMatch "^[0-9]*$" . -}}
        {{ . | int }}
    {{- end -}}
{{- end -}}
{{- end -}}

{{- define "stackstate.cpu_resource.to.cpu_core" -}}
{{ $cpu := . | toString}}
{{- if hasSuffix "m" $cpu -}}
    {{- $ti := trimSuffix "m" $cpu | int -}}
    {{- floor (div $ti 1000) -}}
{{- else -}}
    {{- if regexMatch "^[0-9]*$" $cpu -}}
        {{ $cpu | int }}
    {{- end -}}
{{- end -}}
{{- end -}}

{{/*
Determines the hostname prefix for the different stackstate services. This name is stable across subcharts
*/}}
{{- define "stackstate.hostname.prefix" -}}
{{- template "common.fullname.global" (merge (dict "Base" "suse-observability") .) }}
{{- end -}}

{{/*
Determines the hostname fr the router. This name is stable across subcharts
*/}}
{{- define "stackstate.router.name" -}}
{{- template "stackstate.hostname.prefix" . }}-router
{{- end -}}

{{/*
Ensure pods created by the StatefulSet that was responsible for the server pods in the 4.1.0 release
StackState are no longer running.

N.B.: The first invocation of kubectl get pod is not a mistake. It is there to cause the pod to fail
if that command fails because of a misconfiguration error. The while loop would just exit succesfully
in that case.
*/}}
{{- define "stackstate.initContainer.ensure.no.server.statefulset.pods.are.running" -}}
name: ensure-no-server-statefulset-pods-are-running
image: "{{include "stackstate.containerTools.image.registry" .}}/{{ .Values.stackstate.components.containerTools.image.repository }}:{{ .Values.stackstate.components.containerTools.image.tag }}"
imagePullPolicy: {{ .Values.stackstate.components.containerTools.image.pullPolicy | quote }}
command:
- '/bin/bash'
- '-c'
- 'kubectl get pod {{ template "common.fullname.short" . }}-server-0 --ignore-not-found && while (kubectl get pod {{ template "common.fullname.short" . }}-server-0 ) ; do echo "Waiting for {{ template "common.fullname.short" . }}-server-0 pod to terminate"; sleep 1; done'
{{- end -}}

{{/*
Clean up the directory containing the transaction logs.
*/}}
{{- define "stackstate.initContainer.cleanTransactionLogsDirectory" -}}
name: clean-transaction-logs-directory
image: "{{include "stackstate.containerTools.image.registry" .}}/{{ .Values.stackstate.components.containerTools.image.repository }}:{{ .Values.stackstate.components.containerTools.image.tag }}"
imagePullPolicy: {{ .Values.stackstate.components.containerTools.image.pullPolicy | quote }}
command:
- '/bin/bash'
- '-c'
- 'rm -Rf /opt/docker/logs/*'
volumeMounts:
{{ include "stackstate.service.transactionLog.volumeMount" . }}
{{- end -}}

{{/*
Clean up the /tmp directory, some deployments has mounted PV to /tmp directory, so we have to clean it before each restart
*/}}
{{- define "stackstate.initContainer.cleanTmpDirectory" -}}
name: clean-tmp-directory
image: "{{include "stackstate.containerTools.image.registry" .}}/{{ .Values.stackstate.components.containerTools.image.repository }}:{{ .Values.stackstate.components.containerTools.image.tag }}"
imagePullPolicy: {{ .Values.stackstate.components.containerTools.image.pullPolicy | quote }}
command:
- '/bin/bash'
- '-c'
- 'rm -Rf /tmp/*'
volumeMounts:
  - name: tmp-volume
    mountPath: /tmp
{{- end -}}

{{/*
Init container to load stackpacks from docker image
*/}}
{{- define "stackstate.initContainer.stackpacks" -}}
name: init-stackpacks
{{- $deploymentMode := .Values.stackstate.stackpacks.image.deploymentModeOverride | default .Values.stackstate.deployment.mode | lower -}}
{{- $tag := printf "%s-%s-%s" .Values.stackstate.stackpacks.image.version (lower .Values.stackstate.deployment.edition ) $deploymentMode }}
image: "{{ include "stackstate.stackpacks.image.registry" . }}/{{ .Values.stackstate.stackpacks.image.repository }}:{{ $tag }}"
imagePullPolicy: {{ .Values.stackstate.stackpacks.image.pullPolicy | quote }}
args: ["/var/stackpacks"]
volumeMounts:
{{ include "stackstate.stackpacks.volumeMount" . }}
{{- end -}}

{{/*
Returns a YAML with extra annotations for StackState service, it contains annotations for "all" and service provided within "Name" parameter
*/}}
{{- define "stackstate.component.podExtraAnnotations" -}}
{{- with .Values.stackstate.components.all.podAnnotations }}
{{- toYaml . | nindent 8}}
{{- end }}
{{- with (index .Values.stackstate.components .Name "podAnnotations") }}
{{- toYaml . | nindent 8}}
{{- end }}
{{- end -}}

{{/*
Return env entries to mount existing secret with the custom keys.
*/}}
{{- define "stackstate.component.envsFromExistingSecrets" -}}
{{- range . }}
- name: {{ .name }}
  valueFrom:
    secretKeyRef:
      name: {{ .secretName }}
      key: {{ .secretKey }}
{{- end -}}
{{- end -}}

{{/*
Return ttlSecondsAfterFinished. We make this a very high value for argo so failures cannot be silently ignored.
*/}}
{{- define "stackstate.job.ttlSecondsAfterFinished" -}}
{{- if .Values.deployment.compatibleWithArgoCD }}86400{{- else }}600{{- end -}}
{{- end -}}
