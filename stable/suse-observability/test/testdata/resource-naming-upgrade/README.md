# Resource naming upgrade baseline

These fixtures were rendered from the pre-migration commit
`fb9ac7bdd560540778e6b162582986a1506e68d7`, using that commit's main chart,
local dependencies, and `test/values/full.yaml`. Scenario overrides are defined
in `TestResourceNamingUpgradeCompatibility`. The namespace is explicitly
`observability`, independent of the developer's kubeconfig.

Each JSON file records the previous resource inventory and selected upgrade
contracts: selectors, StatefulSet service names and claim templates, PVC and PDB
specifications, Service ports, Secret names and keys, RBAC binding references,
and pod references to ServiceAccounts, pull Secrets, ConfigMaps, Secrets, and PVCs.
Job timestamp suffixes and the Elasticsearch test Pod's random suffix are
normalized; their name prefixes remain checked. Hook and regular resource
identities are recorded separately.

The test applies an explicit allowlist to the historical fixtures for the UI,
replication checker, and vmagent ConfigMap renames and their pod/RBAC references.
It then compares those expectations with a fresh rendering of the current chart.
Unexpected additions or removals also fail.

This is a compatibility contract, not a full manifest snapshot. It excludes
images, checksums, generated credential values, arbitrary configuration content,
and runtime behavior. Existing naming tests check router destinations and
ServiceMonitor selectors. S3Proxy monitoring is disabled in these fixtures
because the baseline cannot render it; the existing naming tests cover the
monitoring fix. These fixtures do not prove credential lookup reuse, hook
ordering, or live upgrade safety.

Run the comparison from the repository root:

```sh
go test ./stable/suse-observability/test/... -run '^TestResourceNamingUpgradeCompatibility$' -count=1
```

Normal test runs use the committed fixtures and need no Git history or baseline
checkout. Do not regenerate fixtures from the current chart to make a failure
pass. Review any change to the upgrade contract explicitly.

To reproduce the fixtures, export the exact baseline commit into a separate
directory, build its local Helm dependencies from that export (dependencies
before consumers), then run the current test with the baseline chart path:

```sh
RESOURCE_NAMING_BASELINE_CHART=/absolute/path/to/baseline/stable/suse-observability \
  go test ./stable/suse-observability/test/... -run '^TestResourceNamingUpgradeCompatibility$' -count=1
```

This explicit generation mode overwrites the four JSON fixtures without applying
the migration allowlist. Unset the variable for validation.
