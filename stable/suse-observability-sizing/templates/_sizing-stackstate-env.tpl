{{/*
=============================================================================
STACKSTATE ENVIRONMENT VARIABLES TEMPLATES
These templates provide CONFIG_FORCE environment variables for
performance tuning based on sizing profiles
=============================================================================
*/}}

{{/*
Get stackstate all components extraEnv.open CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.all.extraEnv.open" . }}
*/}}
{{- define "common.sizing.stackstate.all.extraEnv.open" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}
CONFIG_FORCE_stackstate_topologyQueryService_maxStackElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_topologyQueryService_maxLoadedElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_agents_agentLimit: "10"
CONFIG_FORCE_stackgraph_retentionWindowMs: "259200000"
CONFIG_FORCE_stackstate_traces_retentionDays: "3"
{{- else if eq $profile "150-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "150"
{{- else if eq $profile "250-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "250"
{{- else if eq $profile "500-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "500"
{{- else if eq $profile "4000-ha" }}
CONFIG_FORCE_stackstate_agents_agentLimit: "4000"
CONFIG_FORCE_stackstate_eventPersistor_kafkaProducerConfig_request.timeout.ms: "25000"
CONFIG_FORCE_stackstate_eventPersistor_kafkaProducerConfig_delivery.timeout.ms: "30000"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate server CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.server.env" . }}
*/}}
{{- define "common.sizing.stackstate.server.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "trial" }}
CONFIG_FORCE_stackstate_sync_initializationBatchParallelism: "1"
CONFIG_FORCE_stackstate_healthSync_initialLoadParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initializationParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initialLoadTransactionSize: "2500"
{{- else if eq $profile "10-nonha" }}
CONFIG_FORCE_stackstate_topologyQueryService_maxStackElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_topologyQueryService_maxLoadedElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_agents_agentLimit: "10"
CONFIG_FORCE_stackstate_sync_initializationBatchParallelism: "1"
CONFIG_FORCE_stackstate_healthSync_initialLoadParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initializationParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initialLoadTransactionSize: "2500"
{{- else if eq $profile "20-nonha" }}
CONFIG_FORCE_stackstate_topologyQueryService_maxStackElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_topologyQueryService_maxLoadedElementsPerQuery: "1000"
CONFIG_FORCE_stackstate_agents_agentLimit: "20"
CONFIG_FORCE_stackstate_sync_initializationBatchParallelism: "1"
CONFIG_FORCE_stackstate_healthSync_initialLoadParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initializationParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initialLoadTransactionSize: "2500"
{{- else if eq $profile "50-nonha" }}
CONFIG_FORCE_stackstate_topologyQueryService_maxStackElementsPerQuery: "5000"
CONFIG_FORCE_stackstate_topologyQueryService_maxLoadedElementsPerQuery: "5000"
CONFIG_FORCE_stackstate_agents_agentLimit: "50"
CONFIG_FORCE_stackstate_sync_initializationBatchParallelism: "1"
CONFIG_FORCE_stackstate_healthSync_initialLoadParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initializationParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initialLoadTransactionSize: "2500"
{{- else if eq $profile "100-nonha" }}
CONFIG_FORCE_stackstate_topologyQueryService_maxStackElementsPerQuery: "7500"
CONFIG_FORCE_stackstate_topologyQueryService_maxLoadedElementsPerQuery: "7500"
CONFIG_FORCE_stackstate_agents_agentLimit: "100"
CONFIG_FORCE_stackstate_sync_initializationBatchParallelism: "1"
CONFIG_FORCE_stackstate_healthSync_initialLoadParallelism: "1"
CONFIG_FORCE_stackstate_stateService_initializationParallelism: "1"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate api CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.api.env" . }}
*/}}
{{- define "common.sizing.stackstate.api.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- /* No profile-specific API env vars in legacy profiles */ -}}
{{- end }}
{{- end }}

{{/*
Get stackstate state CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.state.env" . }}
*/}}
{{- define "common.sizing.stackstate.state.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- /* No profile-specific state env vars for 500-ha */ -}}
{{- end }}
{{- end }}

{{/*
Get stackstate sync CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.sync.env" . }}
*/}}
{{- define "common.sizing.stackstate.sync.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if hasSuffix "-ha" $profile }}
CONFIG_FORCE_stackgraph_featureswitches_readCache: "false"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate healthSync CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.healthSync.env" . }}
*/}}
{{- define "common.sizing.stackstate.healthSync.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "4000-ha" }}
CONFIG_FORCE_stackstate_healthSync_initializationTimeout: "30 minutes"
CONFIG_FORCE_stackstate_healthSync_maxIdentifiersProcessingDelay: "30 minutes"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate correlate CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.correlate.env" . }}
*/}}
{{- define "common.sizing.stackstate.correlate.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if eq $profile "250-ha" }}
CONFIG_FORCE_stackstate_correlate_correlateConnections_workers: "2"
CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers: "2"
CONFIG_FORCE_stackstate_correlate_aggregation_workers: "2"
{{- else if or (eq $profile "500-ha") (eq $profile "4000-ha") }}
CONFIG_FORCE_stackstate_correlate_correlateConnections_workers: "4"
CONFIG_FORCE_stackstate_correlate_correlateHTTPTraces_workers: "4"
CONFIG_FORCE_stackstate_correlate_aggregation_workers: "4"
{{- if eq $profile "4000-ha" }}
CONFIG_FORCE_stackstate_correlate_correlateConnections_kafka_producer_request.timeout.ms: "25000"
CONFIG_FORCE_stackstate_correlate_correlateConnections_kafka_producer_delivery.timeout.ms: "30000"
{{- end }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate receiver CONFIG_FORCE environment variables
Usage: {{ include "common.sizing.stackstate.receiver.env" . }}
*/}}
{{- define "common.sizing.stackstate.receiver.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "trial") (eq $profile "10-nonha") (eq $profile "20-nonha") (eq $profile "50-nonha") (eq $profile "100-nonha") }}
CONFIG_FORCE_akka_http_host__connection__pool_max__open__requests: "256"
{{- else if or (eq $profile "150-ha") (eq $profile "250-ha") (eq $profile "500-ha") }}
CONFIG_FORCE_akka_http_host__connection__pool_max__open__requests: "256"
{{- else if eq $profile "4000-ha" }}
CONFIG_FORCE_akka_http_host__connection__pool_max__open__requests: "384"
CONFIG_FORCE_stackstate_receiver_countBufferSize: "3000000"
CONFIG_FORCE_zstd__decompress__dispatcher_fork__join__executor_parallelism__factor: "4.0"
CONFIG_FORCE_zstd__decompress__dispatcher_fork__join__executor_parallelism__max: "64"
CONFIG_FORCE_akka_actor_default__dispatcher_fork__join__executor_parallelism__min: "16"
CONFIG_FORCE_akka_actor_default__dispatcher_fork__join__executor_parallelism__factor: "4.0"
CONFIG_FORCE_akka_actor_default__dispatcher_fork__join__executor_parallelism__max: "64"
CONFIG_FORCE_akka_actor_default__blocking__io__dispatcher_thread__pool__executor_fixed__pool__size: "64"
CONFIG_FORCE_stackstate_receiver_kafkaProducerConfig_max_request_size: "4194304"
{{- end }}
{{- end }}
{{- end }}

{{/*
Get stackstate kafkaTopicCreate environment variables
Usage: {{ include "common.sizing.stackstate.kafkaTopicCreate.env" . }}
*/}}
{{- define "common.sizing.stackstate.kafkaTopicCreate.env" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
{{- $profile := .Values.global.suseObservability.sizing.profile -}}
{{- if or (eq $profile "500-ha") (eq $profile "4000-ha") }}
KAFKA_PARTITIONS_sts_correlate_endpoints: "40"
KAFKA_PARTITIONS_sts_correlate_http_trace_observations: "40"
KAFKA_PARTITIONS_sts_correlated_connections: "40"
{{- end }}
{{- end }}
{{- end }}
