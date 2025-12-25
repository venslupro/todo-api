{{/*
Expand the name of the chart.
*/}}
{{- define "todo-api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "todo-api.fullname" -}}
{{- if .Values.app.fullname -}}
{{- .Values.app.fullname | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "todo-api.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "todo-api.labels" -}}
helm.sh/chart: {{ include "todo-api.chart" . }}
{{ include "todo-api.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "todo-api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "todo-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Create the name of the service account to use
*/}}
{{- define "todo-api.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
    {{ default (include "todo-api.fullname" .) .Values.serviceAccount.name }}
{{- else -}}
    {{ default "default" .Values.serviceAccount.name }}
{{- end -}}
{{- end -}}

{{/*
Create a random string if the supplied key does not exist
*/}}
{{- define "todo-api.defaultSecret" -}}
{{- if . -}}
{{- . | b64enc | quote -}}
{{- else -}}
{{- randAlphaNum 32 | b64enc | quote -}}
{{- end -}}
{{- end -}}

{{/*
Create environment variables for the application
*/}}
{{- define "todo-api.envVars" -}}
{{- range $key, $value := .Values.app.env }}
- name: {{ $key }}
  value: {{ $value | quote }}
{{- end }}
{{- end -}}

{{/*
Create secret environment variables for the application
*/}}
{{- define "todo-api.secretEnvVars" -}}
{{- $fullname := include "todo-api.fullname" . -}}
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ $fullname }}-secrets
      key: db-password
- name: JWT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $fullname }}-secrets
      key: jwt-secret
- name: STORAGE_S3_BUCKET
  valueFrom:
    secretKeyRef:
      name: {{ $fullname }}-secrets
      key: s3-bucket
- name: STORAGE_S3_REGION
  valueFrom:
    secretKeyRef:
      name: {{ $fullname }}-secrets
      key: s3-region
- name: STORAGE_S3_KEY
  valueFrom:
    secretKeyRef:
      name: {{ $fullname }}-secrets
      key: s3-access-key
- name: STORAGE_S3_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ $fullname }}-secrets
      key: s3-secret-key
{{- end -}}

{{/*
Create image reference
*/}}
{{- define "todo-api.image" -}}
{{- $registry := .Values.app.image.registry -}}
{{- $repository := .Values.app.image.repository -}}
{{- $tag := .Values.app.image.tag -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- else }}
{{- printf "%s:%s" $repository $tag -}}
{{- end }}
{{- end -}}

{{/*
Create migration image reference
*/}}
{{- define "todo-api.migrationImage" -}}
{{- $registry := .Values.migration.image.registry -}}
{{- $repository := .Values.migration.image.repository -}}
{{- $tag := .Values.migration.image.tag -}}
{{- if $registry }}
{{- printf "%s/%s:%s" $registry $repository $tag -}}
{{- else }}
{{- printf "%s:%s" $repository $tag -}}
{{- end }}
{{- end -}}