# CI/CD

GitHub Actions reference for this repository. See the design spec at
[`docs/superpowers/specs/2026-06-01-gitlab-to-github-ci-design.md`](superpowers/specs/2026-06-01-gitlab-to-github-ci-design.md)
for the full rationale and cutover plan.

## Workflow map

| Workflow | Triggers | Purpose |
|---|---|---|
| `ci.yml` | `push` (any ref), `workflow_dispatch` | Build, validate, test, image-validation, version-bump check, resource-usage test. Fans out per chart via matrix from `detect-changes`. |
| `publish.yml` | `push` to `master`, chart tags, `workflow_dispatch` | Builds chart and pushes to the internal chartmuseum. Gated by `vars.PUBLISH_ENABLED`. |
| `publish-public.yml` | `workflow_dispatch` | Manual prerelease + release push of `suse-observability-agent` / `suse-observability-values` to the public chartmuseum. Gated. |
| `publish-rancher.yml` | `workflow_dispatch` | Manual prerelease + release push of the agent chart to the Rancher OCI registry. Gated. |
| `updatecli.yml` | `push` (paths: `updatecli/**`), `workflow_dispatch` | On push: read-only `validate_config` + `validate_diff` for updatecli edits. On dispatch (gated): runs one of three updatecli pipelines (`docker_images` / `stackstate` / `stackpacks`). |
| `beest.yml` | `workflow_dispatch` | POSTs to the beest GitLab pipeline trigger with the current chart version. Gated. |
| `main.yml` | tag push | Creates a GitHub Release with the backup-scripts artifact. |
| `mirror-gitlab.yml` | hourly schedule, `workflow_dispatch` | Mirrors GitLab refs into GitHub during the parallel-run period. Temporary — removed at post-cutover hardening (Task 14). |

## Runner mapping

| Label | Resources | Used for |
|---|---|---|
| `docker` | 2 CPU + 4Gi, DinD sidecar | All light jobs: build, validate, lint, image-validation, version-check, publish, updatecli, beest, release. Required for `container:` directive. |
| `xlarge` | 6 CPU + 20Gi, DinD sidecar | Go terratest jobs (`test_stable`, `test_local`, `resource_usage`). Matches the old GitLab `sts-k8s-xl-runner` on CPU. |
| `small` | 1 CPU + 3Gi, no DinD | `main.yml` (release-on-tag), `mirror-gitlab.yml`. Cannot host `container:` workflows. |

The `cpu-xlarge`, `cpu-xxlarge`, `deploy`, and `infra` runner labels are intentionally NOT used by this repo (no DinD, or wrong purpose).

## Container images

| Image | Contents | Used by |
|---|---|---|
| `${REGISTRY_QUAY_URL}/stackstate/sts-ci-images:stackstate-helm-test-4423bdaa` | helm, kubectl, yq, jsonnet, Go | `detect`, `build`, `test`, `resource_usage` |
| `${REGISTRY_QUAY_URL}/stackstate/sts-ci-images:stackstate-devops-d4e06284` | helm, skopeo, updatecli, gh, aws, cm-push | publish jobs, image-validation, updatecli, beest |
| `quay.io/helmpack/chart-testing:v3.14.0` | yamale, yamllint, helm | chart validation jobs (no Quay creds needed) |

`${REGISTRY_QUAY_URL}` resolves to `registry.tooling.stackstate.io/quay` — the org's Harbor instance, exposing a **proxy-cache project named `quay`** that mirrors `quay.io/*` transparently. So the actual pull URL is `registry.tooling.stackstate.io/quay/stackstate/sts-ci-images:<tag>`, fronting the same upstream image GitLab CI consumes. The first two images require `credentials: { username: vars.REGISTRY_USER, password: secrets.REGISTRY_PASSWORD }` on every job that uses them. The DinD daemon's `docker pull` is a separate auth context from the runner pod's `imagePullSecrets`, so the credentials must be declared at the job level even when the cluster's pull secrets cover the same registry. `REGISTRY_USER` is an organization-level **variable** (the Harbor robot name `robot$github_cicd` is not sensitive); `REGISTRY_PASSWORD` is an organization-level **secret**.

> **Known prerequisite (Task 0):** the `robot$github_cicd` robot must have read access on the Harbor `quay` proxy-cache project. Initial workflow runs hit `Docker login for 'registry.tooling.stackstate.io' failed with exit code 1`. Resolution path is with the platform team — either grant the existing robot read on `quay/*`, or provision a different robot and update `REGISTRY_USER`/`REGISTRY_PASSWORD` accordingly.

Every `container:` block also carries `options: --user 0:0` — running as root inside the container avoids `EACCES` on `/__w/_temp` where Node-based actions (e.g. `actions/checkout@v4`) write state. The runner pod's `fsGroup` (1001) owns that path and the image's default user isn't in that group; root bypasses the mismatch. Matches the behavior GitLab's Kubernetes executor gives the same images by default.

