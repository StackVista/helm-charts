{{- /*
fullname defines a suitably unique name for a resource by combining
the release name and the chart name.

The prevailing wisdom is that names should only contain a-z, 0-9 plus dot (.) and dash (-), and should
not exceed 63 characters.

Parameters:

- .Values.fullnameOverride: Replaces the computed name with this given name
- .Values.fullnamePrefix: Prefix
- .Values.global.fullnamePrefix: Global prefix
- .Values.fullnameSuffix: Suffix
- .Values.global.fullnameSuffix: Global suffix

The applied order is: "global prefix + prefix + name + suffix + global suffix"

Usage: 'name: "{{- template "fullname" . -}}"'
*/ -}}
{{- define "fullname"}}
  {{- $global := default (dict) .Values.global -}}
  {{- $base := default (printf "%s-%s" .Release.Name .Chart.Name) .Values.fullnameOverride -}}
  {{- $gpre := default "" $global.fullnamePrefix -}}
  {{- $pre := default "" .Values.fullnamePrefix -}}
  {{- $suf := default "" .Values.fullnameSuffix -}}
  {{- $gsuf := default "" $global.fullnameSuffix -}}
  {{- $name := print $gpre $pre $base $suf $gsuf -}}
  {{- $name | lower | trunc 54 | trimSuffix "-" -}}
{{- end -}}

{{- /*
fullname.unique adds a random suffix to the unique name.

This takes the same parameters as fullname

*/ -}}
{{- define "fullname.unique" -}}
  {{ template "fullname" . }}-{{ randAlphaNum 7 | lower }}
{{- end }}

{{- /*
fullname.short does not duplicate the release and chart
names if they are the same

This takes the same parameters as fullname

*/ -}}

{{- define "fullname.short"}}
  {{- $global := default (dict) .Values.global -}}
  {{- $base := .Chart.Name -}}
  {{- if .Values.fullnameOverride -}}
    {{- $base = .Values.fullnameOverride -}}
  {{- else if ne $base .Release.Name -}}
    {{- $base = (printf "%s-%s" .Release.Name .Chart.Name) -}}
  {{- end -}}
  {{- $gpre := default "" $global.fullnamePrefix -}}
  {{- $pre := default "" .Values.fullnamePrefix -}}
  {{- $suf := default "" .Values.fullnameSuffix -}}
  {{- $gsuf := default "" $global.fullnameSuffix -}}
  {{- $name := print $gpre $pre $base $suf $gsuf -}}
  {{- $name | lower | trunc 54 | trimSuffix "-" -}}
{{- end -}}

{{/*
'fullname.cluster.unique' creates a cluster-wide unique name for resources that need it,
such as ClusterRole, ClusterRoleBinding, and other non-namespaced resources.
Only if the namespace is different from the Chart name
*/}}
{{- define "fullname.cluster.unique" -}}
  {{- $global := default (dict) .Values.global -}}
  {{- $base := .Chart.Name -}}
  {{- if .Values.fullnameOverride -}}
    {{- $base = .Values.fullnameOverride -}}
  {{- else -}}
    {{- if ne $base .Release.Name -}}
      {{- $base = (printf "%s-%s" .Release.Name .Chart.Name) -}}
    {{- end -}}
    {{- if ne $base .Release.Namespace -}}
      {{- $base = (printf "%s-%s" .Release.Namespace $base) -}}
    {{- end -}}
  {{- end -}}
  {{- $gpre := default "" $global.fullnamePrefix -}}
  {{- $pre := default "" .Values.fullnamePrefix -}}
  {{- $suf := default "" .Values.fullnameSuffix -}}
  {{- $gsuf := default "" $global.fullnameSuffix -}}
  {{- $name := print $gpre $pre $base $suf $gsuf -}}
  {{- $name | lower | trunc 54 | trimSuffix "-" -}}

{{- end -}}
