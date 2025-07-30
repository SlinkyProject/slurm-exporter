{{- /*
SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
SPDX-License-Identifier: Apache-2.0
*/}}

{{/*
Expand the name of the chart.
*/}}
{{- define "slurm-exporter.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "slurm-exporter.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Allow the release namespace to be overridden
*/}}
{{- define "slurm-exporter.namespace" -}}
{{- default .Release.Namespace .Values.namespaceOverride }}
{{- end }}

{{/*
Define exporter port
*/}}
{{- define "slurm-exporter.port" -}}
{{- print "8080" -}}
{{- end }}

{{/*
Define exporter labels
*/}}
{{- define "slurm-exporter.labels" -}}
helm.sh/chart: {{ include "slurm-exporter.chart" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/component: slurm-exporter
{{ include "slurm-exporter.selectorLabels" . }}
{{- end }}

{{/*
Define exporter labels
*/}}
{{- define "slurm-exporter.dashboard.labels" -}}
grafana_dashboard: "1"
{{ include "slurm-exporter.labels" . }}
{{- end }}

{{/*
Define exporter selectorLabels
*/}}
{{- define "slurm-exporter.selectorLabels" -}}
app.kubernetes.io/name: slurm-exporter
app.kubernetes.io/instance: {{ include "slurm-exporter.name" . }}
{{- end }}

{{/*
Common imagePullPolicy
*/}}
{{- define "slurm-exporter.imagePullPolicy" -}}
{{ .Values.imagePullPolicy | default "IfNotPresent" }}
{{- end }}

{{/*
Common imagePullSecrets
*/}}
{{- define "slurm-exporter.imagePullSecrets" -}}
{{- with .Values.imagePullSecrets -}}
imagePullSecrets:
  {{- . | toYaml | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Determine exporter image repository
*/}}
{{- define "image.repository" -}}
{{- print "slinky.slurm.net/slurm-exporter" -}}
{{- end }}

{{/*
Define exporter image tag
*/}}
{{- define "image.tag" -}}
{{- .Chart.Version -}}
{{- end }}

{{/*
Determine exporter image repository
*/}}
{{- define "slurm-exporter.image.repository" -}}
{{- .Values.exporter.image.repository | default (include "image.repository" .) -}}
{{- end }}

{{/*
Determine exporter image tag
*/}}
{{- define "slurm-exporter.image.tag" -}}
{{- .Values.exporter.image.tag | default (include "image.tag" .) -}}
{{- end }}

{{/*
Determine exporter image reference (repo:tag)
*/}}
{{- define "slurm-exporter.imageRef" -}}
{{- printf "%s:%s" (include "slurm-exporter.image.repository" .) (include "slurm-exporter.image.tag" .) | quote -}}
{{- end }}

{{/*
Define restapi name
*/}}
{{- define "slurm-exporter.restapi.name" -}}
{{- printf "%s-restapi" (.Release.Name) -}}
{{- end }}

{{/*
Define restapi port
*/}}
{{- define "slurm-exporter.restapi.port" -}}
{{- print "6820" -}}
{{- end }}

{{/*
Define slurm User
*/}}
{{- define "slurm-exporter.user" -}}
{{- print "slurm" -}}
{{- end }}

{{/*
Define slurm UID
*/}}
{{- define "slurm-exporter.uid" -}}
{{- print "401" -}}
{{- end }}

{{/*
Define slurm securityContext
*/}}
{{- define "slurm-exporter.securityContext" -}}
runAsNonRoot: true
runAsUser: {{ include "slurm-exporter.uid" . }}
runAsGroup: {{ include "slurm-exporter.uid" . }}
{{- end }}
