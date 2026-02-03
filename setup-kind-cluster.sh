#!/bin/bash

# setup-kind-cluster.sh - Setup Kind cluster and deploy Tesla HTTP Proxy

set -e

echo "🚀 Kind Cluster Setup for Tesla HTTP Proxy"
echo "=========================================="
echo ""

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command -v kind &> /dev/null; then
    echo "❌ kind is not installed"
    echo "📖 Install from: https://kind.sigs.k8s.io/docs/user/quick-start/#installation"
    exit 1
fi
echo "✅ kind found"

if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed"
    echo "📖 Install from: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi
echo "✅ kubectl found"

if ! command -v helm &> /dev/null; then
    echo "❌ helm is not installed"
    echo "📖 Install from: https://helm.sh/docs/intro/install/"
    exit 1
fi
echo "✅ helm found"

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    echo "📖 Install from: https://www.docker.com/products/docker-desktop"
    exit 1
fi
echo "✅ Docker found"

echo ""

# Create config directory
mkdir -p config

# Check/generate TLS certificates
if [ ! -f config/tls-cert.pem ] || [ ! -f config/tls-key.pem ]; then
    echo "🔐 Generating TLS certificates..."
    
    # Create OpenSSL config file for compatibility with Windows
    cat > config/openssl.conf <<'OPENSSL_EOF'
[req]
default_bits = 2048
prompt = no
default_md = sha256
distinguished_name = req_distinguished_name
req_extensions = v3_req

[req_distinguished_name]
CN = tesla-proxy

[v3_req]
extendedKeyUsage = serverAuth
keyUsage = digitalSignature, keyCertSign, keyAgreement
subjectAltName = DNS:tesla-proxy,DNS:tesla-proxy.default,DNS:tesla-proxy.default.svc,DNS:tesla-proxy.default.svc.cluster.local
OPENSSL_EOF

    openssl req -x509 -nodes -newkey ec \
        -pkeyopt ec_paramgen_curve:secp384r1 \
        -pkeyopt ec_param_enc:named_curve \
        -config config/openssl.conf \
        -keyout config/tls-key.pem \
        -out config/tls-cert.pem \
        -sha256 -days 3650
    
    rm config/openssl.conf
    echo "✅ TLS certificates generated"
else
    echo "✅ TLS certificates found"
fi

# Check private key
if [ ! -f config/fleet-key.pem ]; then
    echo "❌ fleet-key.pem not found"
    echo "📝 Run: cp private-key.pem config/fleet-key.pem"
    exit 1
fi
echo "✅ Fleet private key found"

echo ""

# Create/recreate kind cluster
CLUSTER_NAME="tesla-proxy"
echo "🎯 Creating Kind cluster: $CLUSTER_NAME..."

# Check if cluster already exists
if kind get clusters | grep -q "^$CLUSTER_NAME$"; then
    echo "⚠️  Cluster already exists. Using existing cluster..."
else
    kind create cluster --name $CLUSTER_NAME
    echo "✅ Cluster created"
fi

echo ""

# Set kubeconfig context
kubectl cluster-info --context kind-$CLUSTER_NAME

echo ""

# Create Helm chart values file
echo "📝 Creating Helm chart values..."
mkdir -p helm/tesla-proxy

cat > helm/tesla-proxy/values.yaml <<'EOF'
replicaCount: 1

image:
  repository: tesla/vehicle-command
  tag: latest
  pullPolicy: IfNotPresent

service:
  type: NodePort
  port: 4443
  targetPort: 4443
  nodePort: 30443

ingress:
  enabled: false

resources:
  limits:
    cpu: 500m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

env:
  - name: TESLA_VERBOSE
    value: "true"

tlsCert: |
  # Will be replaced with actual cert
tlsKey: |
  # Will be replaced with actual key
privateKey: |
  # Will be replaced with actual key
EOF

cat > helm/tesla-proxy/Chart.yaml <<'EOF'
apiVersion: v2
name: tesla-proxy
description: Tesla Vehicle Command HTTP Proxy
type: application
version: 0.1.0
appVersion: "0.4.0"
EOF

# Create templates directory
mkdir -p helm/tesla-proxy/templates

