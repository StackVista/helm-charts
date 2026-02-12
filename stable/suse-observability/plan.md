# Migration Plan: MinIO to S3Proxy

This document outlines the implementation plan for replacing the MinIO subchart with S3Proxy integrated directly into the SUSE Observability chart.

## Table of Contents

1. [Overview](#overview)
2. [Design Decisions](#design-decisions)
3. [New Values Schema](#new-values-schema)
4. [Template Files to Create](#template-files-to-create)
5. [Template Files to Modify](#template-files-to-modify)
6. [Migration Path](#migration-path)
7. [Testing Strategy](#testing-strategy)
8. [Implementation Tasks](#implementation-tasks)

---

## Overview

### Current State

- MinIO is deployed as a subchart dependency (`stable/minio/`)
- MinIO is conditionally enabled via `global.backup.enabled`
- Backup jobs reference MinIO endpoints via `stackstate.minio.endpoint` helper
- Settings backups use a dedicated PVC (`settings-backup-data`) plus optional S3 upload

### Target State (Option A.1 from minio-migration.md)

- S3Proxy is deployed as a Deployment directly in the suse-observability chart
- S3Proxy provides S3-compatible API for all backup operations
- **Single S3Proxy container with bucket-locator middleware** to route different buckets to different backends:
  - **`settings-local-backup` bucket**: **Always** backed by a dedicated small PVC (~1Gi). This PVC is always present and replaces the current `settings-backup-data` PVC. Settings backups are always written here first.
  - **Main backup buckets** (`sts-backup`, `sts-stackgraph-backup`, etc.): When `global.backup.enabled=true`, S3Proxy routes these to the configured external backend (S3/Azure) or a larger PVC (local persistence mode).
- **Multiple S3Proxy properties files** passed via `--properties` flags:
  - `s3proxy-settings.properties` - Filesystem backend for `settings-local-backup` bucket
  - `s3proxy-main.properties` - Configurable backend (PVC/S3/Azure) for main buckets (only when `global.backup.enabled=true`)
- Single configuration location: `backup.storage.*`
- Backward compatibility with legacy `minio.*` values

**Key insight**: The `settings-local-backup` bucket is **always local** (PVC-backed) regardless of whether `global.backup.enabled` is true or false. When global backup is enabled, settings are written to **both** the local bucket AND uploaded to the remote backend, preserving today's dual-copy safety model.

---

## Design Decisions

### S3Proxy Image

S3Proxy is available on Docker Hub: `andrewgaul/s3proxy:latest` (or tagged releases like `s3proxy-3.0.0`)

We should:
1. Mirror the image to `quay.io/stackstate/s3proxy` for consistency
2. Use a specific version tag (currently `3.0.0`)

### Backend Support

| Backend | S3Proxy Provider | Configuration |
|---------|------------------|---------------|
| Local PVC (filesystem) | `filesystem-nio2` | Data stored at `/data` |
| AWS S3 | `aws-s3-sdk` | Supports IAM roles, IRSA |
| Azure Blob | `azureblob-sdk` | Supports managed identity |
| GCS | `google-cloud-storage` | Via S3 interoperability keys |

### Deployment Model

- **Deployment** (not StatefulSet) with a single replica
- **Single S3Proxy container** using the bucket-locator middleware to route buckets to different backends
- **Two PVCs**:
  - `s3proxy-settings-data` (~1Gi) - Always present, backs the `settings-local-backup` bucket
  - `s3proxy-data` (configurable size) - Only when using PVC backend for main backups
- **One ConfigMap** with multiple properties files:
  - `s3proxy-settings.properties` - Filesystem backend for settings bucket (always local)
  - `s3proxy-main.properties` - Configurable backend (PVC/S3/Azure) for main buckets (only when `global.backup.enabled=true`)
- Secret for credentials (proxy identity/credential + backend credentials)

**Note**: S3Proxy supports the `bucket-locator` middleware which routes requests to different backends based on bucket name. We pass multiple `--properties` files to the S3Proxy container, each defining a backend and which buckets it handles. This allows a single container on port 9000 to serve all buckets.

### Single-Container Architecture with Bucket-Locator

This architecture ensures settings backups are **always available locally** while still supporting remote backup storage for production deployments. A single S3Proxy container handles all requests, routing them to the appropriate backend based on bucket name.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              S3Proxy Pod                                    │
│                                                                             │
│  ┌───────────────────────────────────────────────────────────────────────┐  │
│  │                    s3proxy container (port 9000)                      │  │
│  │                                                                       │  │
│  │   bucket-locator middleware routes by bucket name:                    │  │
│  │                                                                       │  │
│  │   ┌─────────────────────────┐  ┌────────────────────────────────────┐ │  │
│  │   │ s3proxy-settings.properties│  │ s3proxy-main.properties          │ │  │
│  │   │                         │  │ (only when backup enabled)         │ │  │
│  │   │ bucket-locator.1=       │  │                                    │ │  │
│  │   │   settings-local-backup │  │ bucket-locator.1=sts-backup        │ │  │
│  │   │                         │  │ bucket-locator.2=sts-stackgraph-*  │ │  │
│  │   │                         │  │ bucket-locator.3=sts-elasticsearch-*│ │  │
│  │   │ backend: filesystem     │  │ bucket-locator.4=sts-victoria-*    │ │  │
│  │   │ basedir: /settings-data │  │ bucket-locator.5=sts-clickhouse-*  │ │  │
│  │   │                         │  │                                    │ │  │
│  │   └───────────┬─────────────┘  │ backend: filesystem|s3|azure       │ │  │
│  │               │                │ basedir: /main-data (if PVC)       │ │  │
│  │               ▼                └──────────────┬─────────────────────┘ │  │
│  │   ┌─────────────────────┐                     │                       │  │
│  │   │ settings-data PVC   │                     ▼                       │  │
│  │   │ (~1Gi) ALWAYS       │         ┌─────────────────────────┐         │  │
│  │   └─────────────────────┘         │ main-data PVC (500Gi)   │         │  │
│  │                                   │ OR S3/Azure remote      │         │  │
│  │                                   └─────────────────────────┘         │  │
│  └───────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘

Single Service Port: 9000
```

**Scenarios:**

| `global.backup.enabled` | Settings bucket backend | Main buckets backend | PVCs created | Properties files |
|------------------------|------------------------|---------------------|--------------|------------------|
| `false` | Local PVC (settings) | N/A (not served) | settings-data only | settings only |
| `true` + PVC backend | Local PVC (settings) | Local PVC (main) | settings-data + data | settings + main |
| `true` + S3 backend | Local PVC (settings) | Remote S3 | settings-data only | settings + main |
| `true` + Azure backend | Local PVC (settings) | Remote Azure Blob | settings-data only | settings + main |

**Benefits:**
- Settings are **always** stored locally, ensuring restore capability even if remote backend is unavailable
- When `global.backup.enabled=true`, settings are written to both local and remote (dual-copy safety)
- Single container, single port - simpler to manage and monitor
- Bucket-locator middleware handles routing transparently to clients
- No orphaned backups when switching between `global.backup.enabled=true/false`
- Clean migration path from existing settings-backup PVC

---

## New Values Schema

```yaml
backup:
  storage:
    # s3proxy.enabled -- Enable the S3Proxy deployment. Defaults to true when global.backup.enabled is true.
    enabled: true

    image:
      # backup.storage.image.registry -- Image registry for S3Proxy
      registry: quay.io
      # backup.storage.image.repository -- Image repository for S3Proxy
      repository: stackstate/s3proxy
      # backup.storage.image.tag -- Image tag for S3Proxy
      tag: "sha-1281afd"
      # backup.storage.image.pullPolicy -- Image pull policy for S3Proxy
      pullPolicy: IfNotPresent

    # backup.storage.credentials -- Credentials for S3Proxy frontend (used by backup jobs)
    credentials:
      # backup.storage.credentials.accessKey -- Access key for S3Proxy authentication (auto-generated if empty)
      accessKey: ""
      # backup.storage.credentials.secretKey -- Secret key for S3Proxy authentication (auto-generated if empty)
      secretKey: ""
      # backup.storage.credentials.existingSecret -- Use existing secret for credentials (keys: accessKey, secretKey)
      existingSecret: ""

    # backup.storage.settingsPvc -- PVC for local settings backup (settings-local-backup bucket)
    # This PVC is ALWAYS present regardless of global.backup.enabled status.
    # It provides the dual-copy safety model where settings are always stored locally.
    settingsPvc:
      # backup.storage.settingsPvc.size -- Size of the settings backup PVC
      size: 1Gi
      # backup.storage.settingsPvc.storageClass -- Storage class for the settings PVC
      storageClass: ""
      # backup.storage.settingsPvc.accessModes -- Access modes for the settings PVC
      accessModes:
        - ReadWriteOnce

    # backup.storage.backend -- Backend storage configuration for main backup buckets (only one can be enabled)
    # This is only used when global.backup.enabled=true
    backend:
      # backup.storage.backend.pvc -- Use local PVC storage (default when no other backend configured)
      pvc:
        # backup.storage.backend.pvc.enabled -- Enable PVC backend
        enabled: true
        # backup.storage.backend.pvc.size -- Size of the PVC for S3Proxy data
        size: 500Gi
        # backup.storage.backend.pvc.storageClass -- Storage class for the PVC
        storageClass: ""
        # backup.storage.backend.pvc.accessModes -- Access modes for the PVC
        accessModes:
          - ReadWriteOnce

      # backup.storage.backend.s3 -- Use external S3-compatible storage
      s3:
        # backup.storage.backend.s3.enabled -- Enable S3 backend
        enabled: false
        # backup.storage.backend.s3.accessKey -- AWS access key (optional, falls back to instance profile/IRSA)
        accessKey: ""
        # backup.storage.backend.s3.secretKey -- AWS secret key (optional)
        secretKey: ""
        # backup.storage.backend.s3.endpoint -- S3 endpoint URL (optional, defaults to AWS)
        endpoint: ""
        # backup.storage.backend.s3.region -- AWS region (defaults to us-east-1)
        region: "us-east-1"

      # backup.storage.backend.azure -- Use Azure Blob Storage
      azure:
        # backup.storage.backend.azure.enabled -- Enable Azure backend
        enabled: false
        # backup.storage.backend.azure.accountName -- Azure storage account name
        accountName: ""
        # backup.storage.backend.azure.accountKey -- Azure storage account key (optional, falls back to managed identity)
        accountKey: ""
        # backup.storage.backend.azure.endpoint -- Azure blob endpoint (auto-derived from accountName if not set)
        endpoint: ""

    # backup.storage.migration -- Settings for migration from old MinIO installation
    migration:
      # backup.storage.migration.enabled -- Enable migration init container
      enabled: true

# S3Proxy deployment configuration (resources, scheduling, annotations)
s3proxy:
  # s3proxy.resources -- Resource limits and requests for S3Proxy container
  resources:
    limits:
      cpu: 500m
      memory: 512Mi
      ephemeral-storage: 1Gi
    requests:
      cpu: 100m
      memory: 256Mi
      ephemeral-storage: 1Mi
  # s3proxy.nodeSelector -- Node selector for S3Proxy pod (merged with stackstate.components.all.nodeSelector)
  nodeSelector: {}
  # s3proxy.tolerations -- Tolerations for S3Proxy pod (appended to stackstate.components.all.tolerations)
  tolerations: []
  # s3proxy.affinity -- Affinity settings for S3Proxy pod (merged with stackstate.components.all.affinity)
  affinity: {}
  # s3proxy.podAnnotations -- Annotations for S3Proxy pod
  podAnnotations: {}

# Legacy minio values (deprecated, mapped to backup.storage)
minio:
  # minio.accessKey -- DEPRECATED: Use backup.storage.credentials.accessKey
  accessKey: ""
  # minio.secretKey -- DEPRECATED: Use backup.storage.credentials.secretKey
  secretKey: ""
  # minio.fullnameOverride -- Service name override (used for endpoint compatibility)
  fullnameOverride: ""
```

---

## Template Files to Create

### 1. `templates/s3proxy/deployment.yaml`

The deployment runs a **single S3Proxy container** that uses the bucket-locator middleware to route different buckets to different backends. Multiple `--properties` files are passed to configure each backend.

**Important**: This template uses global helpers for image registry, pull secrets, nodeSelector, tolerations, affinity, and securityContext to ensure consistency with other components.

```yaml
{{- if include "stackstate.s3proxy.enabled" . }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "stackstate.s3proxy.fullname" . }}
  labels:
    app.kubernetes.io/name: s3proxy
    app.kubernetes.io/component: s3proxy
    {{- include "suse-observability.labels.global" . | nindent 4 }}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app.kubernetes.io/name: s3proxy
      app.kubernetes.io/component: s3proxy
  template:
    metadata:
      labels:
        app.kubernetes.io/name: s3proxy
        app.kubernetes.io/component: s3proxy
        {{- include "suse-observability.labels.global" . | nindent 8 }}
      annotations:
        checksum/config: {{ include (print $.Template.BasePath "/s3proxy/configmap.yaml") . | sha256sum }}
        checksum/secret: {{ include (print $.Template.BasePath "/s3proxy/secret.yaml") . | sha256sum }}
        {{- with .Values.backup.storage.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
    spec:
      {{- include "stackstate.image.pullSecret.name" (dict "context" $) | nindent 6 }}
      {{- if .Values.stackstate.components.all.securityContext.enabled }}
      securityContext: {{- omit .Values.stackstate.components.all.securityContext "enabled" | toYaml | nindent 8 }}
      {{- end }}
      containers:
        # Single S3Proxy container with bucket-locator middleware
        # Routes different buckets to different backends based on bucket name
        - name: s3proxy
          image: "{{ include "stackstate.s3proxy.image.registry" . }}/{{ .Values.backup.storage.image.repository }}:{{ .Values.backup.storage.image.tag }}"
          imagePullPolicy: {{ default .Values.stackstate.components.all.image.pullPolicy .Values.backup.storage.image.pullPolicy | quote }}
          args:
            # Settings backend (always present) - serves sts-configuration-backup bucket
            - "--properties"
            - "/etc/s3proxy/s3proxy-settings.properties"
            {{- if .Values.global.backup.enabled }}
            # Main backend (only when backup enabled) - serves main backup buckets
            - "--properties"
            - "/etc/s3proxy/s3proxy-main.properties"
            {{- end }}
          ports:
            - name: http
              containerPort: 9000
              protocol: TCP
          env:
            - name: S3PROXY_IDENTITY
              valueFrom:
                secretKeyRef:
                  name: {{ include "stackstate.s3proxy.secretName" . }}
                  key: accessKey
            - name: S3PROXY_CREDENTIAL
              valueFrom:
                secretKeyRef:
                  name: {{ include "stackstate.s3proxy.secretName" . }}
                  key: secretKey
            {{- if and .Values.global.backup.enabled .Values.backup.storage.backend.s3.enabled .Values.backup.storage.backend.s3.accessKey }}
            - name: JCLOUDS_IDENTITY
              valueFrom:
                secretKeyRef:
                  name: {{ include "stackstate.s3proxy.secretName" . }}
                  key: backendAccessKey
            - name: JCLOUDS_CREDENTIAL
              valueFrom:
                secretKeyRef:
                  name: {{ include "stackstate.s3proxy.secretName" . }}
                  key: backendSecretKey
            {{- end }}
            {{- if and .Values.global.backup.enabled .Values.backup.storage.backend.azure.enabled }}
            - name: JCLOUDS_IDENTITY
              valueFrom:
                secretKeyRef:
                  name: {{ include "stackstate.s3proxy.secretName" . }}
                  key: azureAccountName
            {{- if .Values.backup.storage.backend.azure.accountKey }}
            - name: JCLOUDS_CREDENTIAL
              valueFrom:
                secretKeyRef:
                  name: {{ include "stackstate.s3proxy.secretName" . }}
                  key: azureAccountKey
            {{- end }}
            {{- end }}
          volumeMounts:
            - name: config
              mountPath: /etc/s3proxy
              readOnly: true
            - name: settings-data
              mountPath: /settings-data
            {{- if and .Values.global.backup.enabled (include "stackstate.s3proxy.mainBackendUsesPVC" .) }}
            - name: main-data
              mountPath: /main-data
            {{- end }}
          resources:
            {{- toYaml .Values.backup.storage.resources | nindent 12 }}
          livenessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 10
            periodSeconds: 30
          readinessProbe:
            httpGet:
              path: /
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
      {{- include "stackstate.s3proxy.nodeSelector" . | nindent 6 }}
      {{- include "stackstate.s3proxy.affinity" . | nindent 6 }}
      {{- include "stackstate.s3proxy.tolerations" . | nindent 6 }}
      volumes:
        - name: config
          configMap:
            name: {{ include "stackstate.s3proxy.fullname" . }}-config
        # Settings PVC is ALWAYS present
        - name: settings-data
          persistentVolumeClaim:
            claimName: {{ include "stackstate.s3proxy.fullname" . }}-settings-data
        {{- if and .Values.global.backup.enabled (include "stackstate.s3proxy.mainBackendUsesPVC" .) }}
        - name: main-data
          persistentVolumeClaim:
            claimName: {{ include "stackstate.s3proxy.fullname" . }}-data
        {{- end }}
{{- end }}
```

### 2. `templates/s3proxy/service.yaml`

The service exposes a **single port** (9000) for all S3 operations. The bucket-locator middleware handles routing internally.

```yaml
{{- if include "stackstate.s3proxy.enabled" . }}
apiVersion: v1
kind: Service
metadata:
  name: {{ include "stackstate.s3proxy.fullname" . }}
  labels:
    app.kubernetes.io/name: s3proxy
    app.kubernetes.io/component: s3proxy
    {{- include "suse-observability.labels.global" . | nindent 4 }}
spec:
  type: ClusterIP
  ports:
    - name: http
      port: 9000
      targetPort: 9000
      protocol: TCP
  selector:
    app.kubernetes.io/name: s3proxy
    app.kubernetes.io/component: s3proxy
{{- end }}
```

### 3. `templates/s3proxy/configmap.yaml`

One ConfigMap with **multiple properties files** for the bucket-locator middleware. Each backend is configured in a separate file with `s3proxy.bucket-locator.N=bucketname` directives to specify which buckets it handles.

**Important**: The bucket-locator middleware uses glob syntax for bucket names (e.g., `sts-*` matches all buckets starting with `sts-`).

```yaml
{{- if include "stackstate.s3proxy.enabled" . }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "stackstate.s3proxy.fullname" . }}-config
  labels:
    app.kubernetes.io/name: s3proxy
    app.kubernetes.io/component: s3proxy
    {{- include "suse-observability.labels.global" . | nindent 4 }}
data:
  # ===== Settings backend configuration =====
  # Always uses local filesystem (settings PVC) for settings-local-backup bucket
  # This backend is ALWAYS present regardless of global.backup.enabled
  s3proxy-settings.properties: |
    s3proxy.endpoint=http://0.0.0.0:9000
    s3proxy.authorization=aws-v2-or-v4
    # Identity and credential are loaded from environment variables
    # S3PROXY_IDENTITY and S3PROXY_CREDENTIAL

    # Filesystem backend - stores data on the settings PVC
    jclouds.provider=filesystem-nio2
    jclouds.filesystem.basedir=/settings-data

    # Bucket locator: this backend handles the settings bucket only
    s3proxy.bucket-locator.1=settings-local-backup

  {{- if .Values.global.backup.enabled }}
  # ===== Main backup backend configuration =====
  # Backend depends on backup.storage.backend configuration
  # Uses bucket-locator to handle all main backup buckets
  s3proxy-main.properties: |
    s3proxy.endpoint=http://0.0.0.0:9000
    s3proxy.authorization=aws-v2-or-v4
    # Identity and credential are loaded from environment variables
    # S3PROXY_IDENTITY and S3PROXY_CREDENTIAL

    {{- if .Values.backup.storage.backend.pvc.enabled }}
    # Filesystem backend (main PVC)
    jclouds.provider=filesystem-nio2
    jclouds.filesystem.basedir=/main-data
    {{- else if .Values.backup.storage.backend.s3.enabled }}
    # AWS S3 backend
    jclouds.provider=aws-s3-sdk
    {{- if .Values.backup.storage.backend.s3.endpoint }}
    jclouds.endpoint={{ .Values.backup.storage.backend.s3.endpoint }}
    {{- else }}
    jclouds.endpoint=https://s3.{{ .Values.backup.storage.backend.s3.region }}.amazonaws.com
    {{- end }}
    aws-s3-sdk.region={{ .Values.backup.storage.backend.s3.region }}
    aws-s3-sdk.conditional-writes=native
    {{- if not .Values.backup.storage.backend.s3.accessKey }}
    # Using default credentials chain (IAM role, IRSA, etc.)
    jclouds.identity=
    jclouds.credential=
    {{- end }}
    {{- else if .Values.backup.storage.backend.azure.enabled }}
    # Azure Blob Storage backend
    jclouds.provider=azureblob-sdk
    {{- if .Values.backup.storage.backend.azure.endpoint }}
    jclouds.endpoint={{ .Values.backup.storage.backend.azure.endpoint }}
    {{- else }}
    jclouds.endpoint=https://{{ .Values.backup.storage.backend.azure.accountName }}.blob.core.windows.net
    {{- end }}
    {{- if not .Values.backup.storage.backend.azure.accountKey }}
    # Using managed identity
    jclouds.identity=
    jclouds.credential=
    {{- end }}
    {{- end }}

    # Bucket locator: this backend handles all main backup buckets
    # Using glob patterns to match bucket names
    # Note: sts-configuration-backup is included here so settings get backed up
    # to BOTH local (settings backend) AND remote (main backend) for dual-copy safety
    s3proxy.bucket-locator.1={{ .Values.backup.configuration.bucketName }}
    s3proxy.bucket-locator.2={{ .Values.backup.stackGraph.bucketName }}
    s3proxy.bucket-locator.3={{ .Values.backup.elasticsearch.bucketName }}
    s3proxy.bucket-locator.4={{ index .Values "victoria-metrics-0" "backup" "bucketName" }}
    s3proxy.bucket-locator.5={{ index .Values "victoria-metrics-1" "backup" "bucketName" }}
    s3proxy.bucket-locator.6={{ .Values.clickhouse.backup.bucketName }}
  {{- end }}
{{- end }}
```

### 4. `templates/s3proxy/secret.yaml`

```yaml
{{- if include "stackstate.s3proxy.enabled" . }}
{{- if not .Values.backup.storage.credentials.existingSecret }}
apiVersion: v1
kind: Secret
metadata:
  name: {{ include "stackstate.s3proxy.secretName" . }}
  labels:
    app.kubernetes.io/name: s3proxy
    app.kubernetes.io/component: s3proxy
    {{- include "suse-observability.labels.global" . | nindent 4 }}
type: Opaque
data:
  accessKey: {{ include "stackstate.s3proxy.accessKey" . | b64enc | quote }}
  secretKey: {{ include "stackstate.s3proxy.secretKey" . | b64enc | quote }}
  {{- if and .Values.backup.storage.backend.s3.enabled .Values.backup.storage.backend.s3.accessKey }}
  backendAccessKey: {{ .Values.backup.storage.backend.s3.accessKey | b64enc | quote }}
  backendSecretKey: {{ .Values.backup.storage.backend.s3.secretKey | b64enc | quote }}
  {{- end }}
  {{- if .Values.backup.storage.backend.azure.enabled }}
  azureAccountName: {{ .Values.backup.storage.backend.azure.accountName | b64enc | quote }}
  {{- if .Values.backup.storage.backend.azure.accountKey }}
  azureAccountKey: {{ .Values.backup.storage.backend.azure.accountKey | b64enc | quote }}
  {{- end }}
  {{- end }}
{{- end }}
{{- end }}
```

### 5. `templates/s3proxy/pvc.yaml`

This template creates **two PVCs**:
1. `settings-data` - Always present, backs the `sts-configuration-backup` bucket (~1Gi)
2. `data` - Only when using PVC backend for main backups (configurable size)

```yaml
{{- if include "stackstate.s3proxy.enabled" . }}
---
# Settings PVC - ALWAYS present (dual-bucket architecture)
# This PVC backs the sts-configuration-backup bucket and ensures
# settings backups are always stored locally regardless of remote backend config.
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "stackstate.s3proxy.fullname" . }}-settings-data
  labels:
    app.kubernetes.io/name: s3proxy
    app.kubernetes.io/component: s3proxy
    {{- include "suse-observability.labels.global" . | nindent 4 }}
  annotations:
    monitor.kubernetes-v2.stackstate.io/pvc-orphan: |-
      {
        "enabled": false
      }
spec:
  accessModes: {{ .Values.backup.storage.settingsPvc.accessModes | toYaml | nindent 4 }}
  resources:
    requests:
      storage: {{ .Values.backup.storage.settingsPvc.size }}
  {{- if .Values.backup.storage.settingsPvc.storageClass }}
  storageClassName: {{ .Values.backup.storage.settingsPvc.storageClass }}
  {{- else if .Values.global.storageClass }}
  storageClassName: {{ .Values.global.storageClass }}
  {{- end }}

{{- if and .Values.global.backup.enabled (include "stackstate.s3proxy.mainBackendUsesPVC" .) }}
---
# Main backup PVC - Only when global backup enabled with PVC backend
# This PVC stores all main backup data (StackGraph, Elasticsearch, etc.)
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {{ include "stackstate.s3proxy.fullname" . }}-data
  labels:
    app.kubernetes.io/name: s3proxy
    app.kubernetes.io/component: s3proxy
    {{- include "suse-observability.labels.global" . | nindent 4 }}
  annotations:
    monitor.kubernetes-v2.stackstate.io/pvc-orphan: |-
      {
        "enabled": false
      }
spec:
  accessModes: {{ .Values.backup.storage.backend.pvc.accessModes | toYaml | nindent 4 }}
  resources:
    requests:
      storage: {{ .Values.backup.storage.backend.pvc.size }}
  {{- if .Values.backup.storage.backend.pvc.storageClass }}
  storageClassName: {{ .Values.backup.storage.backend.pvc.storageClass }}
  {{- else if .Values.global.storageClass }}
  storageClassName: {{ .Values.global.storageClass }}
  {{- end }}
{{- end }}
{{- end }}
```

### 6. `templates/_helper-s3proxy.tpl`

```yaml
{{/*
Determine if S3Proxy should be enabled.
S3Proxy is ALWAYS enabled because the settings-local PVC needs it.
Even when global.backup.enabled=false, we need S3Proxy to serve the
sts-configuration-backup bucket from the local settings PVC.
*/}}
{{- define "stackstate.s3proxy.enabled" -}}
true
{{- end -}}

{{/*
Full name for S3Proxy resources.
Uses minio.fullnameOverride for backward compatibility.
*/}}
{{- define "stackstate.s3proxy.fullname" -}}
{{- default "suse-observability-s3proxy" .Values.minio.fullnameOverride -}}
{{- end -}}

{{/*
S3Proxy secret name.
*/}}
{{- define "stackstate.s3proxy.secretName" -}}
{{- if .Values.backup.storage.credentials.existingSecret -}}
{{- .Values.backup.storage.credentials.existingSecret -}}
{{- else -}}
{{- include "stackstate.s3proxy.fullname" . -}}
{{- end -}}
{{- end -}}

{{/*
S3Proxy endpoint (single port 9000).
With bucket-locator middleware, all buckets are served from the same endpoint.
The middleware routes requests to the appropriate backend based on bucket name.
*/}}
{{- define "stackstate.s3proxy.endpoint" -}}
{{ include "stackstate.s3proxy.fullname" . }}:9000
{{- end -}}

{{/*
Determine if the main backup backend uses PVC storage.
True when PVC backend is enabled or no other backend is explicitly enabled.
This is separate from the settings PVC which is always present.
*/}}
{{- define "stackstate.s3proxy.mainBackendUsesPVC" -}}
{{- $pvcEnabled := .Values.backup.storage.backend.pvc.enabled -}}
{{- $s3Enabled := .Values.backup.storage.backend.s3.enabled -}}
{{- $azureEnabled := .Values.backup.storage.backend.azure.enabled -}}
{{- if or $pvcEnabled (not (or $s3Enabled $azureEnabled)) -}}
true
{{- end -}}
{{- end -}}

{{/*
Get S3Proxy access key.
Falls back to minio.accessKey for backward compatibility, or generates random key.
*/}}
{{- define "stackstate.s3proxy.accessKey" -}}
{{- if .Values.backup.storage.credentials.accessKey -}}
{{- .Values.backup.storage.credentials.accessKey -}}
{{- else if .Values.minio.accessKey -}}
{{- .Values.minio.accessKey -}}
{{- else -}}
{{- randAlphaNum 20 -}}
{{- end -}}
{{- end -}}

{{/*
Get S3Proxy secret key.
Falls back to minio.secretKey for backward compatibility, or generates random key.
*/}}
{{- define "stackstate.s3proxy.secretKey" -}}
{{- if .Values.backup.storage.credentials.secretKey -}}
{{- .Values.backup.storage.credentials.secretKey -}}
{{- else if .Values.minio.secretKey -}}
{{- .Values.minio.secretKey -}}
{{- else -}}
{{- randAlphaNum 40 -}}
{{- end -}}
{{- end -}}

{{/*
Return the image registry for S3Proxy.
Uses the common.image.registry helper with the backup.storage.image configuration.
*/}}
{{- define "stackstate.s3proxy.image.registry" -}}
{{ include "common.image.registry" (dict "image" .Values.backup.storage.image "context" $) }}
{{- end -}}

{{/*
Merge nodeSelector from stackstate.components.all and backup.storage.
backup.storage.nodeSelector takes precedence over all.nodeSelector.

Usage:
{{- include "stackstate.s3proxy.nodeSelector" . | nindent 6 }}
*/}}
{{- define "stackstate.s3proxy.nodeSelector" -}}
{{- $allNodeSelector := .Values.stackstate.components.all.nodeSelector | default dict -}}
{{- $s3proxyNodeSelector := .Values.backup.storage.nodeSelector | default dict -}}
{{- $merged := merge $s3proxyNodeSelector $allNodeSelector -}}
{{- if $merged -}}
nodeSelector:
  {{- toYaml $merged | nindent 2 }}
{{- end -}}
{{- end -}}

{{/*
Merge affinity from stackstate.components.all and backup.storage.
backup.storage.affinity takes precedence over all.affinity.
Uses the global affinity helper for consistency.

Usage:
{{- include "stackstate.s3proxy.affinity" . | nindent 6 }}
*/}}
{{- define "stackstate.s3proxy.affinity" -}}
{{- $affinity := include "suse-observability.global.affinity" (dict "componentAffinity" .Values.backup.storage.affinity "allAffinity" .Values.stackstate.components.all.affinity "context" .) -}}
{{- if $affinity -}}
affinity:
  {{- $affinity | nindent 2 }}
{{- end -}}
{{- end -}}

{{/*
Concatenate tolerations from stackstate.components.all and backup.storage.
Both lists are combined (backup.storage tolerations are appended after all tolerations).

Usage:
{{- include "stackstate.s3proxy.tolerations" . | nindent 6 }}
*/}}
{{- define "stackstate.s3proxy.tolerations" -}}
{{- $allTolerations := .Values.stackstate.components.all.tolerations | default list -}}
{{- $s3proxyTolerations := .Values.backup.storage.tolerations | default list -}}
{{- $merged := concat $allTolerations $s3proxyTolerations -}}
{{- if $merged -}}
tolerations:
  {{- toYaml $merged | nindent 2 }}
{{- end -}}
{{- end -}}

{{/*
Backward compatibility: Map legacy minio values to new backup.storage values.
This should be called early in templates that need storage configuration.
*/}}
{{- define "stackstate.s3proxy.mapLegacyValues" -}}
{{- /* Map s3gateway to backend.s3 */ -}}
{{- if and .Values.minio.s3gateway .Values.minio.s3gateway.enabled -}}
  {{- $_ := set .Values.backup.storage.backend.s3 "enabled" true -}}
  {{- if .Values.minio.s3gateway.accessKey -}}
    {{- $_ := set .Values.backup.storage.backend.s3 "accessKey" .Values.minio.s3gateway.accessKey -}}
  {{- end -}}
  {{- if .Values.minio.s3gateway.secretKey -}}
    {{- $_ := set .Values.backup.storage.backend.s3 "secretKey" .Values.minio.s3gateway.secretKey -}}
  {{- end -}}
  {{- if .Values.minio.s3gateway.serviceEndpoint -}}
    {{- $_ := set .Values.backup.storage.backend.s3 "endpoint" .Values.minio.s3gateway.serviceEndpoint -}}
  {{- end -}}
{{- end -}}
{{- /* Map azuregateway to backend.azure */ -}}
{{- if and .Values.minio.azuregateway .Values.minio.azuregateway.enabled -}}
  {{- $_ := set .Values.backup.storage.backend.azure "enabled" true -}}
  {{- if .Values.minio.accessKey -}}
    {{- $_ := set .Values.backup.storage.backend.azure "accountName" .Values.minio.accessKey -}}
  {{- end -}}
  {{- if .Values.minio.secretKey -}}
    {{- $_ := set .Values.backup.storage.backend.azure "accountKey" .Values.minio.secretKey -}}
  {{- end -}}
{{- end -}}
{{- end -}}
```

---

## Template Files to Modify

### 1. `templates/_helper-endpoints.tpl`

Replace the MinIO endpoint helper:

```yaml
{{/*
Logic to determine S3Proxy/MinIO endpoint.
*/}}
{{- define "stackstate.minio.endpoint" -}}
{{- include "stackstate.s3proxy.endpoint" . -}}
{{- end -}}

{{/*
Logic to determine S3Proxy/MinIO keys secret name.
*/}}
{{- define "stackstate.minio.keys" -}}
{{- include "stackstate.s3proxy.secretName" . -}}
{{- end -}}
```

### 2. `templates/_helper-backup.tpl`

Update environment variable references (line 102-104):

```yaml
# Change from:
- name: MINIO_ENDPOINT
  value: {{ include "stackstate.minio.endpoint" . | quote }}

# To:
- name: MINIO_ENDPOINT
  value: {{ include "stackstate.s3proxy.endpoint" . | quote }}
- name: S3_ENDPOINT
  value: {{ include "stackstate.s3proxy.endpoint" . | quote }}
```

Update volume mounts (line 112-115):

```yaml
{{- if .Values.global.backup.enabled }}
- name: s3proxy-keys
  mountPath: /aws-keys
{{- end -}}
```

Update volumes (line 126-130):

```yaml
{{- if .Values.global.backup.enabled }}
- name: s3proxy-keys
  secret:
    secretName: {{ include "stackstate.s3proxy.secretName" . }}
{{- end -}}
```

### 3. `templates/configmap-backup-config.yaml`

Update the minio section (line 73-78):

```yaml
  minio:
    enabled: {{ .Values.global.backup.enabled }}
    service:
      name: {{ include "stackstate.s3proxy.fullname" . | quote }}
      port: {{ .Values.backup.storage.service.port | default 9000 }}
      localPortForwardPort: 9000
```

Update elasticsearch endpoint reference (line 141):

```yaml
    endpoint: {{ include "stackstate.s3proxy.endpoint" . | quote }}
```

### 4. `templates/job-backup-init.yaml`

Update wait container endpoint (line 27):

```yaml
/entrypoint -c {{ include "stackstate.es.endpoint" . }},{{ include "stackstate.s3proxy.endpoint" . }} -t 300
```

### 5. `Chart.yaml`

Remove or make conditional the MinIO dependency:

```yaml
# Change the minio dependency condition or remove entirely
# - name: minio
#   repository: file://../minio/
#   version: "8.0.10-stackstate.18"
#   condition: global.backup.enabled
```

**Note:** During transition, you may keep the MinIO dependency but disable it by default, allowing users to opt-in to the old behavior via a flag.

### 6. `values.yaml`

Add the new `backup.storage` section and mark `minio` values as deprecated (see New Values Schema section above).

### 7. `templates/cronjob-backup.yaml`

The configuration backup cronjob currently mounts the `settings-backup-data` PVC directly. This needs to be updated to use S3 API instead. The cronjob should:

1. **Remove the PVC volume mount** - No longer needed since backups go to S3
2. **Add S3Proxy credentials** - Mount the S3Proxy secret for S3 API authentication
3. **Keep the S3 upload logic** - The dual-copy model (local + remote) is now handled by S3Proxy bucket-locator

**Changes required:**

```yaml
# REMOVE this volume mount from the backup container:
# - name: settings-backup-data
#   mountPath: /settings-backup-data

# REMOVE this volume definition:
# - name: settings-backup-data
#   persistentVolumeClaim:
#     claimName: {{ template "common.fullname.short" . }}-settings-backup-data

# ADD S3Proxy credentials as environment variables:
- name: AWS_ACCESS_KEY_ID
  valueFrom:
    secretKeyRef:
      name: {{ include "stackstate.s3proxy.secretName" . }}
      key: accessKey
- name: AWS_SECRET_ACCESS_KEY
  valueFrom:
    secretKeyRef:
      name: {{ include "stackstate.s3proxy.secretName" . }}
      key: secretKey
- name: S3_ENDPOINT
  value: {{ include "stackstate.s3proxy.endpoint" . | quote }}
- name: S3_BUCKET
  value: "settings-local-backup"
```

### 8. `templates/configmap-backup-restore-scripts.yaml`

The backup/restore scripts are embedded in a ConfigMap. These scripts need to be updated to write to S3 instead of the filesystem.

**Current script flow (backup-configuration.sh):**
1. Export settings via API → `/backup-data/settings.sty`
2. Copy to PVC → `/settings-backup-data/settings-<timestamp>.sty`
3. Optionally upload to S3 bucket

**New script flow:**
1. Export settings via API → `/backup-data/settings.sty`
2. Upload directly to S3 → `s3://settings-local-backup/settings-<timestamp>.sty`
3. If `BACKUP_CONFIGURATION_UPLOAD_REMOTE=true`, also upload to main backup bucket

### 9. `scripts/backup-configuration.sh`

This is the main backup script. Key changes:

**Current behavior:**
```bash
# Write to local filesystem
cp "$BACKUP_FILE" "/settings-backup-data/settings-${TIMESTAMP}.sty"

# Optional: upload to S3
if [ "$BACKUP_CONFIGURATION_UPLOAD_REMOTE" = "true" ]; then
  aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/settings-${TIMESTAMP}.sty"
fi
```

**New behavior:**
```bash
# Always write to settings-local-backup bucket via S3 API
aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp "$BACKUP_FILE" \
  "s3://settings-local-backup/settings-${TIMESTAMP}.sty"

# If remote backup enabled, S3Proxy bucket-locator routes to both local AND remote
# No separate upload needed - the settings bucket is configured in both backends
```

**Required environment variables:**
- `AWS_ACCESS_KEY_ID` - S3Proxy access key
- `AWS_SECRET_ACCESS_KEY` - S3Proxy secret key
- `S3_ENDPOINT` - S3Proxy endpoint (e.g., `suse-observability-s3proxy:9000`)

### 10. `scripts/restore-configuration-backup.sh`

This script reads backups from storage. Key changes:

**Current behavior:**
```bash
# List backups from local filesystem
ls /settings-backup-data/*.sty

# Read backup file
cat "/settings-backup-data/${BACKUP_FILE}"
```

**New behavior:**
```bash
# List backups from S3 bucket
aws --endpoint-url "http://${S3_ENDPOINT}" s3 ls "s3://settings-local-backup/"

# Download backup file from S3
aws --endpoint-url "http://${S3_ENDPOINT}" s3 cp \
  "s3://settings-local-backup/${BACKUP_FILE}" /tmp/restore.sty
```

### 11. `scripts/list-configuration-backups.sh`

This script lists available backups. Key changes:

**Current behavior:**
```bash
# List from filesystem
ls -la /settings-backup-data/*.sty
```

**New behavior:**
```bash
# List from S3 bucket
aws --endpoint-url "http://${S3_ENDPOINT}" s3 ls "s3://settings-local-backup/" \
  --human-readable
```

---

## Migration Path

### Phase 1: Add S3Proxy alongside MinIO (Current Release)

1. Add S3Proxy templates with feature flag `backup.storage.useS3Proxy: false` (default)
2. Keep MinIO as default
3. Document migration path for users

### Phase 2: S3Proxy as Default (Next Release)

1. Change default to `backup.storage.useS3Proxy: true`
2. Deprecate MinIO values with warnings
3. Add migration init container for settings backup PVC

### Phase 3: Remove MinIO (Future Release)

1. Remove MinIO dependency from Chart.yaml
2. Remove legacy value mappings
3. Clean up deprecated helpers

### Data Migration Init Container

For Phase 2, add an init container to the S3Proxy deployment that:

1. Checks if old settings-backup PVC exists and has data
2. If S3Proxy PVC is empty, copies `.sty` files to S3Proxy bucket structure
3. Records migration completion in a marker file

```yaml
initContainers:
  - name: migrate-settings
    image: "{{ .Values.backup.storage.image.registry }}/{{ .Values.backup.storage.image.repository }}:{{ .Values.backup.storage.image.tag }}"
    command:
      - /bin/sh
      - -c
      - |
        # Check if migration is needed
        if [ -f /data/.migrated ]; then
          echo "Migration already completed"
          exit 0
        fi

        # Check if old PVC has data
        if [ -d /old-settings ] && [ "$(ls -A /old-settings/*.sty 2>/dev/null)" ]; then
          echo "Migrating settings from old PVC..."
          mkdir -p /data/sts-configuration-backup
          cp /old-settings/*.sty /data/sts-configuration-backup/
          echo "Migration completed"
        fi

        touch /data/.migrated
    volumeMounts:
      - name: data
        mountPath: /data
      - name: old-settings
        mountPath: /old-settings
        readOnly: true
```

---

## Testing Strategy

### Unit Tests

Create new test files:

1. `test/s3proxy_test.go` - Test S3Proxy deployment rendering
2. `test/s3proxy_backend_test.go` - Test backend configuration variants

Example test structure:

```go
func TestS3ProxyDeploymentDefault(t *testing.T) {
    output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
        ValuesFiles: []string{"values/full.yaml"},
        SetValues: map[string]string{
            "global.backup.enabled": "true",
        },
    })
    resources := helmtestutil.NewKubernetesResources(t, output)

    // Verify S3Proxy deployment exists with single container (bucket-locator architecture)
    deployment, ok := resources.Deployments["suse-observability-s3proxy"]
    require.True(t, ok, "S3Proxy deployment should exist")
    assert.Len(t, deployment.Spec.Template.Spec.Containers, 1, "Should have single container with bucket-locator")

    // Verify container has both properties files passed as args
    container := deployment.Spec.Template.Spec.Containers[0]
    assert.Equal(t, "s3proxy", container.Name)
    assert.Contains(t, container.Args, "--properties")
    assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-settings.properties")
    assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-main.properties")

    // Verify ConfigMap has two properties files
    configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
    require.True(t, ok, "S3Proxy configmap should exist")
    assert.Contains(t, configMap.Data, "s3proxy-settings.properties", "Should have settings properties")
    assert.Contains(t, configMap.Data, "s3proxy-main.properties", "Should have main properties")

    // Verify settings backend uses filesystem (always local)
    assert.Contains(t, configMap.Data["s3proxy-settings.properties"], "jclouds.provider=filesystem-nio2")
    assert.Contains(t, configMap.Data["s3proxy-settings.properties"], "s3proxy.bucket-locator.1=settings-local-backup")

    // Verify PVC backend is default for main
    assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.provider=filesystem-nio2")
    assert.Contains(t, configMap.Data["s3proxy-main.properties"], "s3proxy.bucket-locator.1=")

    // Verify both PVCs are created when backup enabled with PVC backend
    _, settingsPvcOk := resources.PersistentVolumeClaims["suse-observability-s3proxy-settings-data"]
    _, mainPvcOk := resources.PersistentVolumeClaims["suse-observability-s3proxy-data"]
    assert.True(t, settingsPvcOk, "Settings PVC should exist")
    assert.True(t, mainPvcOk, "Main PVC should exist with PVC backend")
}

func TestS3ProxyBackupDisabled(t *testing.T) {
    // When backup is disabled, S3Proxy still runs but only serves settings bucket
    output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
        ValuesFiles: []string{"values/full.yaml"},
        SetValues: map[string]string{
            "global.backup.enabled": "false",
        },
    })
    resources := helmtestutil.NewKubernetesResources(t, output)

    // S3Proxy deployment should still exist (settings bucket always needed)
    deployment, ok := resources.Deployments["suse-observability-s3proxy"]
    require.True(t, ok, "S3Proxy deployment should exist")
    assert.Len(t, deployment.Spec.Template.Spec.Containers, 1, "Should have single container")

    // Verify only settings properties file is passed
    container := deployment.Spec.Template.Spec.Containers[0]
    assert.Contains(t, container.Args, "/etc/s3proxy/s3proxy-settings.properties")
    assert.NotContains(t, container.Args, "/etc/s3proxy/s3proxy-main.properties")

    // ConfigMap should only have settings properties
    configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
    require.True(t, ok)
    assert.Contains(t, configMap.Data, "s3proxy-settings.properties", "Should have settings properties")
    assert.NotContains(t, configMap.Data, "s3proxy-main.properties", "Should NOT have main properties")

    // Only settings PVC should exist
    _, settingsPvcOk := resources.PersistentVolumeClaims["suse-observability-s3proxy-settings-data"]
    _, mainPvcOk := resources.PersistentVolumeClaims["suse-observability-s3proxy-data"]
    assert.True(t, settingsPvcOk, "Settings PVC should always exist")
    assert.False(t, mainPvcOk, "Main PVC should NOT exist when backup disabled")
}

func TestS3ProxyS3Backend(t *testing.T) {
    output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
        ValuesFiles: []string{"values/full.yaml"},
        SetValues: map[string]string{
            "global.backup.enabled":                   "true",
            "backup.storage.backend.pvc.enabled":     "false",
            "backup.storage.backend.s3.enabled":      "true",
            "backup.storage.backend.s3.region":       "eu-west-1",
        },
    })
    resources := helmtestutil.NewKubernetesResources(t, output)

    configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
    require.True(t, ok)
    // Settings should still use filesystem (always local)
    assert.Contains(t, configMap.Data["s3proxy-settings.properties"], "jclouds.provider=filesystem-nio2")
    assert.Contains(t, configMap.Data["s3proxy-settings.properties"], "jclouds.filesystem.basedir=/settings-data")
    // Main should use S3
    assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.provider=aws-s3-sdk")
    assert.Contains(t, configMap.Data["s3proxy-main.properties"], "aws-s3-sdk.region=eu-west-1")

    // With S3 backend, only settings PVC should exist (no main PVC)
    _, settingsPvcOk := resources.PersistentVolumeClaims["suse-observability-s3proxy-settings-data"]
    _, mainPvcOk := resources.PersistentVolumeClaims["suse-observability-s3proxy-data"]
    assert.True(t, settingsPvcOk, "Settings PVC should always exist")
    assert.False(t, mainPvcOk, "Main PVC should NOT exist with S3 backend")
}

func TestS3ProxyAzureBackend(t *testing.T) {
    output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
        ValuesFiles: []string{"values/full.yaml"},
        SetValues: map[string]string{
            "global.backup.enabled":                        "true",
            "backup.storage.backend.pvc.enabled":          "false",
            "backup.storage.backend.azure.enabled":        "true",
            "backup.storage.backend.azure.accountName":    "mystorageaccount",
        },
    })
    resources := helmtestutil.NewKubernetesResources(t, output)

    configMap, ok := resources.ConfigMaps["suse-observability-s3proxy-config"]
    require.True(t, ok)
    // Settings should still use filesystem (always local)
    assert.Contains(t, configMap.Data["s3proxy-settings.properties"], "jclouds.provider=filesystem-nio2")
    // Main should use Azure
    assert.Contains(t, configMap.Data["s3proxy-main.properties"], "jclouds.provider=azureblob-sdk")
    assert.Contains(t, configMap.Data["s3proxy-main.properties"], "mystorageaccount.blob.core.windows.net")
}

func TestS3ProxyLegacyMinioValues(t *testing.T) {
    // Test backward compatibility with minio.* values
    output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
        ValuesFiles: []string{"values/full.yaml"},
        SetValues: map[string]string{
            "global.backup.enabled":        "true",
            "minio.accessKey":              "legacyAccessKey",
            "minio.secretKey":              "legacySecretKey",
            "minio.fullnameOverride":       "my-custom-minio",
        },
    })
    resources := helmtestutil.NewKubernetesResources(t, output)

    // Should use fullnameOverride for resource naming (backward compat)
    deployment, ok := resources.Deployments["my-custom-minio"]
    require.True(t, ok, "Deployment should use minio.fullnameOverride name")

    secret, ok := resources.Secrets["my-custom-minio"]
    require.True(t, ok, "Secret should use minio.fullnameOverride name")
    // Verify legacy credentials are used (would need to decode base64 to verify actual values)
}

func TestS3ProxySingleService(t *testing.T) {
    // Verify single service on port 9000 for all buckets
    output := helmtestutil.RenderHelmTemplateOptsNoError(t, "suse-observability", &helm.Options{
        ValuesFiles: []string{"values/full.yaml"},
        SetValues: map[string]string{
            "global.backup.enabled": "true",
        },
    })
    resources := helmtestutil.NewKubernetesResources(t, output)

    service, ok := resources.Services["suse-observability-s3proxy"]
    require.True(t, ok, "S3Proxy service should exist")
    assert.Len(t, service.Spec.Ports, 1, "Should have single port")
    assert.Equal(t, int32(9000), service.Spec.Ports[0].Port, "Port should be 9000")
}
```

### Test Value Files

Create test value files in `test/values/`:

- `s3proxy_default.yaml` - Default PVC backend
- `s3proxy_s3_backend.yaml` - S3 backend configuration
- `s3proxy_azure_backend.yaml` - Azure backend configuration
- `s3proxy_legacy_minio.yaml` - Legacy MinIO values for backward compatibility

### Integration Tests

1. Deploy with PVC backend, verify bucket creation works
2. Deploy with S3 backend (LocalStack), verify proxy functionality
3. Test backup/restore workflows end-to-end

---

## Implementation Tasks

### High Priority

- [x] Create `templates/s3proxy/` directory structure
- [x] Implement `_helper-s3proxy.tpl` helper functions:
  - `stackstate.s3proxy.enabled` - Always returns true (settings bucket always needed)
  - `stackstate.s3proxy.fullname` - Resource naming with `minio.fullnameOverride` backward compatibility
  - `stackstate.s3proxy.secretName` - Secret name resolution
  - `stackstate.s3proxy.endpoint` - Single endpoint (port 9000) for all buckets
  - `stackstate.s3proxy.mainBackendUsesPVC` - Checks if main backend uses PVC
  - `stackstate.s3proxy.accessKey` / `secretKey` - Credential resolution with legacy fallback
  - `stackstate.s3proxy.image.registry` - Image registry using common helper
  - `stackstate.s3proxy.nodeSelector` / `affinity` / `tolerations` - Scheduling helpers
- [x] Create S3Proxy Deployment template:
  - Single container with bucket-locator middleware
  - Multiple `--properties` args passed to s3proxy
  - Single port 9000
  - Settings properties file always present
  - Main properties file conditional on `global.backup.enabled`
- [x] Create S3Proxy Service template (single port 9000)
- [x] Create S3Proxy ConfigMap template:
  - `s3proxy-settings.properties` - Filesystem backend for `settings-local-backup` bucket
  - `s3proxy-main.properties` - Configurable backend (PVC/S3/Azure) for main buckets (conditional)
  - Uses `s3proxy.bucket-locator.N=bucketname` directives for routing
- [x] Create S3Proxy Secret template
- [x] Create S3Proxy PVC template:
  - `settings-data` PVC (~1Gi) - Always present
  - `data` PVC (configurable) - Only when using PVC backend for main backups
- [x] Update `_helper-endpoints.tpl` for S3Proxy
- [x] Update `_helper-backup.tpl` for S3Proxy
- [x] Add `backup.storage` values to `values.yaml` (including `settingsPvc` section)
- [x] Update `configmap-backup-config.yaml`
- [x] Update `job-backup-init.yaml`
- [x] Update settings backup cronjob and scripts:
  - Update `templates/cronjob-backup.yaml` - Remove PVC mount, add S3 credentials
  - Update `scripts/backup-configuration.sh` - Write to S3 bucket instead of PVC
  - Update `scripts/restore-configuration-backup.sh` - Read from S3 bucket
  - Update `scripts/list-configuration-backups.sh` - List from S3 bucket
  - Update `scripts/download-configuration-backup.sh` - Download from S3 bucket
  - Update `templates/configmap-backup-restore-scripts.yaml` - Updated all scripts

### Medium Priority

- [x] Add backward compatibility for `minio.*` values (mapLegacyValues helper)
- [x] Create migration init container (copy from old settings-backup PVC)
- [x] Update Chart.yaml dependencies (removed MinIO subchart)
- [x] Write unit tests for S3Proxy templates:
  - `TestS3ProxyAlwaysEnabled` - Verify S3Proxy is always deployed
  - `TestS3ProxyWithBackupEnabledPVC` - Verify PVC backend configuration
  - `TestS3ProxyWithBackupEnabledS3` - Verify S3 backend configuration
  - `TestS3ProxyWithBackupEnabledAzure` - Verify Azure backend configuration
  - `TestS3ProxyLegacyS3GatewayBackwardCompatibility` - Verify legacy s3gateway values work
  - `TestS3ProxyLegacyAzureGatewayBackwardCompatibility` - Verify legacy azuregateway values work
  - `TestS3ProxyConfigMapBuckets` - Verify bucket-locator configuration
  - `TestS3ProxyService` - Verify single port 9000 service
  - `TestS3ProxyFullnameOverride` - Verify backward compat with minio.fullnameOverride
  - `TestS3ProxyResources` - Verify resource configuration
  - `TestS3ProxyDeploymentArgs` - Verify properties file arguments
  - `TestS3ProxyExistingSecret` - Verify existing secret configuration
  - `TestS3ProxyNodeSelector` - Verify node selector merging
  - `TestS3ProxyTolerations` - Verify tolerations merging
  - `TestS3ProxySettingsPVCSize` - Verify settings PVC configuration
  - `TestS3ProxyMainPVCSize` - Verify main PVC configuration
- [x] Update existing tests for S3Proxy changes:
  - Updated `TestGlobalBackupDisabledEnsureResources` - S3Proxy deployment is now always enabled
  - Updated `TestGetImages` - Updated image count (33 after removing MinIO subchart)
  - Updated `TestK8sAuthz*` tests - Updated role counts (removed MinIO role)
  - Added S3Proxy resources to always-enabled lists
- [x] Write tests for PVC creation scenarios

### Low Priority

- [x] Update documentation (README.md.gotmpl) - Added Backup Storage Configuration section
- [x] Add deprecation warnings for minio values - Added to ConfigMap annotations
- [ ] Create migration guide for users (optional - documentation in README covers basics)
- [ ] Performance testing with large backups (optional - to be done during QA)

---

## Design Decisions (Resolved)

Based on project discussion, the following decisions have been made:

1. **Image hosting**: S3Proxy image will be mirrored to a separate repository (quay.io/stackstate/s3proxy)
2. **PVC migration**: Reuse the existing MinIO PVC where possible (to be validated during testing)
3. **Bucket creation**: No separate bucket creation job needed - S3Proxy auto-creates buckets on first write
4. **Health checks**: HTTP GET on `/` is acceptable for S3Proxy health checks
5. **Resource defaults**: The proposed resource limits are appropriate for production workloads
6. **Settings bucket naming**: Use `settings-local-backup` as the settings bucket name to distinguish it from remote backup buckets
7. **Bucket-locator routing**: All bucket names used by the chart must be explicitly defined in the S3Proxy configuration to ensure proper routing
