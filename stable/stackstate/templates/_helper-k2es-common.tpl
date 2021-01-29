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
image: "{{include "stackstate.wait.image.registry" .}}/{{ .Values.stackstate.components.wait.image.repository }}:{{ .Values.stackstate.components.wait.image.tag }}"
imagePullPolicy: {{ .Values.stackstate.components.wait.image.pullPolicy | quote }}
{{- end -}}

{{- define "stackstate.k2es.deployment.common.container" -}}
env:
{{- $serviceConfig := dict "ServiceName" .K2esName "ServiceConfig" .K2esConfig }}
{{- include "stackstate.service.envvars" (merge $serviceConfig .) }}
{{/*
    Currently we use a single replicationFactor config for all indices on ES, that works fine with calculating the available disk space
    and on the STS processes assigning diskSpaceWeights to each process. But if in the future we have need to configure different
    replicationFactors per index we will need to revisit and adapt the diskSpaceWeights login on STS
*/}}
{{ $replicationFactor := ternary "1" "0" (gt .Values.elasticsearch.replicas 2.0) }}
- name: CONFIG_FORCE_stackstate_kafkaGenericEventsToES_elasticsearch_index_replicas
  value: "{{ $replicationFactor  }}"
- name: CONFIG_FORCE_stackstate_kafkaMultiMetricsToES_elasticsearch_index_replicas
  value: "{{ $replicationFactor  }}"
- name: CONFIG_FORCE_stackstate_kafkaStateEventsToES_elasticsearch_index_replicas
  value: "{{ $replicationFactor  }}"
- name: CONFIG_FORCE_stackstate_kafkaStsEventsToES_elasticsearch_index_replicas
  value: "{{ $replicationFactor  }}"
- name: CONFIG_FORCE_stackstate_kafkaTraceToES_elasticsearch_index_replicas
  value: "{{ $replicationFactor  }}"
{{ $diskSpaceMB := (include "stackstate.storage.to.megabytes" .Values.elasticsearch.volumeClaimTemplate.resources.requests.storage) }}
{{ if $diskSpaceMB  }}
- name: CONFIG_FORCE_stackstate_elasticsearchDiskSpaceMB
  value: "{{ div (mul $diskSpaceMB .Values.elasticsearch.replicas) (add1 $replicationFactor) }}"
{{ end }}
- name: ELASTICSEARCH_URI
  value: "http://{{ include "stackstate.es.endpoint" . }}"
- name: KAFKA_BROKERS
  value: {{ include "stackstate.kafka.endpoint" . | quote }}
image: "{{ include "stackstate.image.registry" . }}/{{ .K2esConfig.image.repository }}{{ .Values.stackstate.components.all.image.repositorySuffix }}:{{ default .Values.stackstate.components.all.image.tag .K2esConfig.image.tag }}"
imagePullPolicy: {{ default .Values.stackstate.components.all.image.pullPolicy .K2esConfig.image.pullPolicy | quote }}
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
  initialDelaySeconds: 10
  timeoutSeconds: 5
{{- with .K2esConfig.resources }}
resources:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- if .Values.stackstate.components.all.securityContext.enabled }}
securityContext: {{- omit .Values.stackstate.components.all.securityContext "enabled" | toYaml | nindent 2 }}
{{- end }}
{{- end -}}
