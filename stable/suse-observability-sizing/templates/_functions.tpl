{{/*
Validate a URL. This helper uses the built-in 'required' function
to ensure the value exists and then checks if the URL has a scheme (e.g., "http://").

Usage:
{{ include "mychart.validate.url" (dict "value" .Values.some.url "errorMessage" "A custom error message for when the value is missing.") }}

Parameters:
  . (dict): A dictionary containing:
    - "value": The URL string to validate.
    - "errorMessage": The error message to display if "value" is empty (used by 'required').
*/}}
{{- define "sts.values.validateUrl" -}}
  {{- $url := required .errorMessage .value -}}
  {{- $parsedUrl := urlParse $url -}}
  {{- if not $parsedUrl.scheme -}}
    {{- fail (printf "Invalid URL format for '%s'. The URL must include a scheme (e.g., 'http://' or 'https://')." $url) -}}
  {{- end -}}
{{- end -}}

{{- define "sts.values.receiverKey" -}}
{{- if .Values.receiverApiKey -}}
{{ .Values.receiverApiKey }}
{{- else -}}
{{ randAlphaNum 32 | quote }}
{{- end -}}
{{- end -}}

{{/* Checks whether an `adminPassword` is set, if not it will generate
a new password and set this for printing. If the password is set,
this function will validate it is correctly `bcrypt` hashed, if not, it will
hash it for the outputted values.
*/}}
{{- define "sts.values.getOrGenerateAdminPassword" -}}
{{- $pwd := .Values.adminPassword -}}
{{- if $pwd -}}
{{- if regexMatch "^\\$2[abxy]{0,1}\\$(0[4-9]|[12][0-9]|3[01])\\$[./0-9a-zA-Z]{53}$" $pwd -}}
{{ $pwd }}
{{- else -}}
{{ $pwd | bcrypt | quote }}
{{- end -}}
{{- else -}}
{{- $pwd = randAlphaNum 16 -}}
{{- $ignored := set .Values "generatedAdminPassword" $pwd -}}
{{ $pwd | bcrypt | quote }}
{{- end -}}
{{- end -}}

{{/* Generate podAntiAffinity configuration */}}
{{- define "sts.values.podAntiAffinity" -}}
{{- $labels := index . 0 -}}
{{- $required := index . 1 -}}
{{- $topologyKey := index . 2 -}}
podAntiAffinity:
{{- if $required }}
  requiredDuringSchedulingIgnoredDuringExecution:
  - labelSelector:
      matchLabels:
{{- range $key, $value := $labels }}
        {{ $key }}: {{ $value }}
{{- end }}
    topologyKey: {{ $topologyKey }}
{{- else }}
  preferredDuringSchedulingIgnoredDuringExecution:
  - weight: 100
    podAffinityTerm:
      labelSelector:
        matchLabels:
{{- range $key, $value := $labels }}
          {{ $key }}: {{ $value }}
{{- end }}
      topologyKey: {{ $topologyKey }}
{{- end }}
{{- end -}}

{{/*
Get list of valid sizing profiles
Usage: {{ include "common.sizing.profiles.list" . }}
Returns: List of valid profile names
*/}}
{{- define "common.sizing.profiles.list" -}}
{{- list "trial" "10-nonha" "20-nonha" "50-nonha" "100-nonha" "150-ha" "250-ha" "500-ha" "4000-ha" -}}
{{- end -}}

{{/*
Validate sizing profile
Usage: {{ include "common.sizing.profiles.validate" . }}
Fails if global.suseObservability.sizing.profile is set but invalid
*/}}
{{- define "common.sizing.profiles.validate" -}}
{{- if and .Values.global .Values.global.suseObservability .Values.global.suseObservability.sizing .Values.global.suseObservability.sizing.profile -}}
  {{- $profile := .Values.global.suseObservability.sizing.profile -}}
  {{- $validProfiles := list "trial" "10-nonha" "20-nonha" "50-nonha" "100-nonha" "150-ha" "250-ha" "500-ha" "4000-ha" -}}
  {{- if not (has $profile $validProfiles) -}}
    {{- fail (printf "Invalid sizing profile '%s'. Valid profiles are: %s" $profile (join ", " (sortAlpha $validProfiles))) -}}
  {{- end -}}
{{- end -}}
{{- end -}}

{{/*
Profile lookup helper - returns value from a dict based on current sizing profile
Usage: {{ include "common.sizing.profileLookup" (dict "profileMap" $profileMap "context" .) }}
Parameters:
  - profileMap: A dict mapping profile names to values
  - context: The template context (.)
Returns: The value for the current profile, or empty string if no profile is set
*/}}
{{- define "common.sizing.profileLookup" -}}
{{- $profileMap := .profileMap -}}
{{- $context := .context -}}
{{- if and $context.Values.global $context.Values.global.suseObservability $context.Values.global.suseObservability.sizing $context.Values.global.suseObservability.sizing.profile -}}
  {{- $profile := $context.Values.global.suseObservability.sizing.profile -}}
  {{- if hasKey $profileMap $profile -}}
    {{- index $profileMap $profile -}}
  {{- end -}}
{{- end -}}
{{- end -}}
