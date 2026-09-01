# Updatecli Docker Image → Values Mapping

Documents where each Docker image from quay.io/stackstate is used in the helm-charts repository.

## Registry

All images: `quay.io/stackstate/<image-name>` (no authentication required for pull).

## Image → Values File Mapping

| Image | Values File | JSONPath Key(s) | Notes |
|-------|-------------|-----------------|-------|
| clickhouse-backup | stable/suse-observability/values.yaml | $.clickhouse.backup.image.tag | Tag format: `<semver>-so<release_increment>` (e.g. `2.6.43-so4`) |
| clickhouse | stable/suse-observability/values.yaml | $.clickhouse.image.tag, $.stackstate.components.clickhouseCleanup.image.tag | Tag format `<semver>-so<release_increment>` |
| container-tools | stable/suse-observability/values.yaml | $.stackstate.components.router.mode.image.tag, $.stackstate.components.containerTools.image.tag | Tag format: `<semver>-so<release_increment>` |
| container-tools | stable/suse-observability/values.yaml | $.victoria-metrics-0.backup.setupCron.image.tag, $.victoria-metrics-1.backup.setupCron.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1) |
| container-tools | stable/suse-observability-agent/values.yaml | $.httpHeaderInjectorWebhook.certificatePrehook.image.tag | When values change, Chart.yaml version is bumped via shell target |
| elasticsearch | stable/suse-observability/values.yaml | $.elasticsearch.imageTag | Tag format `<semver>-so<release_increment>` |
| elasticsearch-exporter | stable/suse-observability/values.yaml | $.elasticsearch.prometheus-elasticsearch-exporter.image.tag | Tag format: `<semver>-so<release_increment>` |
| envoy | stable/suse-observability/values.yaml | $.stackstate.components.router.image.tag | |
| hadoop | stable/suse-observability/values.yaml | $.hbase.hdfs.version | Tag format `<semver>-so<release_increment>` |
| jmx-exporter | stable/suse-observability/values.yaml | $.kafka.metrics.jmx.image.tag | Tag format: `<semver>-so<release_increment>` |
| kafka | stable/suse-observability/values.yaml | $.kafka.image.tag, $.stackstate.components.kafkaTopicCreate.image.tag | Tag format `<semver>-so<release_increment>` |
| kubernetes-rbac-agent | stable/suse-observability/values.yaml | $.kubernetes-rbac-agent.containers.rbacAgent.image.tag | **Disabled:** tag key must be merged to master first (updatecli clones from remote) |
| kubernetes-rbac-agent | stable/suse-observability-agent/values.yaml | $.kubernetes-rbac-agent.containers.rbacAgent.image.tag | When values change, Chart.yaml version is bumped via shell target |
| nginx-prometheus-exporter | stable/suse-observability/values.yaml | $.stackstate.components.nginxPrometheusExporter.image.tag | Tag format: `<semver>-so<release_increment>` |
| suse-observability-mcp | stable/suse-observability/values.yaml | $.stackstate.components.mcp.image.tag | Tag format: `YYYYMMDDHHMMSS-hash`; sorted by timestamp |
| suse-observability-borg | stable/suse-observability/values.yaml | $.stackstate.components.aiAssistant.image.tag | Charted as `aiAssistant`; tag format: `YYYYMMDDHHMMSS-hash`; sorted by timestamp |
| victoria-metrics | stable/suse-observability/values.yaml | $.victoria-metrics-0.server.image.tag, $.victoria-metrics-1.server.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1); tag format `<semver>-so<release_increment>` |
| vmagent | stable/suse-observability/values.yaml | $.stackstate.components.vmagent.image.tag | Tag format: `<semver>-so<release_increment>` |
| vmbackup | stable/suse-observability/values.yaml | $.victoria-metrics-0.backup.vmbackup.image.tag, $.victoria-metrics-1.backup.vmbackup.image.tag | Victoria-metrics-single subchart (aliased as victoria-metrics-0/1); tag format `<semver>-so<release_increment>` |
| vmrestore | stable/suse-observability/values.yaml | $.victoria-metrics.restore.image.tag | Tag format: `<semver>-so<release_increment>` |
| wait | stable/suse-observability/values.yaml | $.global.wait.image.tag | Tag format: `<semver>-so<release_increment>` |
| workload-observer | stable/suse-observability/values.yaml | $.stackstate.components.workloadObserver.image.tag | Tag format: `hash-buildId` only |
| zookeeper | stable/suse-observability/values.yaml | $.zookeeper.image.tag | Tag format `<semver>-so<release_increment>` |
| spotlight | stable/suse-observability/values.yaml | $.anomaly-detection.image.tag | Tag format: `X.Y.Z-snapshot.N`, optionally followed by `.rN` or `.rebuild.N` for immutable rebuilds |
| sts-opentelemetry-collector | stable/suse-observability/values.yaml | $.opentelemetry-collector.image.tag | Server (full BOM) image. Tag format: `vX.Y.Z` (plain semver release tags only). Source `stsOpentelemetryCollector` |
| sts-opentelemetry-collector (agent BOM) | stable/suse-observability-agent/values.yaml | $.otel.k8sResourceCollector.image.tag | Strict agent BOM variant of the same image. Tag format: `vX.Y.Z-agent`. Source `stsOpentelemetryAgentCollector` (regex/semver ranks on the semver core, writes the full `-agent` tag). When values change, Chart.yaml version is bumped via shell target |
| sts-opentelemetry-collector (agent BOM) | stable/suse-observability-agent/values.yaml | $.otel.prometheusScraping.collector.image.tag | OTel Prometheus scraping collector — same agent BOM image/tag as k8sResourceCollector, separate tag entry for independent override |
| promtail | stable/suse-observability-agent/values.yaml | $.logsAgent.image.tag | Agent chart — not a subchart of suse-observability; tag format `<semver>-so<release_increment>` |
| s3proxy | stable/suse-observability/values.yaml | $.s3proxy.image.tag | Tag format: `<semver>-so<release_increment>` |
| http-header-injector-proxy | stable/suse-observability-agent/values.yaml | $.httpHeaderInjectorWebhook.proxy.image.tag | Tag format `<semver>-so<release_increment>`; local chart lives in `local/http-header-injector` |
| http-header-injector-proxy-init | stable/suse-observability-agent/values.yaml | $.httpHeaderInjectorWebhook.proxyInit.image.tag | Tag format `<semver>-so<release_increment>`; local chart lives in `local/http-header-injector` |

