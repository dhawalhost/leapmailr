{{/*
Expand the name of the chart.
*/}}
{{- define "generic-app.name" -}}
{{- default .Chart.Name .Values.global.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "generic-app.fullname" -}}
{{- if .Values.global.fullnameOverride -}}
{{- .Values.global.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := include "generic-app.name" . -}}
{{- printf "%s" $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "generic-app.labels" -}}
app.kubernetes.io/name: {{ include "generic-app.name" . }}
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
{{- define "generic-app.workloadFullname" -}}
{{- $root := index . 0 -}}
{{- $workloadName := index . 1 -}}
{{- printf "%s-%s" (include "generic-app.fullname" $root) $workloadName | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "generic-app.selectorLabels" -}}
{{- $root := index . 0 -}}
{{- $workloadName := index . 1 -}}
app.kubernetes.io/name: {{ include "generic-app.name" $root }}
app.kubernetes.io/instance: {{ $root.Release.Name }}
app.kubernetes.io/component: {{ $workloadName }}
{{- end -}}

{{/*
Resolve imagePullSecrets
*/}}
{{- define "generic-app.imagePullSecrets" -}}
{{- $secrets := (.Values.global.imagePullSecrets | default list) -}}
{{- if $secrets -}}
imagePullSecrets:
{{- range $secrets }}
  - name: {{ .name | default . | quote }}
{{- end }}
{{- end -}}
{{- end -}}
