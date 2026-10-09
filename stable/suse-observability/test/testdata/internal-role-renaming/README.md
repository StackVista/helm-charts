# Internal Role and RoleBinding renaming

Pre-change source: `7979fb76f8d9edd5f27d119265bf2f947fe38dba`.
Extraction baseline: `c851a02ae1f2f7eb64c5538a148194ea36eee4b4`.
Both revisions are resolved and exported from repository Git history.

| Kind | Previous name | Canonical name |
| --- | --- | --- |
| Router hook Role | `<parent-common.fullname.short>-router-mode` | `suse-observability-router-mode` |
| Router hook RoleBinding | `<parent-common.fullname.short>-router-mode` | `suse-observability-router-mode` |
| Bundled operator Role | `<operator-common.fullname.short>-kafkaup` | `suse-observability-kafkaup` |
| Bundled operator RoleBinding | `<operator-common.fullname.short>-kafkaup-binding` | `suse-observability-kafkaup-binding` |

The existing declarations and their `roleRef` consumers already use dedicated
helpers. Change the two router helpers and override the two operator helpers
in the platform only. Standalone subcharts are unchanged.
Both roles and bindings move together; canonical bindings keep immutable
`roleRef` stable during later fullname changes. The RBAC-agent pair retains
its identity because of its relationship with application authorization.
Pod-reader, instance-access and other externally referenced Roles are retained.

Normal tests restore only these four old definitions in a temporary chart.
Complete manifests are compared, allowing exactly four metadata names and
the two associated `roleRef.name` fields to change. Rules, subjects, namespace
scope, hook annotations/order/deletion policies, accounts, Jobs, controllers,
Pod templates, checksums, configuration, credentials, storage and every other
resource remain equal. Each binding must resolve an existing Role and account.
Creation conditions, release labels and external Kafka choices are preserved.

The independent Git comparison uses the actual pre-change source. The
independent-step test copies the extraction baseline and applies only these
four definitions, without the intervening ConfigMap, workload, monitoring or
RBAC renames. Test utilities used here already exist in that baseline.
Subsequent upgrade and rollback checks apply Helm 3 three-way strategic merge
patches to canonical bindings after changing parent and operator fullnames.
They assert that `roleRef` never enters a patch and that mutable subjects and
metadata converge while retaining the current resource version.
No historical snapshots are regenerated.

To reproduce, export `local/` and `stable/suse-observability/` from both revisions
using `git archive`. Rebuild all `file://` dependencies from children to parents
with Helm 3 `dependency build --skip-refresh`, and build current dependencies
the same way. From the current repository root:

```sh
KUBECONFIG=/tmp/offline-kubeconfig.yaml \
INTERNAL_ROLE_BASELINE_CHART=/tmp/pre-change/stable/suse-observability \
INTERNAL_ROLE_EXTRACTION_CHART=/tmp/extraction/stable/suse-observability \
go test ./stable/suse-observability/test/... \
  -run '^TestInternalRoleNaming' -count=1
```

Use a synthetic kubeconfig with server `http://127.0.0.1:1`.
All checks render locally; no Kubernetes environment is accessed.
Sixteen scenarios cover install/upgrade, releases, namespaces, root/global
and operator naming settings, long names, Argo CD, sizing, monolithic servers,
external Kafka selectors and disabled components.

Actual Helm/Argo CD replacement, hook execution and API authorization remain
integration checks. Interrupted operations and failed legacy hooks can leave
old RBAC; pruning, cleanup and external automation considerations are documented
in the chart README. No recovery-time guarantee is claimed.
