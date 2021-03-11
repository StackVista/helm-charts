{{- define "stackstate.backup.envvars" -}}
- name: BACKUP_ELASTICSEARCH_BUCKET_NAME
  value: {{ .Values.backup.elasticsearch.bucketName | quote }}
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
- name: backup-log
  mountPath: /opt/docker/etc_log
- name: backup-restore-scripts
  mountPath: /backup-restore-scripts
- name: minio-keys
  mountPath: /aws-keys
{{- end -}}

{{- define "stackstate.backup.volumes" -}}
- name: backup-log
  configMap:
    name: {{ template "common.fullname.short" . }}-backup-log
- name: backup-restore-scripts
  configMap:
    name: {{ template "common.fullname.short" . }}-backup-restore-scripts
    defaultMode: 0755
- name: minio-keys
  secret:
    secretName: {{ include "stackstate.minio.keys" . }}
{{- end -}}
