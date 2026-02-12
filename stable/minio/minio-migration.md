# Migration from Minio AND settings-backup-pvc -> S3Proxy

This migration replaces the MinIO subchart with an always-on S3-compatible proxy (S3Proxy) and moves all backup/restore traffic through that proxy.

## Target design (decision)

* Use **S3Proxy** as the always-present gateway.
* Settings backup defaults to S3Proxy using a small PVC-backed bucket (e.g. 1Gi) when global backup is disabled.
* When `global.backup.enabled=true`, S3Proxy proxies all buckets to the configured external S3 backend so backups land in the configured provider.
* Single place for authentication/authorization (S3Proxy), which also supports AWS instance profiles / IRSA without duplicating credentials across components.
* When versitygw adds support for S3 isntance profiles etc, it should be easy to switch to that instead

## Scope and compatibility

* Primary backend interface is **S3-compatible APIs**.
* **Azure Blob Storage** remains supported by using S3Proxy with an Azure backend (gateway).
* **GCP** remains supported via S3 interoperability keys (XML API).

## Migration of configuration

```
global:
  backup:
    enabled: true
backup:
  stackGraph:
    bucketName: AWS_STACKGRAPH_BUCKET
  elasticsearch:
    bucketName: AWS_ELASTICSEARCH_BUCKET
  configuration:
    bucketName: AWS_CONFIGURATION_BUCKET
victoria-metrics-0:
  backup:
    bucketName: AWS_VICTORIA_METRICS_BUCKET
victoria-metrics-1:
  backup:
    bucketName: AWS_VICTORIA_METRICS_BUCKET
clickhouse:
  backup:
    bucketName: AWS_CLICKHOUSE_BUCKET
minio:
  accessKey: YOUR_ACCESS_KEY
  secretKey: YOUR_SECRET_KEY
  s3gateway:
    enabled: true
    accessKey: AWS_ACCESS_KEY
    secretKey: AWS_SECRET_KEY
```

Alternative config using local PVC or Azure Blob Storage as backend for settings backup:
```
minio:
  accessKey: AZURE_STORAGE_ACCOUNT_NAME
  secretKey: AZURE_STORAGE_ACCOUNT_KEY
  azuregateway:
    enabled: true

minio:
  accessKey: YOUR_ACCESS_KEY
  secretKey: YOUR_SECRET_KEY
  persistence:
    enabled: true
```

New configuration options:
* The backup subsections remain the same.
* We drop the `minio` section and introduce `backup.storage` as the single, solution-agnostic place to configure the proxy and its backend.

```
backup:
  storage:
    credentials:
      accessKey: <make-up-an-access-key>
      secretKey: <make-up-a-secret-key>
    backend:
      # One of the following options can be enabled
      pvc:
        size: 500Gi
      s3:
        # Optional: if access key and secret key are not provided, the proxy will fallback to default authentication flow (based on instance profile, etc.)
        accessKey: AWS_ACCESS_KEY
        secretKey: AWS_SECRET_KEY
        # Optional: if not provided, the proxy will fallback to the default AWS endpoint for the configured region
        endpoint:
        # Optional: if not provided, the proxy will fallback to the default AWS region
        region: eu-west-1
      azure:
        accountName: AZURE_STORAGE_ACCOUNT_NAME
        accountKey: AZURE_STORAGE_ACCOUNT_KEY
```

Backward compatibility can be preserved by accepting the legacy `minio.*` values and mapping them to `backup.storage.*` on render. For Azure, legacy credentials can be reused for both proxy and backend, but the strongly recommended setup is to define distinct proxy credentials.

### Mapping: old `minio.*` to new `backup.storage.*`

| Old value | New value | Notes |
| --- | --- | --- |
| `minio.accessKey` | `backup.storage.credentials.accessKey` | Proxy (frontend) credentials used by all backup/restore jobs. |
| `minio.secretKey` | `backup.storage.credentials.secretKey` | Proxy (frontend) credentials used by all backup/restore jobs. |
| `minio.persistence.enabled` | `backup.storage.backend.pvc` | If enabled, use PVC backend for proxy storage. Size comes from `backup.storage.backend.pvc.size`. |
| `minio.s3gateway.enabled` | `backup.storage.backend.s3` | Use external S3 backend through the proxy. |
| `minio.s3gateway.accessKey` | `backup.storage.backend.s3.accessKey` | Optional; falls back to instance profile / default SDK flow. |
| `minio.s3gateway.secretKey` | `backup.storage.backend.s3.secretKey` | Optional. |
| `minio.s3gateway.endpoint` | `backup.storage.backend.s3.endpoint` | Optional. |
| `minio.s3gateway.region` | `backup.storage.backend.s3.region` | Optional. |
| `minio.azuregateway.enabled` | `backup.storage.backend.azure` | Azure Blob backend through the proxy. |
| `minio.azuregateway.accountName` | `backup.storage.backend.azure.accountName` | Required when Azure backend is used. |
| `minio.azuregateway.accountKey` | `backup.storage.backend.azure.accountKey` | Required when Azure backend is used. |

