{{/*
Connection configuration interface. Parents may supply configuration by overriding this
helper. ConfigMap creation, data and references must all use this same interface.
Values are evaluated with tpl by their consumers, preserving the caller's context.

In consuming templates, resolve the configuration first:
  {{- $configuration := include "kubernetes-rbac-agent.connection" . | fromYaml -}}
Then use $configuration.url.value, $configuration.url.fromConfigMap,
$configuration.clusterName.value and $configuration.clusterName.fromConfigMap.
Do not read .Values.url or .Values.clusterName directly outside this default
implementation: doing so bypasses the parent chart's configuration and can make
ConfigMap contents and workload references disagree.
TestConnectionTemplatesUseConfigurationHelper checks consuming templates for
direct .Values.url and .Values.clusterName access.
*/}}
{{- define "kubernetes-rbac-agent.connection" -}}
{{ dict "url" (default (dict) .Values.url) "clusterName" (default (dict) .Values.clusterName) | toYaml }}
{{- end -}}
