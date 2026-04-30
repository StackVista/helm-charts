# Updatecli Docker Image → Values Mapping

Documents where each Docker image from quay.io/stackstate is used in the helm-charts repository.

## Registry

All images: `quay.io/stackstate/<image-name>` (no authentication required for pull).

## Image → Values File Mapping

| Image | Values File | JSONPath Key(s) | Notes |
|-------|-------------|-----------------|-------|
| clickhouse-backup | stable/suse-observability/values.yaml | $.clickhouse.backup.image.tag | Tag format: `2.6.43-4ac12b1a-main-13`; standard `main` tags only |
| clickhouse | stable/suse-observability/values.yaml | $.clickhouse.image.tag, $.stackstate.components.clickhouseCleanup.image.tag | |
| container-tools | stable/suse-observability/values.yaml | $.stackstate.components.router.mode.image.tag, $.stackstate.components.containerTools.image.tag | Tag format: `1.8.6-fa52bb17-main-4`; standard `main` tags only |
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
| nginx-prometheus-exporter | stable/suse-observability/values.yaml | $.stackstate.components.nginxPrometheusExporter.image.tag | |
| victoria-metrics | stable/suse-observability/values.yaml | $.victoria-metrics-0.server.image.tag, $.victoria-metrics-1.server.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1) |
| vmagent | stable/suse-observability/values.yaml | $.stackstate.components.vmagent.image.tag | |
| vmbackup | stable/suse-observability/values.yaml | $.victoria-metrics-0.backup.vmbackup.image.tag, $.victoria-metrics-1.backup.vmbackup.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1) |
| vmrestore | stable/suse-observability/values.yaml | $.victoria-metrics.restore.image.tag | |
| wait | stable/suse-observability/values.yaml | $.global.wait.image.tag | |
| workload-observer | stable/suse-observability/values.yaml | $.stackstate.components.workloadObserver.image.tag | Tag format: `hash-buildId` only |
| zookeeper | stable/suse-observability/values.yaml | $.zookeeper.image.tag | |
| spotlight | stable/suse-observability/values.yaml | $.anomaly-detection.image.tag | Tag format: `X.Y.Z-snapshot.N` (master snapshots only, semver with pre-release) |
| promtail | stable/suse-observability-agent/values.yaml | $.logsAgent.image.tag | Agent chart — not a subchart of suse-observability |
| s3proxy | stable/suse-observability/values.yaml | $.s3proxy.image.tag | Tag format: `3.1.0-hash-release-buildId` |
| http-header-injector (chart) | stable/suse-observability-agent/Chart.yaml | dependencies[].version | Helm chart version from helm.stackstate.io — shell target matches by name |

## Tag Format Reference

- **Standard:** `prefix-hash-release-buildId` (e.g. `v1.109.0-614527d8-release-138`), tagfilter `.*-[a-f0-9]{8}-release-[0-9]+$`
- **workload-observer, kubernetes-rbac-agent:** `hash-buildId-release` (e.g. `f40221cf-76-release`), tagfilter `^[a-f0-9]{8}-[0-9]+-release$`
- **GitHub main:** `version-hash-main-run` (e.g. `2.6.43-4ac12b1a-main-13`), tagfilter `^[0-9]+\.[0-9]+\.[0-9]+-[a-f0-9]{8}-main-[0-9]+$`
- **container-tools:** same `main` tag format for customer-runtime tags; dev tags such as `1.8.6_dev-*` are intentionally ignored

## Run Locally

```bash
cd helm-charts
source .env
docker run --rm -v $(pwd):/workspace -w /workspace -e GITLAB_TOKEN \
  quay.io/stackstate/container-tools:1.8.6_dev-fa52bb17-main-4 \
  updatecli diff --config updatecli/updatecli.d/update-docker-images/ --values updatecli/values.d/values.yaml
```
