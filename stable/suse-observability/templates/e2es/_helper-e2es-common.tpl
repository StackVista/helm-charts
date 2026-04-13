{{/*
Common container items for e2es (events-to-elasticsearch) deployments.
*/}}
{{- define "stackstate.e2es.deployment.common.initcontainer" -}}
name: k2es-init
command:
- sh
- -c
- |
  /entrypoint -c {{ include "stackstate.kafka.endpoint" . }},{{ include "stackstate.es.endpoint" . }} -t 300
image: "{{include "stackstate.wait.image.registry" .}}/{{ .Values.global.wait.image.repository }}:{{ .Values.global.wait.image.tag }}"
imagePullPolicy: {{ .Values.global.wait.image.pullPolicy | quote }}
{{- end -}}

{{- define "stackstate.e2es.deployment.common.container" -}}
{{- $sizingResources := include (printf "common.sizing.stackstate.%s.resources" .E2esName) . | trim -}}
{{- $sizingResourcesDict := fromYaml $sizingResources }}
{{- $evaluatedResources := merge (dict) .E2esConfig.resources $sizingResourcesDict }}
{{- $componentConfigWithResources := merge (dict "resources" $evaluatedResources) .E2esConfig -}}
env:
{{- $serviceConfig := dict "ServiceName" .E2esName "ServiceConfig" $componentConfigWithResources }}
{{/*
    Currently we use a single replicationFactor config for all indices on ES, that works fine with calculating the available disk space
    and on the STS processes assigning diskSpaceWeights to each process. But if in the future we have need to configure different
    replicationFactors per index we will need to revisit and adapt the diskSpaceWeights login on STS
*/}}
{{- $esReplicas := include "suse-observability.sizing.elasticsearch.replicas" . | trim | int -}}
{{- $replicationFactor := ternary "1" "0" (gt $esReplicas 2) -}}
{{/* Run validation of total ESDiskShare */}}
{{- include "stackstate.elastic.storage.total" . -}}
{{- $esStorage := include "suse-observability.sizing.elasticsearch.volumeClaimTemplate.resources.requests.storage" . | trim -}}
{{- $diskSpaceMB := (include "stackstate.storage.to.megabytes" $esStorage) -}}
{{- $e2esShare := .Values.stackstate.components.e2es.esDiskSpaceShare | int -}}
{{- $e2EsSharePct := (mulf (divf (.Values.stackstate.components.e2es.esDiskSpaceShare | int) $e2esShare) 100) | int -}}
{{/* Build deployment-specific env vars that can be overridden via extraEnv */}}
{{- $deploymentEnv := dict
    "CONFIG_FORCE_stackstate_kafkaTopologyEventsToES_elasticsearch_index_replicas" $replicationFactor
    "CONFIG_FORCE_stackstate_kafkaTopologyEventsToES_elasticsearch_index_diskSpaceWeight" ($e2EsSharePct | toString)
    "CONFIG_FORCE_stackstate_kafkaTopologyEventsToES_elasticsearch_index_maxIndicesRetained" (include "common.sizing.stackstate.e2es.retention" . | default (.Values.stackstate.components.e2es.retention | toString))
    "ELASTICSEARCH_URI" (printf "http://%s" (include "stackstate.es.endpoint" .))
    "KAFKA_BROKERS" (include "stackstate.kafka.endpoint" .)
    "PROMETHEUS_WRITE_ENDPOINT" (printf "http://%s:8429/api/v1/write" (include "stackstate.vmagent.endpoint" .))
}}
{{- if $diskSpaceMB }}
  {{- $_ := set $deploymentEnv "CONFIG_FORCE_stackstate_elasticsearchDiskSpaceMB" (divf (mulf (divf (mulf $diskSpaceMB $esReplicas) (add1 $replicationFactor)) .esDiskSpaceShare) 100 | int | toString) }}
{{- end }}
{{- include "stackstate.service.envvars" (merge (dict "DeploymentEnv" $deploymentEnv) $serviceConfig .) }}
image: "{{ include "stackstate.image.registry" . }}/{{ .E2esConfig.image.repository }}{{ .Values.stackstate.components.all.image.repositorySuffix }}:{{ default .Values.stackstate.components.all.image.tag .E2esConfig.image.tag }}"
imagePullPolicy: {{ default .Values.stackstate.components.all.image.pullPolicy .E2esConfig.image.pullPolicy | quote }}
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
