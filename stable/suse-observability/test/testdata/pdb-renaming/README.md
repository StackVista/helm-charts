# PDB renaming baseline and inventory

R0 freezes the compatibility contract for the first proposed rename family.
R1 canonicalizes all main-chart PDB names, ignoring root/global fullname
overrides, prefixes and suffixes. Temporary eviction unavailability during
upgrade and rollback is accepted and documented.

The fixtures were generated from main-chart commit
`c851a02ae1f2f7eb64c5538a148194ea36eee4b4`, with Helm 3.19.0 and the
platform dependency archives recorded in `dependencies.json`. Check the full
commit identifier against Git when rebasing this work; the abbreviated source
identifier is `c851a02ae`.

The main-chart metadata, defaults, templates and values files were exported
from that commit into a temporary directory. Dependency archives were copied
from the built platform chart. Their hashes record the exact inputs used for
these snapshots; normal tests do not require packages to have identical archive
bytes, since Helm packaging timestamps may differ.

## Frozen fixtures

| Fixture | Server mode / sizing | Main-chart PDBs |
| --- | --- | --- |
| `split.json` | Split server, legacy values; also `150-ha` sizing | 17 |
| `mono.json` | Monolithic server, legacy values | 11 |
| `nonha.json` | `50-nonha` sizing profile | 10 |

The reference release is `suse-observability`, in namespace `observability`.
Inputs are `test/values/full.yaml`, with the corresponding
`global_sizing_150_ha.yaml` or `global_sizing_50_nonha.yaml` for sizing cases.
Server split is requested explicitly. Assistant, MCP, workload observer,
Kubernetes authorization and both VictoriaMetrics instances are enabled;
receiver/correlate worker splitting is disabled in the reference snapshots.
The sizing profile still controls effective server splitting and non-HA PDB
creation according to the existing chart.
The HA snapshot was verified to be identical to the split snapshot and shares
the same fixture to avoid duplicating it.

Each snapshot contains the complete rendered main-chart PDBs, keyed by component.
Only `helm.sh/chart` and `app.kubernetes.io/version` labels are removed so normal
chart/application version updates do not invalidate the naming contract.
Specifications, selectors, other labels, annotations and creation conditions
remain checked. Subchart-owned PDBs are outside this rename family.

`TestPDBNamingBaselineCompatibility` also exercises non-default and long release
names, a second namespace, fullname overrides and local/global prefixes/suffixes,
long overrides, split workers and disabled optional components, on install and
upgrade. Every scenario expects the canonical names from the unchanged
default-release fixtures; names are not obtained from the helpers being tested.

Keep these fixtures unchanged. The tests expect canonical PDB names even when
legacy fullname settings are supplied. No specification or selector changes are
allowed. The historical resource naming fixtures also remain unchanged; their test applies
an exact main-chart PDB metadata-name allowlist for all naming scenarios.

## Exact first-family inventory

All rows below have kind `PodDisruptionBudget`. Helpers are in
`stable/suse-observability/templates/_names.tpl`; paths in the table are relative
to the main chart's `templates/` directory. The default target identity is
`suse-observability-<suffix>`. Each row changes only its metadata name;
root/global fullname settings no longer affect these identities.

| Helper component | Suffix | Declaration | Classification |
| --- | --- | --- | --- |
| `aiAssistant` | `ai-assistant` | `ai-assistant/pdb-ai-assistant.yaml` | Canonical |
| `api` | `api` | `api/pdb-api.yaml` | Canonical |
| `authorizationSync` | `authorization-sync` | `authorizationSync/pdb-authorizationSync.yaml` | Canonical |
| `checks` | `checks` | `checks/pdb-checks.yaml` | Canonical |
| `correlate` | `correlate` | `correlate/pdb-correlate.yaml` | Canonical |
| `e2es` | `e2es` | `e2es/pdb-e2es.yaml` | Canonical |
| `victoriametrics` | `victoriametrics` | `global/pdb-victoriametrics.yaml` | Canonical |
| `healthSync` | `health-sync` | `healthSync/pdb-healthSync.yaml` | Canonical |
| `mcp` | `mcp` | `mcp/pdb-mcp.yaml` | Canonical |
| `notification` | `notification` | `notification/pdb-notification.yaml` | Canonical |
| `receiver` | `receiver` | `receiver/pdb-receiver.yaml` | Canonical |
| `router` | `router` | `router/pdb-router.yaml` | Canonical |
| `server` | `server` | `server/pdb-server.yaml` | Canonical |
| `state` | `state` | `state/pdb-state.yaml` | Canonical |
| `sync` | `sync` | `sync/pdb-sync.yaml` | Canonical |
| `ui` | `ui` | `ui/pdb-ui.yaml` | Canonical |
| `vmagent` | `vmagent` | `vmagent/pdb-vmagent.yaml` | Canonical |
| `workloadObserver` | `workload-observer` | `workload-observer/pdb-workload-observer.yaml` | Canonical |

