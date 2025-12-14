{{/*
Expand the name of the chart.
*/}}
{{- define "leapmailr.name" -}}
{{- default .Chart.Name .Values.global.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "leapmailr.fullname" -}}
{{- if .Values.global.fullnameOverride -}}
{{- .Values.global.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "leapmailr.name" . -}}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "leapmailr.labels" -}}
app.kubernetes.io/name: {{ include "leapmailr.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" }}
{{- range $k, $v := (.Values.global.labels | default dict) }}
{{ $k }}: {{ $v | quote }}
{{- end }}
{{- end -}}

{{/*
Workload fullname
*/}}
{{- define "leapmailr.workloadFullname" -}}
{{- $root := index . 0 -}}
{{- $workloadName := index . 1 -}}
{{- printf "%s-%s" (include "leapmailr.fullname" $root) $workloadName | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "leapmailr.selectorLabels" -}}
{{- $root := index . 0 -}}
{{- $workloadName := index . 1 -}}
app.kubernetes.io/name: {{ include "leapmailr.name" $root }}
app.kubernetes.io/instance: {{ $root.Release.Name }}
app.kubernetes.io/component: {{ $workloadName }}
{{- end -}}

{{/*
Resolve imagePullSecrets
*/}}
{{- define "leapmailr.imagePullSecrets" -}}
{{- $secrets := (.Values.global.imagePullSecrets | default list) -}}
{{- if $secrets -}}
imagePullSecrets:
{{- range $secrets }}
  - name: {{ .name | default . | quote }}
{{- end }}
{{- end -}}
{{- end -}}
