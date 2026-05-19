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

Usage: 'name: "{{- template "common.fullname" . -}}"'
*/ -}}
{{- define "common.fullname"}}
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
common.fullname.unique adds a random suffix to the unique name.

This takes the same parameters as common.fullname

*/ -}}
{{- define "common.fullname.unique" -}}
  {{ template "common.fullname" . }}-{{ randAlphaNum 7 | lower }}
{{- end }}

{{- /*
common.fullname.short does not duplicate the release and chart
names if they are the same

This takes the same parameters as common.fullname

*/ -}}

{{- define "common.fullname.short"}}
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


{{- /*
Generate a shortname that can be used in any subchart and generate the same result. To this end it will not use any variables
that change between subcharts.
*/ -}}
{{- define "common.fullname.global"}}
  {{- $global := default (dict) .Values.global -}}
  {{- $base := .Chart.Name -}}
  {{- if .Base -}}
    {{- $base = .Base -}}
  {{- end -}}
  {{- if $global.fullnameOverride -}}
    {{- $base = $global.fullnameOverride -}}
  {{- else if ne $base .Release.Name -}}
    {{- $base = (printf "%s-%s" .Release.Name $base) -}}
  {{- end -}}
  {{- $gpre := default "" $global.fullnamePrefix -}}
  {{- $gsuf := default "" $global.fullnameSuffix -}}
  {{- $name := print $gpre $base $gsuf -}}
  {{- $name | lower | trunc 54 | trimSuffix "-" -}}
{{- end -}}

{{/*
'common.fullname.cluster.unique' creates a cluster-wide unique name for resources that need it,
such as ClusterRole, ClusterRoleBinding, and other non-namespaced resources.
Only if the namespace is different from the Chart name
*/}}
{{- define "common.fullname.cluster.unique" -}}
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