For the default release these names are already canonical. With release `nightly`
and default naming values, they change from
`nightly-suse-observability-<suffix>` to `suse-observability-<suffix>`.
Names previously customized through fullname overrides, prefixes or suffixes
also become canonical.

The Kubernetes eviction API consumes the PDB selectors and budgets. Workloads
do not reference these PDBs by name. No in-chart endpoint, volume, credential,
RBAC subject or hook references a main-chart PDB identity. The runtime check
shows that old and new budgets coexist during replacement and temporarily block
evictions. Customer-managed automation selecting default PDB names must use
the new canonical names, including previously customized identities.

## Retained resources for this change

Every non-PDB identity remains unchanged. This includes all subchart PDBs,
Deployments, StatefulSets and their governing Services, PVCs, credential Secrets,
ConfigMaps, ServiceAccounts, RBAC, endpoints, monitoring resources and hooks.
Their later eligibility is assessed in the corresponding tracker steps.

No migration flag, PVC reassignment, data copying, credential migration or
custom operational script is introduced.

## R1 compatibility result

The candidate was checked with Helm 3.19.0 against an isolated localhost
Kubernetes 1.35.0 API server and disruption controller. The fixture charts
contained only PDBs rendered from the baseline and candidate parent charts.
Two synthetic Pod API objects and a ReplicaSet supplied a UI workload selector;
Pod readiness was set by the test. No application containers or platform were
deployed.

The baseline budget reported two healthy Pods and allowed one disruption.
Dry-run eviction returned HTTP 201. During upgrade, Helm created every new
PDB before deleting its predecessor. A test proxy paused deletion of the old
UI budget to inspect the overlap. A real dry-run eviction then returned HTTP
500 with `more than one PodDisruptionBudget` in the response. The proxy
extended the interval for observation; it did not establish its normal duration.

Upgrade subsequently completed, leaving the expected candidate inventory and
allowing eviction again. Repeated upgrade preserved PDB UIDs. Rollback restored
legacy names without replacing fixture Pods. That initial prototype retained
custom names with fullname settings; the final implementation ignores those
settings, and the render cases verify the additional metadata-name changes.
A destination owned by another release correctly rejected the candidate
upgrade before changing either the original or foreign budgets.

Temporary unavailability during the transition is accepted. The final chart
adopts canonical names in every naming scenario. Frozen fixtures stay unchanged
and tests allow only the reviewed PDB metadata-name changes.

The acceptance checks are:

- Ignore root/global fullname settings and document the canonical names.
- Preserve selectors, specifications, labels and creation conditions.
- Check destination ownership and name collisions. The platform already creates
  fixed UI/MCP and other names within a namespace; canonicalizing PDBs cannot
  introduce support for installing multiple platforms in that same namespace.
  Different namespaces remain independent, and Helm ownership stays with the
  original release.
- In a disposable cluster, install the baseline, upgrade normally, upgrade
  again, and roll back. Verify one intended final PDB per enabled component.
- Exercise eviction during replacement and document its temporary rejection.
  Verify that completed upgrade and rollback restore normal eviction behavior
  without customer repair.
- Verify a conflicting destination's Helm ownership rejects the upgrade
  before changing original or foreign budgets.

These focused API/controller checks do not prove full platform lifecycle
behavior or Argo CD sync safety. The baseline tests separately preserve the
render contract.

Run the focused tests from the repository root:

```sh
go test ./stable/suse-observability/test/... \
  -run 'TestPDB|TestResourceNamingUpgradeCompatibility' -count=1
```
