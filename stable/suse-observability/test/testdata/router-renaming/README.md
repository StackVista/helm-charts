# Router Deployment naming compatibility

The three snapshots were generated before the router rename from revision
`bc48f2d39214bdf67f552ab3a93f9100b3ed80f4`, with its packaged dependencies and
Helm 3.19.0. The source revision is an ancestor of this change.

Only the Deployment name moves from `<legacy-fullname>-router` to
`suse-observability-router`. The Service, selectors, accounts, RBAC, credential
references, static and dynamic ConfigMaps, hook resources and Ingress keep their
identities. Automatic-mode scripts retain their dynamic ConfigMap target and
select Deployments by component and Helm release labels in the release namespace.
This reaches the old router before an upgrade and the replacement afterwards,
including both during overlap. Empty selection skips restarting; disappearance
while waiting is handled as before.
Rollout status is polled without watch, matching the existing hook Role's
get/list/patch permissions.

The router is a stateless Envoy proxy. Its default volumes project configuration;
it has no PVC or background writer. Old and new proxies can temporarily coexist
behind the retained Service, or connections can be interrupted during
replacement. Custom images or containers require their own lifecycle review.
The RBAC agent and Kafka upgrade operator remain unchanged because concurrent
instances can perform conflicting background work.

`active.json`, `maintenance.json` and `automatic.json` freeze complete Deployment
specifications; the automatic fixture also freezes the script ConfigMap.
Inputs are release `nightly`, namespace `observability`, `values/full.yaml`, and
`routerNamingFixtureValues(mode)`. Main-chart version and appVersion are pinned
to `0.0.0` and `naming-fixture` in a temporary copy, and the router image tag is
pinned to `naming-fixture`. This keeps configuration checksums and metadata
independent of unrelated chart/image version updates. The fixtures contain only
synthetic test values.

To independently reproduce them, export that revision, build its local
dependencies, and point `ROUTER_NAMING_BASELINE_CHART` at the exported platform
chart:

```shell
ROUTER_NAMING_BASELINE_CHART=/path/to/export/stable/suse-observability \
  go test ./stable/suse-observability/test/... \
  -run '^TestRouterNamingBaselineReproduction$' -count=1
```

This check compares the baseline to the frozen fixtures; it never rewrites them.
Install and upgrade rendering must reproduce identical snapshots. Additional
before/after comparisons cover namespaces, root/global overrides and affixes,
long names, split/monolithic services, Argo CD, and all three router modes.

`TestRouterModeScriptsReachOldAndNewDeployments` executes both rendered scripts
with a synthetic kubectl. It covers old/new/overlapping/absent routers, pending
rollouts, disappearance while waiting, list errors, and namespace/release
isolation. These tests do not connect to Kubernetes or run Envoy; actual
controller replacement, readiness, routed traffic and rollback remain
integration checks.
