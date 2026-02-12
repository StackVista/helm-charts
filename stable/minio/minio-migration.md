# Migration from Minio AND settings-backup-pvc -> S3Proxy

The new setup should have:
* S3Proxy always deployed (https://github.com/gaul/s3proxy)? Or Versity Gateway (https://github.com/versity/versitygw)?  Using s3proxy as placeholder in the rest of the doc.
* Settings backup by default using S3Proxy instead of a local PVC, with a bucket that uses a small PVC as storage backend (e.g. 1Gi)
* When global backup is enabled S3Proxy should proxy the other buckets to the configured S3 provider, so that the backup can be stored there

Questions:
* Next to running a local object storage solution (which could be anything from rustfs, garagefs or S3Proxy), do we want to support anything except S3 compatible APIs? Mainly Azure Blob storage, for Google you need to do some extra steps to use (get specific credentials) but it works fine especially for simply storing backups.
* Do we want to have a proxy for S3 compatible apis or simply directly configure the S3 credentials in the different backup settings?

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
* The backup subsections remain the same
* We drop the `minio` section in favor of introducing a `backup.storage` section that contains a solution-agnostic configuration:

```
backup:
  storage:
    credentials:
      accessKey: <make-up-an-access-key>
      secretKey: <make-up-a-secret-key>
    localPvc:
      enabled: true
      size: 500Gi
    s3

    type: s3proxy
    s3proxy:
      accessKey: AWS_ACCESS_KEY
      secretKey: AWS_SECRET_KEY
      endpoint: http://s3proxy:9000
      region: us-east-1
      bucketName: BACKUP_BUCKET_NAME
    azure:
      accountName: AZURE_STORAGE_ACCOUNT_NAME
      accountKey: AZURE_STORAGE_ACCOUNT_KEY
      containerName: BACKUP_CONTAINER_NAME
```

Questions:
* Do we need to specify credentials for connection to S3Proxy or do we leave it unauthenticated (the former seems to be the safer choice tbh)?


## Migration of data
* Local PVC for all backups (global.backup.enabled=true):
    * If this PVC was already in use by Minio before it should be reused by S3Proxy, so that the data is not lost and the migration is seamless
* Settings backup PVC:
    * We cannot reuse the same PVC, because in some cases the existing PVCs (settings backup and Minio) can be in a different region/zone which will make it impossible to mount both
    * When upgrading we should preserve the backups from this PVC to the PVC for S3Proxy, so that the migration is seamless and no data is lost. We can do this by copying the data from the existing PVC to the new PVC for S3Proxy before deleting the old PVC via a k8s job

Alternative migration setup:
* Local PVC for all backups (global.backup.enabled=true):
    * If this PVC was already in use by Minio before it should be reused by S3Proxy, so that the data is not lost and the migration is seamless
* Settings backup PVC:
    * We reuse the same PVC for S3Proxy, but we change the settings backup logic to only use a local PVC when the global backup is not enabled.
    * This works because in the existing situation settings backups are stored both on the local PVC **and** in the object store.
    * Benefits:
        * No need to copy data from the old PVC to the new PVC, so the migration is faster and simpler
        * We can drop the special handling for settings backups in the restore and backup logic, backups always go to s3proxy and the bucket location depends on the global backup setting
    * Downsides:
        * First installing observability with just the local settings backup PVC and then enabling global backup will cause the existing settings backups to be lost. This is a bit of an edge case, but it can be mitigated by adding a warning in the documentation about this. Note that the local settings backup is limited to only 10 backups anyway and a backup is made daily. A similar job (or init container on S3Proxy) can be added to take care of copying the existing settings backups to the S3Proxy bucket when global backup is enabled, so that the data is not lost and the migration is seamless in this case as well. It can also do the cleanup of the old PVC after copying the data. This can even be done in both directions, but it also feels like an optimization we could do later.





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
