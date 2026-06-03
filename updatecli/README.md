# updatecli

Pipelines and shared values for the three updatecli flows in this repo:

- `updatecli.d/update-docker-images/` — bumps the `quay.io/stackstate/*` image tags across the SUSE Observability charts and opens a PR.
- `updatecli.d/update-stackstate/` — bumps the StackState image tag, the chart `appVersion`, and the StackGraph version. Commits directly to the target branch (no PR).
- `updatecli.d/update-stackpacks/` — bumps the StackPacks image version. Commits directly to the target branch (no PR).

`values.d/values.yaml` holds the shared SCM/source configuration that all three pipelines template into their yaml.

For pipeline architecture, tag-format conventions, and how to add or remove images, see [`../UPDATECLI.md`](../UPDATECLI.md).

## How to trigger

The pipelines run as GitHub Actions via the [`updatecli.yml`](../.github/workflows/updatecli.yml) workflow (`workflow_dispatch` only — they never run on push).

The safe way to validate a change to this directory is to **trigger a manual run on a test branch** and point updatecli at that branch instead of `master`. None of the production data ends up on `master` unless you set `target_branch=master` explicitly.

### Via the GitHub UI

1. Push your changes to a branch (e.g. `STAC-1234/updatecli-tweak`).
2. **Actions → Updatecli → Run workflow.**
3. Pick the workflow branch (the branch the workflow file is read from — usually `master`).
4. Fill the inputs (see below).
5. Run. Inspect the resulting commits / PR on `target_branch`.

### Via the CLI

```bash
gh workflow run updatecli.yml \
  -f pipeline=docker_images \
  -f target_branch=STAC-1234/updatecli-tweak

gh workflow run updatecli.yml \
  -f pipeline=stackstate \
  -f target_branch=STAC-1234/updatecli-tweak \
  -f stackstate_source_branch=master

gh workflow run updatecli.yml \
  -f pipeline=stackpacks \
  -f target_branch=STAC-1234/updatecli-tweak \
  -f stackpacks_source_branch=master
```

### Inputs

| Input | Meaning |
|---|---|
| `pipeline` | One of `docker_images`, `stackstate`, `stackpacks`. Required. |
| `target_branch` | Branch in `StackVista/helm-charts-internal` that updatecli clones, commits to, and (for `docker_images`) opens the PR against. **Use a test branch when validating changes.** |
| `stackstate_source_branch` | Branch segment in `stackstate-receiver` Docker tags (e.g. `master`, `stac-1234-feature`). Also used to fetch `stackgraph_version` from that branch of `stackvista/stackstate`. Only relevant for the `stackstate` pipeline. |
| `stackpacks_source_branch` | Branch segment in `stackpacks` Docker tags. Only relevant for the `stackpacks` pipeline. |

### Example: validate the stackstate pipeline against a feature branch

```
pipeline                   = stackstate
target_branch              = STAC-1234/updatecli-tweak
stackstate_source_branch   = master
```

The job will clone `STAC-1234/updatecli-tweak`, fetch the latest `master` `stackstate-receiver` tag + `stackgraph_version`, apply the value updates, and push a single squashed commit back to `STAC-1234/updatecli-tweak`. Inspect the commit; if it looks wrong, fix the pipeline yaml and re-run.

## PR validation

When you push an updatecli config change, `validate_updatecli_config` (yamllint) and `validate_updatecli_diff` (`updatecli diff` against all three pipelines) run automatically as jobs in [`ci.yml`](../.github/workflows/ci.yml) whenever `updatecli/**` changes. The diff job uses `target_branch = ${{ github.base_ref || 'master' }}` and `STACK*_SOURCE_BRANCH = master`, so it shows the diff that *would* be applied if the PR merged today.
