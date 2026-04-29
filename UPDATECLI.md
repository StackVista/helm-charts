# UPDATECLI.md

Guidance for working with updatecli pipelines in this repository. For general Helm chart work, see `CLAUDE.md` and `AGENTS.md`.

## Overview

Updatecli automates Docker image tag updates across Helm chart values files. It watches `quay.io/stackstate/*` images for new release tags and writes the latest tag into the appropriate `values.yaml` keys.

## File Layout

```
updatecli/
├── updatecli.d/
│   ├── update-docker-images/
│   │   ├── update.yaml              # Main pipeline: sources + targets for 21 images
│   │   └── README.md
│   └── finalize-docker-images/
│       └── update.yaml              # Post-update: bump agent chart version, create MR
├── values.d/
│   ├── values.yaml                  # SCM + registry config (templated into pipelines)
│   └── values-local.yaml            # Local dev override (sets clone directory)
└── updatecli-image-mapping.md       # Reference table: image -> values file key(s)
```

## Pipeline Architecture

There are two pipelines that run sequentially in CI:

### 1. `update-docker-images/update.yaml` (main pipeline)

- **Pipeline ID:** `docker-images`
- Contains 21 `sources` (one per Docker image) and 21 `targets` (the values keys to update).
- All targets are **chained via `dependson`** so they execute sequentially. This is intentional: split pipelines ran in parallel and only the last write survived because they all modify the same values file.
- Almost all targets write to `stable/suse-observability/values.yaml`. Two exceptions write to `stable/suse-observability-agent/values.yaml` (`containerToolsSuseObservabilityAgent`, `kubernetesRbacAgentSuseObservabilityAgent`).
- Some images update multiple keys (use `keys:` list instead of `key:`).

### 2. `finalize-docker-images/update.yaml` (post-update pipeline)

- Shares the same `pipelineid: docker-images` so it operates on the same working branch (`updatecli-docker-images-master`).
- If `stable/suse-observability-agent/values.yaml` changed vs master, runs `scripts/bump-chart-version/bump_chart_version.py suse-observability-agent` to bump the agent chart version.
- Creates a GitLab merge request titled `[master] Bump helm chart docker images`.

### CI Jobs (in `.gitlab-ci.jsonnet`)

Three jobs in the `update` stage, triggered only when `$RUN_UPDATECLI` is set:

```
update_helm_chart_docker_images          # runs update-docker-images pipeline
  -> finalize_helm_chart_docker_images   # runs finalize-docker-images pipeline
    -> open_updatecli_docker_images_mr   # ensures MR exists via .gitlab/open_updatecli_mr.sh
```

## Source Configuration

Every source is a `dockerimage` kind pulling from `quay.io/stackstate/<name>` with `architecture: linux/amd64`.

### Tag Filter Patterns

Images use one of four tag formats. When adding a new image, pick the matching `tagfilter` and `versionfilter`:

| Format | Example Tag | `tagfilter` | `versionfilter` regex |
|--------|-------------|-------------|----------------------|
| Standard | `v1.109.0-614527d8-release-138` | `.*-[a-f0-9]{8}-release-[0-9]+$` | `.*-(\d+)$` |
| Reversed | `f40221cf-76-release` | `^[a-f0-9]{8}-[0-9]+-release$` | `^[a-f0-9]{8}-([0-9]+)-release$` |
| GitHub main | `1.8.6-fa52bb17-main-4` | `^[0-9]+\.[0-9]+\.[0-9]+-[a-f0-9]{8}-main-[0-9]+$` | `.*-(\d+)$` |
| Semver build | `1.8.3-573` | `^[0-9]+\.[0-9]+\.[0-9]+-[0-9]+$` | `.*-(\d+)$` |

All use `versionfilter.kind: regex/semver` which extracts a build number for ordering.

## Target Configuration

Targets are `yaml` kind with `scmid: default`. Key fields:

- `sourceid` - links to the source providing the tag value
- `dependson` - `["target#previousTarget"]` to enforce sequential execution
- `spec.file` - the values file to update
- `spec.key` or `spec.keys` - JSONPath(s) to the image tag field(s)

### Special Cases

