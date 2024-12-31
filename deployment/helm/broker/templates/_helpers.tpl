{{- define "template.name" -}} 
{{- default .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "template.chart" -}} 
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

#https://helm.sh/docs/chart_best_practices/labels/
{{- define "template.labels" -}} 
helm.sh/chart: {{ include "template.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{ include "template.selectorLabels" . }}
{{- end }}

{{- define "template.selectorLabels" -}} 
app.kubernetes.io/name: {{ include "template.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
