{{/*
Resource names retain the existing expressions, including fixed RBAC/PVC names.
Keep manager.name and worker.name separate for labels and selectors.
*/}}
{{- define "anomaly-detection.manager.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-spotlight-manager
{{- end -}}

{{- define "anomaly-detection.worker.deployment.fullname" -}}
{{ template "common.fullname.short" . }}-spotlight-worker
{{- end -}}

{{- define "anomaly-detection.manager.service.fullname" -}}
{{ template "common.fullname.short" . }}-spotlight-manager
{{- end -}}

{{- define "anomaly-detection.manager.servicemonitor.fullname" -}}
{{ template "common.fullname.short" . }}-spotlight-manager
{{- end -}}

{{- define "anomaly-detection.serviceaccount.fullname" -}}
{{ template "common.fullname.short" . }}-sa
{{- end -}}

{{- define "anomaly-detection.authentication.role.fullname" -}}
stackstate-aad
{{- end -}}

{{- define "anomaly-detection.authentication.rolebinding.fullname" -}}
stackstate-aad
{{- end -}}

{{- define "anomaly-detection.authentication.clusterrolebinding.fullname" -}}
{{ template "common.fullname.cluster.unique" . }}-aad-authentication
{{- end -}}

{{- define "anomaly-detection.authentication.secret.fullname" -}}
{{ template "common.fullname.short" . }}-stackstate-auth-secret
{{- end -}}

{{- define "anomaly-detection.pull.secret.fullname" -}}
{{ template "common.fullname.short" . }}-pull-secret
{{- end -}}

{{- define "anomaly-detection.config.configmap.fullname" -}}
{{ template "common.fullname.short" . }}-spotlight-config-base
{{- end -}}

{{- define "anomaly-detection.manager.artifacts.persistentvolumeclaim.fullname" -}}
spotlight-artifacts-volume-claim
{{- end -}}

{{- define "anomaly-detection.ingress.fullname" -}}
{{ include "common.fullname.short" . }}
{{- end -}}

{{- define "anomaly-detection.pdb.fullname" -}}
{{ template "common.fullname.short" . }}-anomaly-detection
{{- end -}}
