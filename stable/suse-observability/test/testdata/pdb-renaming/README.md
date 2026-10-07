# PDB renaming baseline and inventory

R0 freezes the compatibility contract for the first proposed rename family.
No production names change in this step.

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
upgrade. Expected legacy name prefixes are explicit test inputs; they are not
obtained from the resource-name helpers being tested.

Keep these fixtures unchanged when implementing R1. Review an exact
`metadata.name` allowlist in the test rather than regenerating the baseline
from the renamed chart. The existing historical resource naming fixtures are
also unchanged and continue checking storage and other resource references.

## Exact first-family inventory

All rows below have kind `PodDisruptionBudget`. Helpers are in
`stable/suse-observability/templates/_names.tpl`; paths in the table are relative
to the main chart's `templates/` directory. The proposed default identity is
`suse-observability-<suffix>`.

| Helper component | Suffix | Declaration | Classification |
| --- | --- | --- | --- |
| `aiAssistant` | `ai-assistant` | `ai-assistant/pdb-ai-assistant.yaml` | Eligible for R1 review |
| `api` | `api` | `api/pdb-api.yaml` | Eligible for R1 review |
| `authorizationSync` | `authorization-sync` | `authorizationSync/pdb-authorizationSync.yaml` | Eligible for R1 review |
| `checks` | `checks` | `checks/pdb-checks.yaml` | Eligible for R1 review |
| `correlate` | `correlate` | `correlate/pdb-correlate.yaml` | Eligible for R1 review |
| `e2es` | `e2es` | `e2es/pdb-e2es.yaml` | Eligible for R1 review |
| `victoriametrics` | `victoriametrics` | `global/pdb-victoriametrics.yaml` | Eligible for R1 review |
| `healthSync` | `health-sync` | `healthSync/pdb-healthSync.yaml` | Eligible for R1 review |
| `mcp` | `mcp` | `mcp/pdb-mcp.yaml` | Eligible for R1 review |
| `notification` | `notification` | `notification/pdb-notification.yaml` | Eligible for R1 review |
| `receiver` | `receiver` | `receiver/pdb-receiver.yaml` | Eligible for R1 review |
| `router` | `router` | `router/pdb-router.yaml` | Eligible for R1 review |
| `server` | `server` | `server/pdb-server.yaml` | Eligible for R1 review |
| `state` | `state` | `state/pdb-state.yaml` | Eligible for R1 review |
| `sync` | `sync` | `sync/pdb-sync.yaml` | Eligible for R1 review |
| `ui` | `ui` | `ui/pdb-ui.yaml` | Eligible for R1 review |
| `vmagent` | `vmagent` | `vmagent/pdb-vmagent.yaml` | Eligible for R1 review |
| `workloadObserver` | `workload-observer` | `workload-observer/pdb-workload-observer.yaml` | Eligible for R1 review |

For the default release these names are already canonical. With release `nightly`
and default naming values, they currently use
`nightly-suse-observability-<suffix>`. Existing fullname overrides and
prefixes/suffixes must retain their current names in R1.

The Kubernetes eviction API consumes the PDB selectors and budgets. Workloads
do not reference these PDBs by name. No in-chart endpoint, volume, credential,
RBAC subject or hook references a main-chart PDB identity. Runtime compatibility
still needs testing, because old and new budgets can coexist during replacement.
Customer-managed automation may select a PDB by name; such a reference requires
retaining the name, rather than assuming it can be changed safely.

## Retained resources for this change

Every non-PDB identity remains unchanged. This includes all subchart PDBs,
Deployments, StatefulSets and their governing Services, PVCs, credential Secrets,
ConfigMaps, ServiceAccounts, RBAC, endpoints, monitoring resources and hooks.
Their later eligibility is assessed in the corresponding tracker steps.

No migration flag, PVC reassignment, data copying, credential migration or
custom operational script is introduced.

## R1 acceptance gates

- Preserve explicit naming overrides and existing externally selected PDB names.
- Preserve selectors, specifications, labels and creation conditions.
- Check destination ownership and name collisions. The platform already creates
  fixed UI/MCP and other names within a namespace; canonicalizing PDBs cannot
  introduce support for installing multiple platforms in that same namespace.
  Different namespaces remain independent, and Helm ownership stays with the
  original release.
- In a disposable cluster, install the baseline, upgrade normally, upgrade
  again, and roll back. Verify one intended final PDB per enabled component.
- Exercise eviction during replacement. Kubernetes can reject eviction when
  more than one PDB selects a Pod; verify that the transition does not create
  an unacceptable protection gap or require customer repair.
- Check failed-upgrade cleanup and rollback when both names are temporarily
  present. Keep a legacy identity if this cannot satisfy the compatibility rule.

These live gates are intentionally pending. The baseline tests prove the
current render contract, not the safety of a future object replacement.

Run the focused tests from the repository root:

```sh
go test ./stable/suse-observability/test/... \
  -run 'TestPDB|TestResourceNamingUpgradeCompatibility' -count=1
```
