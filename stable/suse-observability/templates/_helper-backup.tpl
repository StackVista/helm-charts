{{- /*
  Adding a trailing slash to a value if it is not empty.
*/ -}}
{{- define "ensureTrailingSlashIfNotEmpty" -}}
  {{- if . -}}
    {{- printf "%s/" (. | trimSuffix "/") -}}
  {{- else -}}
    {{- "" -}}
  {{- end -}}
{{- end -}}

{{- /*
  Removing trailing slashes if any from the string.
*/ -}}
{{- define "trimTrailingSlashes" -}}
{{- . | trimSuffix "/" -}}
{{- end -}}

{{- define "stackstate.backup.envvars" -}}
{{/*
Check if the backup.stackGraph.splitArchiveSize has a valid value.
*/}}
{{- if not (regexMatch "^[0-9]+([KMG]?|[KMG]?[B]?)$" (toString .Values.backup.stackGraph.splitArchiveSize)) }}
  {{- fail (printf ".Values.backup.stackGraph.splitArchiveSize (%v) has to be an integer greater or equal to 0 with an optional suffix K,M,G,KB,MB,GB" .Values.backup.stackGraph.splitArchiveSize)}}
{{- end }}
- name: BACKUP_ELASTICSEARCH_BUCKET_NAME
  value: {{ .Values.backup.elasticsearch.bucketName | quote }}
- name: BACKUP_ELASTICSEARCH_S3_PREFIX
  value: {{ include "trimTrailingSlashes" .Values.backup.elasticsearch.s3Prefix | quote }}
