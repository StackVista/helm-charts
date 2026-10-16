# Bundled collector peripheral renaming

Pre-change source: `2d7562ce7006d406969ce05a56caa165d410c1c9`.
The extraction baseline, before the intervening renaming steps, is
`c851a02ae1f2f7eb64c5538a148194ea36eee4b4`.
Both revisions are resolved and exported from repository Git history for
independent validation.

| Kind | Previous name | New name |
| --- | --- | --- |
| PodDisruptionBudget | `<collector-fullname>` | `suse-observability-otel-collector` |
| ServiceMonitor | `<collector-fullname>` | `suse-observability-otel-collector` |
| PodMonitor | `<collector-fullname>-agent` | `suse-observability-otel-collector-agent` |
| PrometheusRule | `<collector-fullname>` | `suse-observability-otel-collector` |
| NetworkPolicy | `<collector-fullname>` | `suse-observability-otel-collector` |

The default fullname is already `suse-observability-otel-collector`.
Only custom or cleared collector fullname settings change these identities.
The parent supplies five overrides of existing subchart helper interfaces.
Standalone subchart defaults are untouched.

The normal compatibility test removes these five parent definitions in a
temporary chart, allowing the original subchart definitions to resolve the names.
It compares the complete platform with only five metadata-name changes allowed.
Configuration, checksums, endpoints, controllers and Pod templates, selector and
release labels, external choices, autoscalers, Secrets, accounts, RBAC, persistent
claims, hook scripts and every other resource must stay unchanged.
Additional checks ensure peripheral selectors still match the retained Pods.
Current creation conditions remain unchanged in all three collector modes.

The independent Git-baseline test copies the actual pre-change chart, including
its built dependency graph, and adds only the five current parent definitions.
It compares that historical chart before/after this isolated naming step.
Comparing the baseline directly with today's chart would also include later
RBAC, HBase, collector relay configuration and other unrelated changes.
The independent-step test applies the same procedure to the extraction baseline,
without applying intervening renames. The ordinary compatibility test still
verifies the current chart and dependencies with only these five metadata-name
changes allowed. Neither historical test rewrites fixtures or expands its
allowlist to accommodate unrelated changes.

To reproduce, export both `local/` and `stable/suse-observability/` from each
revision with `git archive`. Build each `file://` dependency graph recursively
from children to parents using Helm 3 `dependency build --skip-refresh`; build
the current chart's local graph the same way. From the current repository root:

```sh
KUBECONFIG=/tmp/offline-kubeconfig.yaml \
COLLECTOR_PERIPHERAL_BASELINE_CHART=/tmp/pre-change/stable/suse-observability \
COLLECTOR_PERIPHERAL_EXTRACTION_CHART=/tmp/extraction/stable/suse-observability \
go test ./stable/suse-observability/test/... \
  -run '^TestCollectorPeripheralNaming' -count=1
```

Use a synthetic offline kubeconfig, for example with server
`http://127.0.0.1:1`. All checks render locally and do not contact Kubernetes or
load installed credentials. Scenarios cover modes, releases, namespaces,
collector/parent overrides, long names, external configuration/accounts,
persistent claims, autoscaling, Argo CD, sizing, custom rules/metadata and
disabled-resource conditions.

Metadata renames create and delete objects; unchanged workloads do not need
reference patches or subsequent renames. PDB/monitor overlap and stale-object
pruning/recovery follow the chart README. Actual Kubernetes/Prometheus/network
controller reconciliation remains an integration check; no recovery-time bound
is claimed.
