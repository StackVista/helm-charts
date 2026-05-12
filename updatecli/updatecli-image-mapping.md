# Updatecli Docker Image → Values Mapping

Documents where each Docker image from quay.io/stackstate is used in the helm-charts repository.

## Registry

All images: `quay.io/stackstate/<image-name>` (no authentication required for pull).

## Image → Values File Mapping

| Image | Values File | JSONPath Key(s) | Notes |
|-------|-------------|-----------------|-------|
| clickhouse-backup | stable/suse-observability/values.yaml | $.clickhouse.backup.image.tag | Tag format: `2.6.43-4ac12b1a-main-13`; standard `main` tags only |
| clickhouse | stable/suse-observability/values.yaml | $.clickhouse.image.tag, $.stackstate.components.clickhouseCleanup.image.tag | Tag format: `25.9.6-7703df37-main-12`; standard `main` tags only |
| container-tools | stable/suse-observability/values.yaml | $.stackstate.components.router.mode.image.tag, $.stackstate.components.containerTools.image.tag | Tag format: `1.8.6-fa52bb17-main-4`; standard `main` tags only |
| container-tools | stable/suse-observability/values.yaml | $.victoria-metrics-0.backup.setupCron.image.tag, $.victoria-metrics-1.backup.setupCron.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1) |
| container-tools | stable/suse-observability-agent/values.yaml | $.httpHeaderInjectorWebhook.certificatePrehook.image.tag | When values change, Chart.yaml version is bumped via shell target |
| elasticsearch | stable/suse-observability/values.yaml | $.elasticsearch.imageTag | Tag format: `8.19.14-fdb72fef-main-9`; standard `main` tags only |
| elasticsearch-exporter | stable/suse-observability/values.yaml | $.elasticsearch.prometheus-elasticsearch-exporter.image.tag | |
| envoy | stable/suse-observability/values.yaml | $.stackstate.components.router.image.tag | |
| hadoop | stable/suse-observability/values.yaml | $.hbase.hdfs.version | **Note:** Values expect `java21-8-hash-build`; findsubmatch transformer strips semver prefix from docker tag |
| jmx-exporter | stable/suse-observability/values.yaml | $.kafka.metrics.jmx.image.tag | Tag format: `0.20.0-58a72255-main-319`; standard `main` tags only |
| kafka | stable/suse-observability/values.yaml | $.kafka.image.tag, $.stackstate.components.kafkaTopicCreate.image.tag | Digest-preserved AppCo retag; standard `main` tags only |
| kubernetes-rbac-agent | stable/suse-observability/values.yaml | $.kubernetes-rbac-agent.containers.rbacAgent.image.tag | **Disabled:** tag key must be merged to master first (updatecli clones from remote) |
| kubernetes-rbac-agent | stable/suse-observability-agent/values.yaml | $.kubernetes-rbac-agent.containers.rbacAgent.image.tag | When values change, Chart.yaml version is bumped via shell target |
| nginx-prometheus-exporter | stable/suse-observability/values.yaml | $.stackstate.components.nginxPrometheusExporter.image.tag | Tag format: `1.5.1-61e00a86-main-2`; standard `main` tags only |
| suse-observability-mcp | stable/suse-observability/values.yaml | $.stackstate.components.mcp.image.tag | Tag format: `YYYYMMDDHHMMSS-hash`; sorted by timestamp |
| suse-observability-borg | stable/suse-observability/values.yaml | $.stackstate.components.aiAssistant.image.tag | Charted as `aiAssistant`; tag format: `YYYYMMDDHHMMSS-hash`; sorted by timestamp |
| victoria-metrics | stable/suse-observability/values.yaml | $.victoria-metrics-0.server.image.tag, $.victoria-metrics-1.server.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1); GitHub `main` tags only |
| vmagent | stable/suse-observability/values.yaml | $.stackstate.components.vmagent.image.tag | GitHub `main` tags only |
| vmbackup | stable/suse-observability/values.yaml | $.victoria-metrics-0.backup.vmbackup.image.tag, $.victoria-metrics-1.backup.vmbackup.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1); GitHub `main` tags only |
| vmrestore | stable/suse-observability/values.yaml | $.victoria-metrics.restore.image.tag | GitHub `main` tags only |
| wait | stable/suse-observability/values.yaml | $.global.wait.image.tag | Tag format: `1.0.12-61e00a86-main-2`; standard `main` tags only |
| workload-observer | stable/suse-observability/values.yaml | $.stackstate.components.workloadObserver.image.tag | Tag format: `hash-buildId` only |
| zookeeper | stable/suse-observability/values.yaml | $.zookeeper.image.tag | Tag format: `3.9.5-af2b6748-main-1`; standard `main` tags only |
| spotlight | stable/suse-observability/values.yaml | $.anomaly-detection.image.tag | Tag format: `X.Y.Z-snapshot.N` (master snapshots only, semver with pre-release) |
| sts-opentelemetry-collector | stable/suse-observability/values.yaml | $.opentelemetry-collector.image.tag | Tag format: `vX.Y.Z` (semver release tags only) |
| sts-opentelemetry-collector | stable/suse-observability-agent/values.yaml | $.k8sCrdCollector.image.tag | When values change, Chart.yaml version is bumped via shell target |
| promtail | stable/suse-observability-agent/values.yaml | $.logsAgent.image.tag | Agent chart — not a subchart of suse-observability; tag format: `3.6.10-bc99e3ee-main-8`; standard `main` tags only |
| s3proxy | stable/suse-observability/values.yaml | $.s3proxy.image.tag | Tag format: `3.1.0-bc99e3ee-main-4`; standard `main` tags only |
| http-header-injector (chart) | stable/suse-observability-agent/Chart.yaml | dependencies[].version | Helm chart version from helm.stackstate.io — shell target matches by name |

## Tag Format Reference

- **Standard:** `prefix-hash-release-buildId` (e.g. `v1.109.0-614527d8-release-138`), tagfilter `.*-[a-f0-9]{8}-release-[0-9]+$`
- **VictoriaMetrics GitHub main:** `v<semver>-hash-main-run`
- **workload-observer, kubernetes-rbac-agent:** `hash-buildId-release` (e.g. `f40221cf-76-release`), tagfilter `^[a-f0-9]{8}-[0-9]+-release$`
- **GitHub main:** `version-hash-main-run` (e.g. `2.6.43-4ac12b1a-main-13`), tagfilter `^[0-9]+\.[0-9]+\.[0-9]+-[a-f0-9]{8}-main-[0-9]+$`
- **container-tools:** same `main` tag format for customer-runtime tags; dev tags such as `1.8.6_dev-*` are intentionally ignored
- **kafka:** digest-preserved SUSE Application Collection retag; downstream updatecli still tracks the standard GitHub `main` tag format because the Quay tag is an alias to the preserved upstream digest
- **zookeeper:** digest-preserved SUSE Application Collection retag; uses the ZooKeeper application version as the generated tag prefix while the docker-images source metadata keeps the full AppCo source tag
- **suse-observability-mcp, suse-observability-borg:** `YYYYMMDDHHMMSS-hash` (e.g. `20260430073230-dc6221d7`), tagfilter `^[0-9]{14}-[0-9a-f]{8}$`; updatecli uses `regex/time` to select the newest timestamped tag

## Run Locally

```bash
cd helm-charts
source .env
docker run --rm -v $(pwd):/workspace -w /workspace -e GITLAB_TOKEN \
  quay.io/stackstate/container-tools:1.8.6_dev-fa52bb17-main-4 \
  updatecli diff --config updatecli/updatecli.d/update-docker-images/ --values updatecli/values.d/values.yaml
```
