{{/*
Expand the name of the chart.
*/}}
{{- define "webrtc.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "webrtc.fullname" -}}
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
Create chart label.
*/}}
{{- define "webrtc.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels.
*/}}
{{- define "webrtc.labels" -}}
helm.sh/chart: {{ include "webrtc.chart" . }}
{{ include "webrtc.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels.
*/}}
{{- define "webrtc.selectorLabels" -}}
app.kubernetes.io/name: {{ include "webrtc.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Backend full name.
*/}}
{{- define "webrtc.backend.fullname" -}}
{{- printf "%s-backend" (include "webrtc.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Frontend full name.
*/}}
{{- define "webrtc.frontend.fullname" -}}
{{- printf "%s-frontend" (include "webrtc.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}