- name: BACKUP_ELASTICSEARCH_RESTORE_ENABLED
  value: {{ .Values.backup.elasticsearch.restore.enabled | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_ENABLED
  value: {{ .Values.backup.elasticsearch.scheduled.enabled | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SCHEDULED
  value: {{ .Values.backup.elasticsearch.scheduled.schedule | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_INDICES
  value: {{ .Values.backup.elasticsearch.scheduled.indices | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME
  value: {{ .Values.backup.elasticsearch.snapshotRepositoryName | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_POLICY_NAME
  value: {{ .Values.backup.elasticsearch.scheduled.snapshotPolicyName | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_NAME_TEMPLATE
  value: {{ .Values.backup.elasticsearch.scheduled.snapshotNameTemplate | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_EXPIRE_AFTER
  value: {{ .Values.backup.elasticsearch.scheduled.snapshotRetentionExpireAfter | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_MIN_COUNT
  value: {{ .Values.backup.elasticsearch.scheduled.snapshotRetentionMinCount | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_MAX_COUNT
  value: {{ .Values.backup.elasticsearch.scheduled.snapshotRetentionMaxCount | quote }}
- name: BACKUP_STACKGRAPH_BUCKET_NAME
  value: {{ .Values.backup.stackGraph.bucketName | quote }}
- name: BACKUP_STACKGRAPH_S3_PREFIX
  value: {{ include "ensureTrailingSlashIfNotEmpty" .Values.backup.stackGraph.s3Prefix }}
- name: BACKUP_STACKGRAPH_RESTORE_ENABLED
  value: {{ .Values.backup.stackGraph.restore.enabled | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_ENABLED
  value: {{ .Values.backup.stackGraph.scheduled.enabled | quote }}
- name: BACKUP_STACKGRAPH_ARCHIVE_SPLIT_SIZE
  value: {{ .Values.backup.stackGraph.splitArchiveSize | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_TEMPLATE
  value: {{ .Values.backup.stackGraph.scheduled.backupNameTemplate | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_PARSE_REGEXP
  value: {{ .Values.backup.stackGraph.scheduled.backupNameParseRegexp | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT
  value: {{ .Values.backup.stackGraph.scheduled.backupDatetimeParseFormat | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_RETENTION_TIME_DELTA
  value: {{ .Values.backup.stackGraph.scheduled.backupRetentionTimeDelta | quote }}
- name: BACKUP_CONFIGURATION_UPLOAD_REMOTE
  value: {{ .Values.backup.enabled | toString | lower | quote }}
- name: BACKUP_CONFIGURATION_BUCKET_NAME
  value: {{ .Values.backup.configuration.bucketName | quote }}
- name: BACKUP_CONFIGURATION_S3_PREFIX
  value: {{ include "ensureTrailingSlashIfNotEmpty" .Values.backup.configuration.s3Prefix }}
- name: BACKUP_CONFIGURATION_RESTORE_ENABLED
  value: {{ .Values.backup.configuration.restore.enabled | quote }}
- name: BACKUP_CONFIGURATION_SCHEDULED_ENABLED
  value: {{ .Values.backup.configuration.scheduled.enabled | quote }}
- name: BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_TEMPLATE
  value: {{ .Values.backup.configuration.scheduled.backupNameTemplate | quote }}
- name: BACKUP_CONFIGURATION_SCHEDULED_BACKUP_NAME_PARSE_REGEXP
  value: {{ .Values.backup.configuration.scheduled.backupNameParseRegexp | quote }}
- name: BACKUP_CONFIGURATION_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT
  value: {{ .Values.backup.configuration.scheduled.backupDatetimeParseFormat | quote }}
- name: BACKUP_CONFIGURATION_SCHEDULED_BACKUP_RETENTION_TIME_DELTA
  value: {{ .Values.backup.configuration.scheduled.backupRetentionTimeDelta | quote }}
- name: BACKUP_CONFIGURATION_MAX_LOCAL_FILES
  value: {{ .Values.backup.configuration.maxLocalFiles | quote }}
- name: STACKSTATE_ROUTER_ENDPOINT
  value: {{ include "stackstate.router.endpoint" . | quote }}
- name: ELASTICSEARCH_ENDPOINT
  value: {{ include "stackstate.es.endpoint" . | quote }}
- name: BACKUP_VICTORIA_METRICS_0_ENABLED
  value: {{ index .Values "victoria-metrics-0" "backup" "enabled" | quote }}
- name: BACKUP_VICTORIA_METRICS_0_BUCKET_NAME
  value: {{ index .Values "victoria-metrics-0" "backup" "bucketName" | quote }}
- name: BACKUP_VICTORIA_METRICS_0_S3_PREFIX
  value: {{ include "trimTrailingSlashes" (index .Values "victoria-metrics-0" "backup" "s3Prefix") }}
- name: BACKUP_VICTORIA_METRICS_0_RESTORE_ENABLED
  value: {{ index .Values "victoria-metrics-0" "restore" "enabled" | quote }}
- name: BACKUP_VICTORIA_METRICS_1_ENABLED
  value: {{ index .Values "victoria-metrics-1" "backup" "enabled" | quote }}
- name: BACKUP_VICTORIA_METRICS_1_BUCKET_NAME
  value: {{ index .Values "victoria-metrics-1" "backup" "bucketName" | quote }}
- name: BACKUP_VICTORIA_METRICS_1_S3_PREFIX
  value: {{ include "trimTrailingSlashes" (index .Values  "victoria-metrics-1" "backup" "s3Prefix") }}
- name: BACKUP_VICTORIA_METRICS_1_RESTORE_ENABLED
  value: {{ index .Values "victoria-metrics-1" "restore" "enabled" | quote }}
- name: BACKUP_CLICKHOUSE_BUCKET_NAME
  value: {{ .Values.clickhouse.backup.bucketName | quote }}
- name: BACKUP_CLICKHOUSE_S3_PREFIX
  value: {{ include "ensureTrailingSlashIfNotEmpty" .Values.clickhouse.backup.s3Prefix }}
- name: BACKUP_CLICKHOUSE_RESTORE_ENABLED
  value: {{ .Values.clickhouse.restore.enabled | quote }}
- name: BACKUP_CLICKHOUSE_SCHEDULED_ENABLED
  value: {{ .Values.clickhouse.backup.enabled | quote }}
- name: MINIO_ENDPOINT
  value: {{ include "stackstate.minio.endpoint" . | quote }}
{{- include "stackstate.env.platform_version" . }}
{{- end -}}

{{- define "stackstate.backup.volumeMounts" -}}
- name: backup-log
  mountPath: /opt/docker/etc_log
- name: backup-restore-scripts
  mountPath: /backup-restore-scripts
{{- if .Values.backup.enabled }}
- name: minio-keys
  mountPath: /aws-keys
{{- end -}}
{{- end -}}

{{- define "stackstate.backup.volumes" -}}
- name: backup-log
  configMap:
    name: {{ template "common.fullname.short" . }}-backup-log
- name: backup-restore-scripts
  configMap:
    name: {{ template "common.fullname.short" . }}-backup-restore-scripts
    defaultMode: 0755
{{- if .Values.backup.enabled }}
- name: minio-keys
  secret:
    secretName: {{ include "stackstate.minio.keys" . }}
{{- end -}}
{{- end -}}
