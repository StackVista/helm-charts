{{- define "stackstate.backup.hbase.configmap" -}}
{{- $escapedBucketName := .Values.backup.stackGraph.bucketName | replace "_" "___" | replace "-" "__" | replace "." "_" -}}
metadata:
  {{- /* Using a globally unique name here, because the hbase subhcart should pickup this secret */}}
  name: {{ template "common.fullname.global" (merge (dict "Base" "backup") .) }}-sts-hbase-backup
data:
  HBASE_CONF_fs_s3a_bucket_{{ $escapedBucketName }}_endpoint: {{ include "stackstate.s3proxy.endpoint" . | quote }}
  HBASE_CONF_fs_s3a_bucket_{{ $escapedBucketName }}_endpoint_region: "us-east-1"
  HBASE_CONF_fs_s3a_bucket_{{ $escapedBucketName }}_path_style_access: "true"
  HBASE_CONF_fs_s3a_bucket_{{ $escapedBucketName }}_connection_ssl_enabled: "false"
  HBASE_CONF_fs_s3a_bucket_{{ $escapedBucketName }}_aws_credentials_provider: "org.apache.hadoop.fs.s3a.SimpleAWSCredentialsProvider"
{{- end -}}

{{- $commonConfigMap := fromYaml (include "common.configmap" .) -}}
{{- $backupHBaseEnvConfigMap := fromYaml (include "stackstate.backup.hbase.configmap" .) -}}
{{ toYaml (merge $backupHBaseEnvConfigMap $commonConfigMap) }}
