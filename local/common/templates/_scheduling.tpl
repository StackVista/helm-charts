{{/*
Resolve pod tolerations. The first layer that is a list wins, falling back to
global.suseObservability.tolerations; an explicit empty list clears the value.

Usage: {{ include "common.scheduling.tolerations" (dict "layers" (list .Values.component.tolerations .Values.all.tolerations) "context" $) }}
*/}}
{{- define "common.scheduling.tolerations" -}}
{{- $global := ((.context.Values.global | default dict).suseObservability | default dict).tolerations -}}
{{- $result := list -}}
{{- $found := false -}}
{{- range $layer := concat (.layers | default list) (list $global) -}}
  {{- if and (not $found) (eq (kindOf $layer) "slice") -}}
    {{- $result = $layer -}}
    {{- $found = true -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Resolve pod affinity per field (nodeAffinity, podAffinity, podAntiAffinity).
For each field the first layer that sets it wins, falling back to the generated
value; an explicit empty map clears the field. Lists inside a field are replaced, not merged.

Usage: {{ include "common.scheduling.affinity" (dict "layers" (list .Values.component.affinity) "generated" $generated) }}
*/}}
{{- define "common.scheduling.affinity" -}}
{{- $generated := .generated | default dict -}}
{{- $layers := .layers | default list -}}
{{- $result := dict -}}
{{- range $field := list "nodeAffinity" "podAffinity" "podAntiAffinity" -}}
  {{- $value := get $generated $field -}}
  {{- $found := false -}}
  {{- range $layer := $layers -}}
    {{- if and (not $found) (eq (kindOf $layer) "map") -}}
      {{- if eq (kindOf (get $layer $field)) "map" -}}
        {{- $value = get $layer $field -}}
        {{- $found = true -}}
      {{- end -}}
    {{- end -}}
  {{- end -}}
  {{- if and (eq (kindOf $value) "map") $value -}}
    {{- $_ := set $result $field $value -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Affinity generated from global.suseObservability.affinity. nodeAffinity applies to
every workload; podAffinity and an explicit Kubernetes podAntiAffinity only to application workloads.

Usage: {{ include "common.scheduling.globalAffinity" (dict "context" $ "application" true) | fromYaml }}
*/}}
{{- define "common.scheduling.globalAffinity" -}}
{{- $global := ((.context.Values.global | default dict).suseObservability | default dict).affinity | default dict -}}
{{- $result := dict -}}
{{- if and (eq (kindOf $global.nodeAffinity) "map") $global.nodeAffinity -}}
  {{- $_ := set $result "nodeAffinity" $global.nodeAffinity -}}
{{- end -}}
{{- if .application -}}
  {{- if and (eq (kindOf $global.podAffinity) "map") $global.podAffinity -}}
    {{- $_ := set $result "podAffinity" $global.podAffinity -}}
  {{- end -}}
  {{- $anti := $global.podAntiAffinity | default dict -}}
  {{- if eq (kindOf $anti) "map" -}}
    {{- if or (eq (kindOf (get $anti "requiredDuringSchedulingIgnoredDuringExecution")) "slice") (eq (kindOf (get $anti "preferredDuringSchedulingIgnoredDuringExecution")) "slice") -}}
      {{- $_ := set $result "podAntiAffinity" $anti -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- if $result -}}
{{- toYaml $result -}}
{{- end -}}
{{- end -}}

{{/*
Resolve an infrastructure workload's affinity: local layers over the sizing-generated
affinity, which carries the global nodeAffinity and HA podAntiAffinity.

Usage: {{ include "common.scheduling.infraAffinity" (dict "layers" (list .Values.affinity) "generated" $sizing "context" $) }}
*/}}
{{- define "common.scheduling.infraAffinity" -}}
{{- $generated := include "common.scheduling.globalAffinity" (dict "context" .context) | fromYaml -}}
{{- range $key, $value := (.generated | default dict) -}}
  {{- $_ := set $generated $key $value -}}
{{- end -}}
{{- include "common.scheduling.affinity" (dict "layers" .layers "generated" $generated) -}}
{{- end -}}
