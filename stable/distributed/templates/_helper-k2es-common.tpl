{{/*
Common container items for K2ES based deployments.
*/}}
{{- define "distributed.k2es.deployment.common.initcontainer" -}}
name: k2es-init
command:
- sh
- -c
- |
  /entrypoint -c {{ include "distributed.kafka.endpoint" . }},{{ include "distributed.es.endpoint" . }} -t 300
image: "docker.io/dokkupaas/wait:latest"
imagePullPolicy: Always
{{- end -}}

{{- define "distributed.k2es.deployment.common.container" -}}
env:
{{- include "distributed.common.envvars" . }}
{{- include "distributed.k2es.envvars" . }}
- name: CONFIG_FORCE_stackstate_kafkaGenericEventsToES_elasticsearch_index_replicas
  value: "1"
- name: CONFIG_FORCE_stackstate_kafkaMultiMetricsToES_elasticsearch_index_replicas
  value: "1"
- name: CONFIG_FORCE_stackstate_kafkaStateEventsToES_elasticsearch_index_replicas
  value: "1"
- name: CONFIG_FORCE_stackstate_kafkaStsEventsToES_elasticsearch_index_replicas
  value: "1"
- name: CONFIG_FORCE_stackstate_kafkaToEs_healthEndpointsEnabled
  value: "true"
- name: ELASTICSEARCH_URI
  value: "http://{{ include "distributed.es.endpoint" . }}"
- name: KAFKA_BROKERS
  value: {{ include "distributed.kafka.endpoint" . | quote }}
image: "{{ .Values.stackstate.components.k2es.image.repository }}:{{ default .Values.stackstate.components.all.image.tag .Values.stackstate.components.k2es.image.tag }}"
imagePullPolicy: {{ default .Values.stackstate.components.all.image.pullPolicy .Values.stackstate.components.k2es.image.pullPolicy | quote }}
livenessProbe:
  httpGet:
    path: /liveness
    port: k2es
  initialDelaySeconds: 90
  timeoutSeconds: 5
ports:
- containerPort: 9404
  name: monitoring
- containerPort: 7070
  name: k2es
readinessProbe:
  httpGet:
    path: /readiness
    port: k2es
  initialDelaySeconds: 90
  timeoutSeconds: 5
resources:
{{- with .Values.stackstate.components.k2es.resources }}
  {{- toYaml . | nindent 2 }}
{{- end }}
securityContext:
  runAsGroup: 65534
  runAsNonRoot: true
  runAsUser: 65534
{{- end -}}