## Required secrets and variables

### Repository variables

| Name | Purpose |
|---|---|
| `REGISTRY_QUAY_URL` | Harbor proxy-cache path for Quay images. Set to `registry.tooling.stackstate.io/quay` (Harbor mirrors `quay.io/X` at `<host>/quay/X`). Used as the prefix for `sts-ci-images` pulls in all `container:` blocks. |
| `REGISTRY_USER` | Harbor robot account name (`robot$github_cicd`). Non-sensitive; lives as a variable. |
| `PUBLISH_ENABLED` | `'false'` during the parallel-run, `'true'` at cutover. A single flip enables every gated job. |

### Repository secrets

| Name | Purpose |
|---|---|
| `REGISTRY_PASSWORD` | Harbor robot password paired with `vars.REGISTRY_USER`. Robot must have read access on the `quay` proxy-cache project — see the "Known prerequisite" note above. Both provisioned at the organization level. |
| `GH_APP_CLIENT_ID`, `GH_APP_PRIVATE_KEY` | GitHub App credentials for bot-identity commits and PR creation. |
| `HELM_CHARTS_PAT` | Passed to updatecli; populated at runtime with a GitHub App installation token. Name preserved from the GitLab era; the value is no longer a GitLab PAT. |
| `CHARTMUSEUM_INTERNAL_URL`, `CHARTMUSEUM_INTERNAL_USERNAME`, `CHARTMUSEUM_INTERNAL_PASSWORD` | Internal chartmuseum (used by `publish.yml`). |
| `CHARTMUSEUM_URL`, `CHARTMUSEUM_USERNAME`, `CHARTMUSEUM_PASSWORD` | Public chartmuseum (used by `publish-public.yml`). |
| `RANCHER_HELM_REGISTRY_USERNAME`, `RANCHER_HELM_REGISTRY_PASSWORD`, `RANCHER_HELM_REGISTRY_DISTRIBUTION_ID` | Rancher S3 + CloudFront (used by `publish-rancher.yml`). |
| `RANCHER_CONTAINER_REGISTRY_USERNAME`, `RANCHER_CONTAINER_REGISTRY_PASSWORD` | Rancher container registry (used by `publish-rancher.yml`). |
| `STS_BEEST_TRIGGER_TOKEN`, `BEEST_PROJECT_ID` | GitLab beest trigger (used by `beest.yml`). |
| `GITLAB_TOKEN` | Temporary; only used by `mirror-gitlab.yml`. Removed at post-cutover. |

## `PUBLISH_ENABLED` semantics

With `PUBLISH_ENABLED='false'` (parallel-run):

- `publish.yml`'s `push_to_internal` job skips entirely; `detect` and `build_stable` still run to produce artifacts.
- `publish-public.yml`, `publish-rancher.yml`, `updatecli.yml`, and `beest.yml` skip every job.
- `ci.yml` runs normally — build/validate/test are always desired and are not gated.

With `PUBLISH_ENABLED='true'` (post-cutover) the gated jobs run for real.

To flip the gate: **Settings → Secrets and variables → Actions → Variables → `PUBLISH_ENABLED`**.

## Triggering manual workflows

Via the UI: **Actions** tab → select the workflow → **Run workflow** → fill inputs → submit.

Via the CLI:

```bash
gh workflow run updatecli.yml \
  -f pipeline=docker_images \
  -f target_branch=STAC-12345/my-feature
```

## Branch conventions

Branches are prefixed with the Jira ticket number (e.g. `STAC-12345/my-feature`).

During the parallel-run period the GitLab mirror force-pushes refs by name, so a branch created only on GitHub is safe **unless its name collides with a branch that exists in GitLab**. The Jira-number convention makes collisions unlikely. A `gh/` prefix is documented as optional for experimental GitHub-only branches (e.g. when iterating on the workflows themselves).

See [`.github/workflows/mirror-gitlab.yml`](../.github/workflows/mirror-gitlab.yml) and the design spec for details.

## Detect-changes mechanism

`scripts/ci/detect-changes.sh` (wrapped by the `.github/actions/detect-changes` composite action) computes which charts are affected by the current diff and emits a JSON matrix consumed by the downstream matrix jobs. The chart list is read directly from `.jsonnet-libs/extras/helm_chart_repo/variables.libsonnet` — the single source of truth shared with the GitLab CI definitions.

The diff base falls back gracefully:

1. `$BASE_SHA` if set explicitly,
2. otherwise `$GITHUB_EVENT_BEFORE` (push events),
3. otherwise `origin/master`,
4. over-fire on empty diff (treat all charts as changed).

## Local scripts

Most scripts under `scripts/ci/` are ported from `.gitlab/` and are runnable locally for debugging. For example:

```bash
BASE_SHA=origin/master ./scripts/ci/detect-changes.sh
```

prints the matrix outputs to stdout.
