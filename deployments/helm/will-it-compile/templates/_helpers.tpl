{{/*
Expand the name of the chart.
*/}}
{{- define "will-it-compile.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "will-it-compile.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "will-it-compile.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "will-it-compile.labels" -}}
helm.sh/chart: {{ include "will-it-compile.chart" . }}
{{ include "will-it-compile.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "will-it-compile.selectorLabels" -}}
app.kubernetes.io/name: {{ include "will-it-compile.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app: will-it-compile
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "will-it-compile.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "will-it-compile.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}
