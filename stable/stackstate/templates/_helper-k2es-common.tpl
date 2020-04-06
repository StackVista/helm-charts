{{/*
Common container items for K2ES based deployments.
*/}}
{{- define "stackstate.k2es.deployment.common.initcontainer" -}}
name: k2es-init
command:
- sh
- -c
- |
  /entrypoint -c {{ include "stackstate.kafka.endpoint" . }},{{ include "stackstate.es.endpoint" . }} -t 300
image: "docker.io/dokkupaas/wait:latest"
imagePullPolicy: Always
{{- end -}}

{{- define "stackstate.k2es.deployment.common.container" -}}
env:
{{- include "stackstate.common.envvars" . }}
{{- include "stackstate.k2es.envvars" . }}
- name: CONFIG_FORCE_stackstate_kafkaGenericEventsToES_elasticsearch_index_replicas
  value: "1"
- name: CONFIG_FORCE_stackstate_kafkaMultiMetricsToES_elasticsearch_index_replicas
  value: "1"
- name: CONFIG_FORCE_stackstate_kafkaStateEventsToES_elasticsearch_index_replicas
  value: "1"
- name: CONFIG_FORCE_stackstate_kafkaStsEventsToES_elasticsearch_index_replicas
  value: "1"
- name: ELASTICSEARCH_URI
  value: "http://{{ include "stackstate.es.endpoint" . }}"
- name: KAFKA_BROKERS
  value: {{ include "stackstate.kafka.endpoint" . | quote }}
image: "{{ .Values.stackstate.components.k2es.image.repository }}:{{ default .Values.stackstate.components.all.image.tag .Values.stackstate.components.k2es.image.tag }}"
imagePullPolicy: {{ default .Values.stackstate.components.all.image.pullPolicy .Values.stackstate.components.k2es.image.pullPolicy | quote }}
livenessProbe:
  httpGet:
    path: /liveness
    port: health
  initialDelaySeconds: 60
  timeoutSeconds: 5
ports:
- containerPort: 1618
  name: health
- containerPort: 7070
  name: k2es
{{- if .Values.stackstate.components.all.metrics.enabled }}
- containerPort: 9404
  name: metrics
{{- end }}
readinessProbe:
  httpGet:
    path: /readiness
    port: health
  initialDelaySeconds: 60
  timeoutSeconds: 5
{{- with .Values.stackstate.components.k2es.resources }}
resources:
  {{- toYaml . | nindent 2 }}
{{- end }}
securityContext:
  runAsGroup: 65534
  runAsNonRoot: true
  runAsUser: 65534
{{- end -}}
