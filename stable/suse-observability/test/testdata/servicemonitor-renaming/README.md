# ServiceMonitor naming compatibility

These frozen fixtures have been independently reproduced from published
pre-rename revision
[`1e092cd119359c8184bc5997211f85bca4ecc9ec`](https://github.com/StackVista/helm-charts-internal/commit/1e092cd119359c8184bc5997211f85bca4ecc9ec)
on `master`, using Helm 3.19.0. This revision precedes both the PDB and
ServiceMonitor renames and remains reachable independently of PR rebases.

Verification exported that revision into a temporary directory and rebuilt the
platform chart's file dependencies from the same revision, including nested
dependencies. All three complete ServiceMonitor inventories matched the existing
JSON fixtures in install and upgrade rendering, excluding only the two version
labels described below. The frozen files and their SHA-256 hashes are unchanged.

| Fixture | Server and worker mode | Main-chart monitors |
| --- | --- | --- |
| `split.json` | Split server, unsplit receiver/correlate | 15 |
| `mono.json` | Monolithic server, unsplit receiver/correlate | 7 |
| `workers.json` | Split server and receiver/correlate workers | 19 |

The reference release is `nightly`, with application namespace `observability`
and monitoring namespace `monitoring`. Inputs use `test/values/full.yaml`,
with metrics and monitoring enabled, Kubernetes authorization and backups
enabled, explicit worker memory sizing, and custom monitor annotations and labels.
`serviceMonitorCompatibilityValues` records the matching test settings.

Each fixture records the complete main-chart ServiceMonitors. Only
`helm.sh/chart` and `app.kubernetes.io/version` labels are excluded so automatic
chart and application version updates do not invalidate the contract.
Tests change only the expected metadata names, release labels and selected
application namespace when the scenario requests them. Everything else,
including endpoint ports, paths, intervals, schemes, relabeling, selectors,
monitoring namespace, annotations and discovery labels, must match.

The exact rename allowlist has these 20 component suffixes:

```text
api, authorization-sync, checks, correlate, correlate-connection,
correlate-http-tracing, correlate-aggregator, e2es, health-sync,
initializer, notification, receiver, receiver-base, receiver-logs,
receiver-process-agent, router, server, slicing, state, sync
```

For example, `ServiceMonitor/nightly-suse-observability-api` becomes
`ServiceMonitor/suse-observability-api`. Names customized through root/global
fullname overrides, prefixes or suffixes also become canonical. UI and S3Proxy
monitor names are already canonical and remain unchanged. All subchart monitor
names, Services, application workloads, configuration and storage stay unchanged.

The test exercises install and upgrade rendering across default/custom releases,
two application namespaces, root/global fullname settings, long names,
split/monolithic servers, split workers, HA/non-HA profiles and Argo CD mode.
It verifies that Prometheus label selectors continue selecting the same monitors.
Existing helper tests cover feature enablement and disablement.

Keep the frozen fixtures unchanged. Do not generate new expectations from the
renamed chart to make a test pass.

To reproduce the baseline, run these commands from the current repository root
with Helm 3.19.0 available as `helm`. This exports sources, builds local packages
and renders templates; it does not install anything in Kubernetes:

```sh
git fetch origin 1e092cd119359c8184bc5997211f85bca4ecc9ec
baseline_dir="$(mktemp -d)"
git archive 1e092cd119359c8184bc5997211f85bca4ecc9ec \
  | tar -x -C "$baseline_dir"

# Build the nested exporter before its Elasticsearch parent.
helm dependency build --skip-refresh \
  "$baseline_dir/local/prometheus-elasticsearch-exporter"
for dependency in anomaly-detection elasticsearch hbase kafka kafkaup-operator \
  zookeeper victoria-metrics-single clickhouse opentelemetry-collector \
  kubernetes-rbac-agent
do
  helm dependency build --skip-refresh "$baseline_dir/local/$dependency"
done
helm dependency build --skip-refresh "$baseline_dir/stable/suse-observability"

SERVICE_MONITOR_BASELINE_CHART="$baseline_dir/stable/suse-observability" \
  go test ./stable/suse-observability/test/... \
    -run '^TestServiceMonitorNamingBaselineReproduction$' -count=1
```

The explicit reproduction test uses `serviceMonitorCompatibilityValues` with
the three server/worker combinations in the table above and checks install and
upgrade rendering against the unchanged fixtures. It skips unless the baseline
path is supplied; normal tests need no old checkout or Git history.

Local lifecycle checks use a localhost-only Kubernetes API with a minimal
ServiceMonitor CRD and temporary ServiceMonitor/Service fixture charts.
They check final inventories, specifications, discovery labels, selected
Service/endpoint definitions, repeated upgrade, rollback and foreign ownership.
They do not run Prometheus Operator or scrape application metrics.

The baseline router monitor requests a `metrics` port absent from its selected
Service, and the HA receiver has no matching Service with these test values.
This step preserves that existing configuration; discovery comparisons must
preserve baseline results rather than require every monitor to yield a target.

Run the relevant tests from the repository root:

```sh
go test ./stable/suse-observability/test/... \
  -run 'TestServiceMonitor|TestMetricsServiceNames|TestResourceNaming' -count=1
```