cat > helm/tesla-proxy/templates/deployment.yaml <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "tesla-proxy.fullname" . }}
  labels:
    {{- include "tesla-proxy.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "tesla-proxy.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "tesla-proxy.selectorLabels" . | nindent 8 }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - name: https
          containerPort: 4443
          protocol: TCP
        args:
        - "-tls-key"
        - "/etc/tesla-proxy/tls/tls-key.pem"
        - "-cert"
        - "/etc/tesla-proxy/tls/tls-cert.pem"
        - "-key-file"
        - "/etc/tesla-proxy/secrets/fleet-key.pem"
        - "-host"
        - "0.0.0.0"
        - "-port"
        - "4443"
        volumeMounts:
        - name: tls
          mountPath: /etc/tesla-proxy/tls
          readOnly: true
        - name: private-key
          mountPath: /etc/tesla-proxy/secrets
          readOnly: true
        env:
        {{- range .Values.env }}
        - name: {{ .name }}
          value: {{ .value | quote }}
        {{- end }}
        livenessProbe:
          httpGet:
            path: /health
            port: 4443
            scheme: HTTPS
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 4443
            scheme: HTTPS
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          {{- toYaml .Values.resources | nindent 12 }}
      volumes:
      - name: tls
        secret:
          secretName: tesla-proxy-tls
      - name: private-key
        secret:
          secretName: tesla-proxy-keys
EOF

cat > helm/tesla-proxy/templates/service.yaml <<'EOF'
apiVersion: v1
kind: Service
metadata:
  name: {{ include "tesla-proxy.fullname" . }}
  labels:
    {{- include "tesla-proxy.labels" . | nindent 4 }}
spec:
  type: {{ .Values.service.type }}
  ports:
    - port: {{ .Values.service.port }}
      targetPort: https
      protocol: TCP
      name: https
      {{- if and (eq .Values.service.type "NodePort") .Values.service.nodePort }}
      nodePort: {{ .Values.service.nodePort }}
      {{- end }}
  selector:
    {{- include "tesla-proxy.selectorLabels" . | nindent 4 }}
EOF

cat > helm/tesla-proxy/templates/_helpers.tpl <<'EOF'
{{/*
Expand the name of the chart.
*/}}
{{- define "tesla-proxy.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "tesla-proxy.fullname" -}}
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
{{- define "tesla-proxy.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "tesla-proxy.labels" -}}
helm.sh/chart: {{ include "tesla-proxy.chart" . }}
{{ include "tesla-proxy.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tesla-proxy.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tesla-proxy.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}
EOF

echo "✅ Helm chart created"
echo ""

# Create namespace
echo "📦 Creating Kubernetes namespace..."
kubectl create namespace tesla-proxy --context kind-$CLUSTER_NAME 2>/dev/null || echo "Namespace already exists"

echo ""

# Create secrets from files
echo "🔑 Creating Kubernetes secrets..."
kubectl create secret tls tesla-proxy-tls \
    --cert=config/tls-cert.pem \
    --key=config/tls-key.pem \
    -n tesla-proxy \
    --context kind-$CLUSTER_NAME \
    --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null

kubectl create secret generic tesla-proxy-keys \
    --from-file=fleet-key.pem=config/fleet-key.pem \
    -n tesla-proxy \
    --context kind-$CLUSTER_NAME \
    --dry-run=client -o yaml | kubectl apply -f - 2>/dev/null

echo "✅ Secrets created"
echo ""

# Deploy using Helm
echo "🚀 Deploying Tesla Proxy with Helm..."
helm upgrade --install tesla-proxy helm/tesla-proxy \
    --namespace tesla-proxy \
    --context kind-$CLUSTER_NAME

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📊 Checking deployment status..."
kubectl get all -n tesla-proxy --context kind-$CLUSTER_NAME

echo ""
echo "🔗 To access the proxy:"
echo "  kubectl port-forward -n tesla-proxy service/tesla-proxy 4443:4443 --context kind-$CLUSTER_NAME"
echo ""
echo "📝 Commands:"
echo "  - View logs: kubectl logs -n tesla-proxy -l app.kubernetes.io/instance=tesla-proxy --context kind-$CLUSTER_NAME"
echo "  - Delete cluster: kind delete cluster --name $CLUSTER_NAME"