### Current chart usage that must be updated

The following areas reference MinIO explicitly and should be updated to use the new proxy endpoint and credentials:

* Backup env vars use `MINIO_ENDPOINT` and MinIO keys: [stable/suse-observability/templates/_helper-backup.tpl](stable/suse-observability/templates/_helper-backup.tpl)
* Backup config map has a `minio` section: [stable/suse-observability/templates/configmap-backup-config.yaml](stable/suse-observability/templates/configmap-backup-config.yaml)
* Helper endpoints and secret naming are MinIO-specific: [stable/suse-observability/templates/_helper-endpoints.tpl](stable/suse-observability/templates/_helper-endpoints.tpl)
* Values still require `minio.accessKey` / `minio.secretKey`: [stable/suse-observability/values.yaml](stable/suse-observability/values.yaml)

## Migration of data

### Current setup (before migration)

Two independent storage paths exist for settings (configuration) backups:

| Storage | Created when | What's on it | Retention |
| --- | --- | --- | --- |
| **Settings-backup PVC** (`<release>-settings-backup-data`, ~1Gi) | Always (regardless of `global.backup.enabled`) | Last N `.sty` files (max `backup.configuration.maxLocalFiles`, default 10) | Rolling, up to 10 files |
| **MinIO / external S3** (`sts-configuration-backup` bucket) | Only when `global.backup.enabled=true` | Same `.sty` files, uploaded after each backup | Independent retention (default 365 days) |

The backup CronJob always writes to the PVC first. If `BACKUP_CONFIGURATION_UPLOAD_REMOTE=true` (i.e. `global.backup.enabled=true`), it also uploads to MinIO/S3. Restores check the PVC first, then fall back to S3.

Other backup types (StackGraph, Elasticsearch, Victoria Metrics, ClickHouse) only go through MinIO/S3 and are not affected by the settings-backup PVC.

### Goal

* Remove the settings-backup PVC as a directly-mounted volume from backup jobs.
* Route **all** backup traffic (including settings) through S3Proxy.
* Keep at least one copy of every existing backup during the upgrade.

---

### Option A: S3Proxy-only with a local PVC backend

S3Proxy always uses a PVC as its backend. Settings backups go to the `sts-configuration-backup` bucket on S3Proxy, which stores them on its own PVC. When `global.backup.enabled=true`, S3Proxy proxies the bucket to the external S3 provider instead.

**New behavior:**
* Settings backups always go through S3Proxy, never directly to a PVC.
* When `global.backup.enabled=false`: S3Proxy stores to a local PVC (replaces the old settings-backup PVC).
* When `global.backup.enabled=true`: S3Proxy proxies to the external backend; no local PVC needed for settings.
* Only one copy of settings backups exists (either local PVC via proxy **or** remote S3).

**Migration:**
* `global.backup.enabled=false`:
  * A one-off init container/job copies `.sty` files from the old settings-backup PVC into the S3Proxy PVC bucket before the old PVC is removed.
* `global.backup.enabled=true`:
  * Backups are already in the external S3 bucket. No data migration needed. The old settings-backup PVC can be removed directly.
* MinIO PVC (local persistence mode): reused asthe S3Proxy PVC.

**Pros:**
* Clean, uniform model — everything goes through S3Proxy.
* Settings-backup PVC goes away entirely.
* Remote-backup users need zero data migration for settings.

**Cons:**
* Local-only users require a copy job on upgrade.
* Switching from `global.backup.enabled=false` to `true` later means the local PVC stops being used for settings; existing backups on it become orphaned unless a copy-up job runs. *(Mitigated by the sync init container described below.)*

### Option A.1: S3Proxy-only but keeping 2 PVCs (dual-bucket)

