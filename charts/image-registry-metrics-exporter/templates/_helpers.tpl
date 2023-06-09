{{/*
Expand the name of the chart.
*/}}
{{- define "image-registry-metrics-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "image-registry-metrics-exporter.fullname" -}}
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
{{- define "image-registry-metrics-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "image-registry-metrics-exporter.annotations" -}}
{{- if .Values.commonAnnotations -}}
{{ .Values.commonAnnotations | toYaml }}
{{- end }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "image-registry-metrics-exporter.labels" -}}
helm.sh/chart: {{ include "image-registry-metrics-exporter.chart" . }}
{{ include "image-registry-metrics-exporter.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- if .Values.commonLabels }}
{{ .Values.commonLabels | toYaml }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "image-registry-metrics-exporter.selectorLabels" -}}
app.kubernetes.io/name: {{ include "image-registry-metrics-exporter.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Values.podLabels }}
{{ .Values.podLabels | toYaml }}
{{- end }}
{{- end }}

{{/*
Return the proper image name
ref: https://github.com/bitnami/charts/blob/main/bitnami/common/templates/_images.tpl
*/}}
{{- define "image-registry-metrics-exporter.image" -}}
{{- $registryName := .Values.image.registry -}}
{{- $repositoryName := .Values.image.repository -}}
{{- $separator := ":" -}}
{{- $termination := .Values.image.tag | default .Chart.AppVersion | toString -}}
{{- if .Values.global }}
    {{- if .Values.global.imageRegistry }}
     {{- $registryName = .Values.global.imageRegistry -}}
    {{- end -}}
{{- end -}}
{{- if .Values.image.digest }}
    {{- $separator = "@" -}}
    {{- $termination = .Values.image.digest | toString -}}
{{- end -}}
{{- printf "%s/%s%s%s" $registryName $repositoryName $separator $termination -}}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "image-registry-metrics-exporter.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "image-registry-metrics-exporter.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

