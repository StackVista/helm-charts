{{- define "stackstate.backup.envvars" -}}
- name: BACKUP_ELASTICSEARCH_BUCKET_NAME
  value: {{ .Values.backup.elasticSearch.bucketName | quote }}
- name: BACKUP_ELASTICSEARCH_RESTORE_ENABLED
  value: {{ .Values.backup.elasticSearch.restore.enabled | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_ENABLED
  value: {{ .Values.backup.elasticSearch.scheduled.enabled | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SCHEDULED
  value: {{ .Values.backup.elasticSearch.scheduled.schedule | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_INDICES
  value: {{ .Values.backup.elasticSearch.scheduled.indices | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_REPOSITORY_NAME
  value: {{ .Values.backup.elasticSearch.snapshotRepositoryName | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_POLICY_NAME
  value: {{ .Values.backup.elasticSearch.scheduled.snapshotPolicyName | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_NAME_TEMPLATE
  value: {{ .Values.backup.elasticSearch.scheduled.snapshotNameTemplate | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_EXPIRE_AFTER
  value: {{ .Values.backup.elasticSearch.scheduled.snapshotRetentionExpireAfter | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_MIN_COUNT
  value: {{ .Values.backup.elasticSearch.scheduled.snapshotRetentionMinCount | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULED_SNAPSHOT_RETENTION_MAX_COUNT
  value: {{ .Values.backup.elasticSearch.scheduled.snapshotRetentionMaxCount | quote }}
- name: BACKUP_STACKGRAPH_BUCKET_NAME
  value: {{ .Values.backup.stackGraph.bucketName | quote }}
- name: BACKUP_STACKGRAPH_RESTORE_ENABLED
  value: {{ .Values.backup.stackGraph.restore.enabled | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_ENABLED
  value: {{ .Values.backup.stackGraph.scheduled.enabled | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_TEMPLATE
  value: {{ .Values.backup.stackGraph.scheduled.backupNameTemplate | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_NAME_PARSE_REGEXP
  value: {{ .Values.backup.stackGraph.scheduled.backupNameParseRegexp | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_DATETIME_PARSE_FORMAT
  value: {{ .Values.backup.stackGraph.scheduled.backupDatetimeParseFormat | quote }}
- name: BACKUP_STACKGRAPH_SCHEDULED_BACKUP_RETENTION_TIME_DELTA
  value: {{ .Values.backup.stackGraph.scheduled.backupRetentionTimeDelta | quote }}
- name: ELASTICSEARCH_ENDPOINT
  value: {{ include "stackstate.es.endpoint" . | quote }}
- name: MINIO_ENDPOINT
  value: {{ include "stackstate.minio.endpoint" . | quote }}
{{- end -}}

{{- define "stackstate.backup.volumeMounts" -}}
- name: backup-restore-scripts
  mountPath: /backup-restore-scripts
- name: minio-keys
  mountPath: /aws-keys
{{- end -}}

{{- define "stackstate.backup.volumes" -}}
- name: backup-restore-scripts
  configMap:
    name: {{ template "common.fullname.short" . }}-backup-restore-scripts
    defaultMode: 0755
- name: minio-keys
  secret:
    secretName: {{ include "stackstate.minio.keys" . }}
{{- end -}}