Settings backups are stored in two places as they are today, but both are accessed via S3Proxy. S3Proxy always has a small dedicated settings PVC backing a `settings-local` bucket; backups are always written there. When `global.backup.enabled=true`, S3Proxy also proxies the normal `sts-configuration-backup` bucket to the external S3 provider (or, in local persistence mode, to the main S3Proxy PVC that replaces the old MinIO PVC).

**New behavior:**
* Settings backups always go through S3Proxy to the `settings-local` bucket (backed by a dedicated small PVC). This replaces the old settings-backup PVC.
* When `global.backup.enabled=true`: backups are additionally written to the `sts-configuration-backup` bucket, which S3Proxy proxies to the external backend (or the main S3Proxy PVC in local persistence mode).
* When `global.backup.enabled=false`: only the `settings-local` bucket is used.
* Two copies of settings backups exist when global backup is enabled (local + remote), matching today's behavior.

**Migration:**
* `global.backup.enabled=false`:
  * A one-off init container copies `.sty` files from the old settings-backup PVC into the `settings-local` bucket on the new S3Proxy settings PVC.
* `global.backup.enabled=true`:
  * Backups are already in the external S3 bucket. The init container copies the old PVC files into `settings-local` for completeness, but this could even be skipped.
  * The old settings-backup PVC can then be removed.
* MinIO PVC (local persistence mode): reused as the main S3Proxy PVC.

**Pros:**
* Preserves today's dual-copy safety model — a local copy always exists regardless of remote backend availability.
* Switching `global.backup.enabled` on or off never orphans backups; the `settings-local` bucket always has a copy.
* No sync init container needed for backend switches (the local copy is always present).

**Cons:**
* Two PVCs instead of one (the settings PVC is small, ≤1Gi).
* Slightly more complex S3Proxy configuration (two buckets with different backends).
* Backup jobs need to write to two buckets when global backup is enabled (or S3Proxy replicates internally).

---

### Option B: S3Proxy with the settings-backup PVC reused as its backend

Instead of provisioning a new PVC for S3Proxy, reuse the existing settings-backup PVC as the S3Proxy filesystem backend. S3Proxy serves buckets from that PVC.

**New behavior:**
* Same as Option A from the application's perspective (all traffic via S3Proxy).
* The PVC is "owned" by S3Proxy now, not mounted by backup jobs directly.
* When `global.backup.enabled=true`, S3Proxy proxies to the external backend; the local PVC may still exist but isn't used for settings.

**Migration:**
* `global.backup.enabled=false`:
  * The existing settings-backup PVC is adopted by the S3Proxy StatefulSet/Deployment.
  * S3Proxy needs to be configured to use the existing directory layout or a one-time restructure (move files into a bucket-prefixed directory) is performed by an init container.
* `global.backup.enabled=true`:
  * Same as Option A — no data copy needed for settings.

**Pros:**
* No new PVC; no copy job for local-only users (if directory layout is compatible).
* Simplest upgrade path.

**Cons:**
* The internal layout of the old PVC (flat `.sty` files) may not match S3Proxy's expected bucket directory structure, requiring a restructure step anyway.
* Reusing an existing PVC across different workload types can hit zone/topology constraints if the PVC was bound to a node in a different zone.
* Switching from `global.backup.enabled=false` to `true` later means the local PVC stops being used for settings; existing backups on it become orphaned unless a copy-up job runs. *(Mitigated by the sync init container described below.)*

---

### Mitigating backup loss when switching PVC backends

Options A and B both have the risk that switching from `global.backup.enabled=false` to `true` (or vice versa) silently orphans the backups stored on the local PVC. This can be handled by an **init container on the S3Proxy pod** that runs before S3Proxy starts:

1. **Detect backend change.** The init container compares the currently configured backend (PVC vs remote S3) against a small marker file on the local PVC (e.g. `.backend-type`).
2. **If the backend changed from PVC → remote S3** (user enabled global backup):
   * List all `.sty` files on the local PVC bucket directory.
   * Upload each file to the remote S3 bucket via the configured backend credentials.
   * Update the marker file.
3. **If the backend changed from remote S3 → PVC** (user disabled global backup):
   * List the remote bucket and download the most recent N files to the local PVC.
   * Update the marker file.
4. **If the backend did not change**, do nothing.

This makes the transition seamless in both directions. Because settings backups are small (≤10 files, each a few MB at most), the sync completes in seconds.

The same init container can also handle the **initial migration from the old settings-backup PVC**: if S3Proxy's own PVC is empty and the old PVC still exists (mounted read-only as a second volume), it copies the files over before the first start. After one successful upgrade cycle the old PVC volume mount can be removed from the chart.

