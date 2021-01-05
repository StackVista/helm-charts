{{- define "stackstate.backup.envvars" -}}
- name: BACKUP_ELASTICSEARCH_ENABLED
  value: {{ .Values.backup.elasticSearch.enabled | quote }}
- name: BACKUP_ELASTICSEARCH_SCHEDULE
  value: {{ .Values.backup.elasticSearch.schedule | quote }}
- name: BACKUP_ELASTICSEARCH_BUCKET_NAME
  value: {{ .Values.backup.elasticSearch.bucketName | quote }}
- name: BACKUP_ELASTICSEARCH_INDICES
  value: {{ .Values.backup.elasticSearch.indices | quote }}
- name: BACKUP_ELASTICSEARCH_SNAPSHOT_REPOSITORY_NAME
  value: {{ .Values.backup.elasticSearch.snapshotRepositoryName | quote }}
- name: BACKUP_ELASTICSEARCH_SNAPSHOT_POLICY_NAME
  value: {{ .Values.backup.elasticSearch.snapshotPolicyName | quote }}
- name: BACKUP_ELASTICSEARCH_SNAPSHOT_NAME_TEMPLATE
  value: {{ .Values.backup.elasticSearch.snapshotNameTemplate | quote }}
- name: BACKUP_ELASTICSEARCH_SNAPSHOT_RETENTION_EXPIRE_AFTER
  value: {{ .Values.backup.elasticSearch.snapshotRetentionExpireAfter | quote }}
- name: BACKUP_ELASTICSEARCH_SNAPSHOT_RETENTION_MIN_COUNT
  value: {{ .Values.backup.elasticSearch.snapshotRetentionMinCount | quote }}
- name: BACKUP_ELASTICSEARCH_SNAPSHOT_RETENTION_MAX_COUNT
  value: {{ .Values.backup.elasticSearch.snapshotRetentionMaxCount | quote }}
- name: BACKUP_STACKGRAPH_ENABLED
  value: {{ .Values.backup.stackGraph.enabled | quote }}
- name: BACKUP_STACKGRAPH_BACKUP_NAME_TEMPLATE
  value: {{ .Values.backup.stackGraph.backupNameTemplate | quote }}
- name: BACKUP_STACKGRAPH_BACKUP_NAME_PARSE_REGEXP
  value: {{ .Values.backup.stackGraph.backupNameParseRegexp | quote }}
- name: BACKUP_STACKGRAPH_BACKUP_DATETIME_PARSE_FORMAT
  value: {{ .Values.backup.stackGraph.backupDatetimeParseFormat | quote }}
- name: BACKUP_STACKGRAPH_BACKUP_RETENTION_TIME_DELTA
  value: {{ .Values.backup.stackGraph.backupRetentionTimeDelta | quote }}
- name: BACKUP_STACKGRAPH_BUCKET_NAME
  value: {{ .Values.backup.stackGraph.bucketName | quote }}
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
