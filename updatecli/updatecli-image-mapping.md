# Updatecli Docker Image → Values Mapping

Documents where each Docker image from quay.io/stackstate is used in the helm-charts repository.

## Registry

All images: `quay.io/stackstate/<image-name>` (no authentication required for pull).

## Image → Values File Mapping

| Image | Values File | JSONPath Key(s) | Notes |
|-------|-------------|-----------------|-------|
| clickhouse-backup | stable/suse-observability/values.yaml | $.clickhouse.backup.image.tag | |
| clickhouse | stable/suse-observability/values.yaml | $.clickhouse.image.tag, $.stackstate.components.clickhouseCleanup.image.tag | |
| container-tools | stable/suse-observability/values.yaml | $.stackstate.components.router.mode.image.tag, $.stackstate.components.containerTools.image.tag | Tag format: `1.8.3-573` |
| container-tools | stable/suse-observability/values.yaml | $.victoria-metrics-0.backup.setupCron.image.tag, $.victoria-metrics-1.backup.setupCron.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1) |
| container-tools | stable/suse-observability-agent/values.yaml | $.httpHeaderInjectorWebhook.certificatePrehook.image.tag | When values change, Chart.yaml version is bumped via shell target |
| elasticsearch | stable/suse-observability/values.yaml | $.elasticsearch.imageTag | |
| elasticsearch-exporter | stable/suse-observability/values.yaml | $.elasticsearch.prometheus-elasticsearch-exporter.image.tag | |
| envoy | stable/suse-observability/values.yaml | $.stackstate.components.router.image.tag | |
| hadoop | stable/suse-observability/values.yaml | $.hbase.hdfs.version | **Note:** Values expect `java21-8-hash-build`; findsubmatch transformer strips semver prefix from docker tag |
| jmx-exporter | stable/suse-observability/values.yaml | $.kafka.metrics.jmx.image.tag | |
| kafka | stable/suse-observability/values.yaml | $.kafka.image.tag, $.stackstate.components.kafkaTopicCreate.image.tag | |
| kubernetes-rbac-agent | stable/suse-observability/values.yaml | $.kubernetes-rbac-agent.containers.rbacAgent.image.tag | **Disabled:** tag key must be merged to master first (updatecli clones from remote) |
| kubernetes-rbac-agent | stable/suse-observability-agent/values.yaml | $.kubernetes-rbac-agent.containers.rbacAgent.image.tag | When values change, Chart.yaml version is bumped via shell target |
| minio | stable/suse-observability/values.yaml | $.minio.image.tag | |
| nginx-prometheus-exporter | stable/suse-observability/values.yaml | $.stackstate.components.nginxPrometheusExporter.image.tag | |
| victoria-metrics | stable/suse-observability/values.yaml | $.victoria-metrics-0.server.image.tag, $.victoria-metrics-1.server.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1) |
| vmagent | stable/suse-observability/values.yaml | $.stackstate.components.vmagent.image.tag | |
| vmbackup | stable/suse-observability/values.yaml | $.victoria-metrics-0.backup.vmbackup.image.tag, $.victoria-metrics-1.backup.vmbackup.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1) |
| vmrestore | stable/suse-observability/values.yaml | $.victoria-metrics.restore.image.tag | |
| wait | stable/suse-observability/values.yaml | $.global.wait.image.tag | |
| workload-observer | stable/suse-observability/values.yaml | $.stackstate.components.workloadObserver.image.tag | Tag format: `hash-buildId` only |
| zookeeper | stable/suse-observability/values.yaml | $.zookeeper.image.tag | |

## Tag Format Reference

- **Standard:** `prefix-hash-buildId` (e.g. `v1.109.0-614527d8-138`), tagfilter `.*-[a-f0-9]{8}-[0-9]+$`
- **workload-observer, kubernetes-rbac-agent:** `hash-buildId` only (e.g. `f40221cf-76`), tagfilter `^[a-f0-9]{8}-[0-9]+$`
- **container-tools:** `version-buildId` (e.g. `1.8.3-573`), tagfilter `^[0-9]+\.[0-9]+\.[0-9]+-[0-9]+$`

## Run Locally

```bash
cd helm-charts
source .env
docker run --rm -v $(pwd):/workspace -w /workspace -e GITLAB_TOKEN \
  quay.io/stackstate/container-tools:23b0c0da-591-dev \
  updatecli diff --config updatecli/updatecli.d/update-docker-images/ --values updatecli/values.d/values.yaml
```