## Tag Format Reference

- **Standard:** `prefix-hash-release-buildId` (e.g. `v1.109.0-614527d8-release-138`), tagfilter `.*-[a-f0-9]{8}-release-[0-9]+$`
- **SUSE Observability semver:** `<semver>-so<release_increment>` (e.g. `1.143.0-so2`)
- **spotlight:** `X.Y.Z-snapshot.N`, with optional immutable rebuild suffix `.rN` or `.rebuild.N`
- **workload-observer, kubernetes-rbac-agent:** `hash-buildId-release` (e.g. `f40221cf-76-release`), tagfilter `^[a-f0-9]{8}-[0-9]+-release$`
- **suse-observability-mcp, suse-observability-borg:** `YYYYMMDDHHMMSS-hash` (e.g. `20260430073230-dc6221d7`), tagfilter `^[0-9]{14}-[0-9a-f]{8}$`; updatecli uses `regex/time` to select the newest timestamped tag
- **sts-opentelemetry-collector (server):** `vX.Y.Z` (e.g. `v0.0.46`), tagfilter `^v[0-9]+\.[0-9]+\.[0-9]+$`
- **sts-opentelemetry-collector (agent BOM):** `vX.Y.Z-agent` (e.g. `v0.0.46-agent`), tagfilter `^v[0-9]+\.[0-9]+\.[0-9]+-agent$`; updatecli uses `regex/semver` (regex `^v([0-9]+\.[0-9]+\.[0-9]+)-agent$`) to rank on the semver core while writing the full `-agent` tag

## Run Locally

```bash
cd helm-charts
source .env
# HELM_CHARTS_PAT — env var name preserved from the GitLab era; the value is now a GitHub App token.
docker run --rm -v $(pwd):/workspace -w /workspace -e HELM_CHARTS_PAT \
  quay.io/stackstate/container-tools:1.8.6_dev-fa52bb17-main-4 \
  updatecli pipeline diff --config updatecli/updatecli.d/update-docker-images/ --values updatecli/values.d/values.yaml
```
