# updatecli

Pipelines and shared values for the three updatecli flows in this repo:

- `updatecli.d/update-docker-images/` — bumps the `quay.io/stackstate/*` image tags across the SUSE Observability charts and opens an MR.
- `updatecli.d/update-stackstate/` — bumps the StackState image tag, the chart `appVersion`, and the StackGraph version. Commits directly to the target branch (no MR).
- `updatecli.d/update-stackpacks/` — bumps the StackPacks image version. Commits directly to the target branch (no MR).

`values.d/values.yaml` holds the shared SCM/source configuration that all three pipelines template into their yaml.

For pipeline architecture, tag-format conventions, and how to add or remove images, see [`../UPDATECLI.md`](../UPDATECLI.md).

## Testing changes in CI

The safe way to validate a change to this directory is to **trigger a manual pipeline on a test branch** and point updatecli at that branch instead of `master`. None of the production data ends up on master unless you set `TARGET_BRANCH=master` explicitly.

Steps:

1. Push your changes to a branch (e.g. `stac-1234-updatecli-tweak`).
2. In GitLab: **CI/CD → Pipelines → Run pipeline**, select your branch.
3. Add the variables below.
4. Run. Inspect the resulting commits / MR on `TARGET_BRANCH`.

### Variables

The pipeline only runs an updatecli job when its trigger variable is set. Set exactly one per test run:

| Variable | When to set it |
|---|---|
| `RUN_UPDATECLI` | Trigger the docker-images pipeline (creates `updatecli-${TARGET_BRANCH}-docker-images` working branch + MR). |
| `RUN_UPDATECLI_STACKSTATE` | Trigger the stackstate + stackgraph pipeline (commits directly to `TARGET_BRANCH`). |
| `RUN_UPDATECLI_STACKPACKS` | Trigger the stackpacks pipeline (commits directly to `TARGET_BRANCH`). |

Always set the branch variables — there are no defaults, the jobs fail fast if these are missing:

| Variable | Meaning |
|---|---|
| `TARGET_BRANCH` | Branch in `devops/helm-charts` that updatecli clones, commits to, and (for docker-images) opens the MR against. **Use a test branch when validating changes.** |
| `STACKSTATE_SOURCE_BRANCH` | Branch segment in `stackstate-receiver` Docker tags (e.g. `master`, `stac-1234-feature`). Also used to fetch `stackgraph_version` from that branch of `stackvista/stackstate`. Only relevant for the stackstate pipeline. |
| `STACKPACKS_SOURCE_BRANCH` | Branch segment in `stackpacks` Docker tags. Only relevant for the stackpacks pipeline. |

### Example: validate the stackstate pipeline against a feature branch

```
RUN_UPDATECLI_STACKSTATE = true
TARGET_BRANCH = stac-1234-updatecli-tweak
STACKSTATE_SOURCE_BRANCH = master
```

The job will clone `stac-1234-updatecli-tweak`, fetch the latest master `stackstate-receiver` tag + `stackgraph_version`, apply the value updates, and push a single squashed commit back to `stac-1234-updatecli-tweak`. Inspect the commit; if it looks wrong, fix the pipeline yaml and re-run.

### MR-pipeline validation

When you push an updatecli config change in an MR, `validate_updatecli_config` (yamllint) and `validate_updatecli_diff` (`updatecli diff` against all three pipelines) run automatically. The diff job uses `TARGET_BRANCH = $CI_MERGE_REQUEST_TARGET_BRANCH_NAME` and `STACK*_SOURCE_BRANCH = master`, so it shows the diff that *would* be applied if the MR merged today.
