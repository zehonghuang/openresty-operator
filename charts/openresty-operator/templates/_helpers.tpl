{{- define "openresty-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{ .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- printf " %s" $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "openresty-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.name }}
{{ .Values.serviceAccount.name }}
{{- else }}
{{- include "openresty-operator.fullname" . }}
{{- end }}
{{- end }}

{{/* Fail if a required value is missing */}}
{{- define "crds.required" -}}
{{- if not (index .Values (index . "field")) }}
{{- fail (printf "Missing required field: .Values.%s" (index . "field")) }}
{{- end -}}
{{- end }}