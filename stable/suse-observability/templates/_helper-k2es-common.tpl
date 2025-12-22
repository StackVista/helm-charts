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
image: "{{include "stackstate.wait.image.registry" .}}/{{ .Values.global.wait.image.repository }}:{{ .Values.global.wait.image.tag }}"
imagePullPolicy: {{ .Values.global.wait.image.pullPolicy | quote }}
{{- end -}}

{{- define "stackstate.k2es.deployment.common.container" -}}
{{- $profileResources := include (printf "common.sizing.stackstate.%s.resources" .K2esName) . | trim -}}
{{- $defaultResources := .K2esConfig.resources }}
{{- $evaluatedResources := $defaultResources }}
{{- if $profileResources }}
{{- $profileResourcesDict := fromYaml $profileResources }}
{{- $evaluatedResources = merge $profileResourcesDict $defaultResources }}
{{- end }}
{{- $componentConfigWithResources := merge (dict "resources" $evaluatedResources) .K2esConfig -}}
env:
{{- $serviceConfig := dict "ServiceName" .K2esName "ServiceConfig" $componentConfigWithResources }}
{{- include "stackstate.service.envvars" (merge $serviceConfig .) }}
{{/*
    Currently we use a single replicationFactor config for all indices on ES, that works fine with calculating the available disk space
    and on the STS processes assigning diskSpaceWeights to each process. But if in the future we have need to configure different
    replicationFactors per index we will need to revisit and adapt the diskSpaceWeights login on STS
*/}}
{{- $esReplicas := include "suse-observability.sizing.elasticsearch.replicas" . | trim | int -}}
{{ $replicationFactor := ternary "1" "0" (gt $esReplicas 2) }}
- name: CONFIG_FORCE_stackstate_kafkaTopologyEventsToES_elasticsearch_index_replicas
  value: "{{ $replicationFactor  }}"
{{/*
Run validation of total ESDiskShare
*/}}
{{ include "stackstate.elastic.storage.total" . }}
{{- $esStorage := include "suse-observability.sizing.elasticsearch.volumeClaimTemplate.resources.requests.storage" . | trim -}}
{{ $diskSpaceMB := (include "stackstate.storage.to.megabytes" $esStorage) }}
{{ if $diskSpaceMB  }}
- name: CONFIG_FORCE_stackstate_elasticsearchDiskSpaceMB
  value: "{{ divf (mulf (divf (mulf $diskSpaceMB $esReplicas) (add1 $replicationFactor)) .esDiskSpaceShare) 100 | int }}"
{{ end }}
{{- $k2esShare := .Values.stackstate.components.e2es.esDiskSpaceShare | int -}}
{{- $e2EsShare := (mulf (divf (.Values.stackstate.components.e2es.esDiskSpaceShare | int) $k2esShare) 100) | int  -}}
- name: CONFIG_FORCE_stackstate_kafkaTopologyEventsToES_elasticsearch_index_diskSpaceWeight
  value: "{{ $e2EsShare }}"
- name: CONFIG_FORCE_stackstate_kafkaTopologyEventsToES_elasticsearch_index_maxIndicesRetained
  value: "{{ include "common.sizing.stackstate.e2es.retention" . | default .Values.stackstate.components.e2es.retention }}"
- name: ELASTICSEARCH_URI
  value: "http://{{ include "stackstate.es.endpoint" . }}"
- name: KAFKA_BROKERS
  value: {{ include "stackstate.kafka.endpoint" . | quote }}
- name: PROMETHEUS_WRITE_ENDPOINT
  value: http://{{ include "stackstate.vmagent.endpoint" . }}:8429/api/v1/write
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
{{- with $evaluatedResources }}
resources:
  {{- toYaml . | nindent 2 }}
{{- end }}
{{- end -}}