- **hadoop**: Uses a `transformers.findsubmatch` to strip the semver prefix from the tag. The values file expects `java21-8-hash-buildId` but the Docker tag is `version-java21-8-hash-release-buildId`.
- **container-tools**: Updates 5 keys across 3 targets (suse-observability, victoria-metrics backup cron, suse-observability-agent). Matches only standard `main` tags; `1.8.6_dev-*` tags are CI/dev images and must not be used in customer-runtime chart defaults.
- **kubernetes-rbac-agent (suse-observability target)**: May need to be temporarily disabled if the tag key doesn't exist on master yet (updatecli clones from remote).

## Values Templating

Pipeline YAML files use Go templates referencing values from `values.d/values.yaml`:

- `{{ .scm.owner }}`, `{{ .scm.repository }}`, `{{ .scm.branch }}` - GitLab project coordinates
- `{{ env .scm.userEnv }}`, `{{ requiredEnv .scm.tokenEnv }}` - credentials from environment
- `{{ .scm.directory }}` - optional clone directory (only set for local testing)

## Adding a New Image

1. Add a **source** entry in `update-docker-images/update.yaml` with the image name, `tagfilter`, and `versionfilter` matching the image's tag format (see table above).
2. Add a **target** entry with:
   - `sourceid` pointing to the new source
   - `dependson` pointing to the last target in the chain (currently `zookeeper`)
   - `spec.file` and `spec.key`/`spec.keys` for the values file location(s)
3. Update the `dependson` so the new target becomes the last in the chain (or insert it alphabetically and fix the chain).
4. Update `updatecli-image-mapping.md` with the new image's mapping.
5. If the target writes to `stable/suse-observability-agent/values.yaml`, the finalize pipeline will automatically bump the agent chart version.

## Removing an Image

1. Remove the source and target entries from `update-docker-images/update.yaml`.
2. Fix the `dependson` chain so the target before and after the removed one are linked.
3. Update `updatecli-image-mapping.md`.

## Local Testing

Dry-run without committing or pushing (requires `GITLAB_TOKEN` in `.env`):

```bash
source .env
docker run --rm --network host -v $(pwd):/workspace:z -w /workspace -e GITLAB_TOKEN \
  quay.io/stackstate/container-tools:1.8.6_dev-fa52bb17-main-4 \
  updatecli diff \
    -c updatecli/updatecli.d/update-docker-images/ \
    -v updatecli/values.d/values.yaml
```

To apply locally and inspect changes (writes to `updatecli-work/` subdirectory):

```bash
source .env
docker run --rm --network host -v $(pwd):/workspace:z -w /workspace -e GITLAB_TOKEN \
  quay.io/stackstate/container-tools:1.8.6_dev-fa52bb17-main-4 \
  updatecli apply --commit=false --push=false --clean=false \
    -c updatecli/updatecli.d/update-docker-images/ \
    -v updatecli/values.d/values.yaml -v updatecli/values.d/values-local.yaml
```

**Warning:** Never set `scm.directory` to `/workspace` in local values - updatecli may clear it before cloning.

## Related Scripts

- **`scripts/bump-chart-version/bump_chart_version.py`** - Bumps a chart's version and cascades to all dependents via the reverse dependency graph. Called by the finalize pipeline for agent chart changes. Also useful standalone:
  ```bash
  python3 scripts/bump-chart-version/bump_chart_version.py common              # patch bump
  python3 scripts/bump-chart-version/bump_chart_version.py -t minor elasticsearch
  python3 scripts/bump-chart-version/bump_chart_version.py -d -v common        # dry run
  python3 scripts/bump-chart-version/bump_chart_version.py --check common      # validate deps
  ```
- **`.gitlab/open_updatecli_mr.sh`** - Ensures a GitLab merge request exists for the updatecli branch. Idempotent (skips if MR already open).

## Required Environment Variables

| Variable | Purpose |
|----------|---------|
| `GITLAB_TOKEN` | GitLab API token (required) |
| `UPDATE_CLI_USER` | Git committer name |
| `UPDATE_CLI_EMAIL` | Git committer email |
| `UPDATE_CLI_PGP_KEY` | GPG signing key |
| `UPDATE_CLI_PGP_PASSPHRASE` | GPG key passphrase |
| `REGISTRY_USER` | quay.io registry username |
| `REGISTRY_PASSWORD` | quay.io registry password |