---

### Option C: Dual-write transition period

Keep the settings-backup PVC for one release cycle while also writing to S3Proxy. After the transition release, drop the PVC.

**Release N (transition):**
* Backup CronJob writes settings to both the old PVC (directly) and the S3Proxy bucket.
* Restore prefers S3Proxy, falls back to PVC.
* All existing PVC backups remain accessible.

**Release N+1 (cleanup):**
* Remove PVC mounts from backup jobs.
* Old PVC is no longer used and can be deleted.

**Pros:**
* Zero-risk migration; no copy job; no data loss at any point.
* Users can roll back to Release N-1 without losing data.

**Cons:**
* Two releases to fully complete the migration.
* Increased complexity during the transition release (dual-write logic).
* Temporary storage overhead (two copies of each new backup during the transition).

---

### Recommendation

**Option A.1** has the least changes in runtime behavior and gives a consistent endstate without having to deal with orphaned backups when enabling/disabling global backup. The copy job for local-only users is a one-time cost that can be automated in an init container, and the dual-bucket setup preserves the safety of having a local copy regardless of remote backend status:

* For the production users that really care about backups (`global.backup.enabled=true`),  backups are already in S3.
* For local-only users, a simple init container copies ≤10 small `.sty` files from the old PVC to the S3Proxy PVC and then the old PVC can be cleaned up.
* The MinIO PVC (when it exists) can be reused or replaced by the S3Proxy PVC.

**Option B** is the simplest from a chart perspective but has more risks around PVC reuse and orphaned backups when switching backends.

## S3 compatibility of cloud providers

Answered by Gemini but aligns exactly with my own experience.

---

Among the major cloud providers, **Microsoft Azure** is the only one that does not use the S3 API as its primary or native interface.

While almost every other major provider (Google, Oracle, IBM, Alibaba, DigitalOcean) has adopted the S3 API as a standard—often to the point where they have no other API—Azure has maintained its own distinct **Azure Blob Storage REST API**.

Here is the breakdown of how the major providers handle this:

### 1. Microsoft Azure (The Outlier)

* **Native API:** Azure Blob Storage REST API.
* **S3 Compatibility:** **Not native by default.**
* Historically, if you wanted to use S3 tools with Azure, you had to use third-party gateways (like MinIO or Flexify) that sat between your application and Azure to translate the requests.
* Microsoft is currently working on native S3 compatibility (titled "S3 API on Azure Blob Storage"), but it is largely a specific feature/preview capability rather than the default way the storage works.


* **Why it matters:** If you have an application written strictly for AWS S3, you cannot simply point it to Azure Blob Storage by changing the endpoint URL and credentials (as you can with other providers). You usually need to rewrite the storage logic to use the Azure SDK or deploy a translation gateway.

### 2. Google Cloud Platform (GCP)

* **Native API:** Google Cloud Storage JSON API.
* **S3 Compatibility:** **Native Interoperability.**
* GCP offers a feature called "Interoperability" or the **XML API**.
* You can generate separate "Interoperability Keys" (HMAC keys) that look exactly like AWS Access/Secret keys.
* Once generated, you can point standard S3 tools (like the AWS CLI or SDKs) at `storage.googleapis.com`, and they will work seamlessly.



### 3. Other Major Providers

The rest of the market has largely standardized on S3 to ensure customers can migrate easily from AWS.

| Provider | S3 API Status | Notes |
| --- | --- | --- |
| **AWS** | **Native** | The creator of the standard. |
| **Oracle Cloud (OCI)** | **Native** | S3 Compatibility API is a core feature; often used as the primary method. |
| **IBM Cloud** | **Native** | Object Storage is built on COS (Cloud Object Storage) which is S3 compatible. |
| **Alibaba Cloud** | **Native** | Their OSS (Object Storage Service) is fully S3 compatible. |
| **DigitalOcean / Linode / Vultr** | **Native** | Their storage products (Spaces, Object Storage) are built specifically to be S3 drop-in replacements (often using Ceph). |
| **Cloudflare R2** | **Native** | Built entirely around the S3 API; they do not have a proprietary object API. |
| **Backblaze B2** | **Native-ish** | Originally had a proprietary API, but introduced a fully native S3-compatible API layer several years ago that is now the standard for most users. |

### Summary

If you are looking for a provider where your S3-based code **will not work** out of the box without modification or extra infrastructure, **Microsoft Azure** is the answer.
