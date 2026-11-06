# Updatecli chart adoption

The active workflow is [`.github/workflows/updatecli.yml`](.github/workflows/updatecli.yml).
It evaluates and applies the manifests under `updatecli/updatecli.d/`.
The public chart repository is a mirror, not a second updater or editing target.

## CVE ownership

Producer fixers publish ready normal components, including agent images, under standing
approval regardless of registry visibility. Source merged through normal repository
protections, including approved bot automerge, qualifies. Review the full release range
to know what ships, not to add approval for every change; preserve immutable releases.
Producers must not write chart pins, open adoption PRs or dispatch chart-mutating
workflows. Read-only inspection and disposable renders with external values overrides
are allowed.

The CVE planner coordinates adoption through Updatecli, reusing existing automatic
work and batching compatible ready releases. A planner may explicitly assign a
repair for a demonstrated updater/template defect and its necessary tests/docs.
That exception does not permit routine image bumps or candidate-pin restoration.
Final platform and agent-chart releases are outside CVE remediation execution; hand them
to the existing release owner and runbooks. Here, agent release means publication of the
suse-observability-agent chart, not its component images. Human PR review covering the
actual release changes is the approval record; do not request a second confirmation. A
producer PR review alone does not approve those release changes or authorize
deployments, infrastructure changes or other runbook side effects.

## Evaluation and application

A push changing `updatecli/**` or the workflow triggers configuration validation
and `updatecli pipeline diff`. The workflow evaluates the pushed branch; ensure any
referenced scripts exist in that checkout. An actual diff resolves sources and
executes the manifest's shell resources, so inspect those commands before running it.

Manual dispatch chooses a pipeline and target branch:

| Pipeline | Manifest | Apply follow-up |
| --- | --- | --- |
| `docker_images` | `update-docker-images/update.yaml` | Force-reset updater branch, add a signed squash commit, then find/open a PR. Re-dispatch can close/recreate an existing PR and invalidate review/CI. |
| `stackstate` | `update-stackstate/update.yaml` | Squash-push to the selected target branch. |
| `stackpacks` | `update-stackpacks/update.yaml` | Squash-push to the selected target branch. |

**A dispatch with `dry_run=true` skips both apply and evaluation.** Its summary
records inputs; green status does not prove that the selected pipeline was tested.
For repair branches changing `updatecli/**`, use the push-validation job. For
adoption without a source change, use a real local pipeline diff:

```bash
export TARGET_BRANCH=master
updatecli pipeline diff \
  --config updatecli/updatecli.d/update-docker-images/ \
  --values updatecli/values.d/values.yaml
```

Supply credentials through the existing environment, never as literal command
arguments. Select the branch actually under review rather than assuming master
contains a proposed resolver. Match the local checkout and target-branch revision:
shell resources can use either working directory. A diff can create local SCM
checkouts; it is not a zero-side-effect evaluation of arbitrary unreviewed manifests.
Diff may omit generated README, quoted-tag and version changes; inspect the full resulting PR diff before
merge or handoff.

An applying dispatch uses `dry_run=false`. Inspect its complete proposed changes
and workflow side effects first. Use each pipeline's existing supported apply path.
StackState/StackPack workflows write directly to their selected target; verify that
target and the complete normal update before dispatch. Standing adoption approval
covers those workflow writes, not manual default-branch pushes or unrelated App
bypass. Existing producer-triggered automation continues normally.

The PR helper alone reuses an open head/base PR. The preceding squash step force-
resets the updater branch to the target tip before adding the new commit; GitHub
can close the old PR and discard its approval/auto-merge state. Reuse an active run
or open candidate by following it, not re-dispatching. Record the dispatch identity
and candidate link in the planner notebook. Only a deliberate correction should
replace an open candidate, with fresh CI and review; late routine arrivals wait.
Existing producer-triggered automation continues; do not race it with manual apply.

The planner has standing approval to run existing Updatecli adoption workflows and merge
their verified normal-update PRs, including normal publication side effects. No
additional human approval is needed for that adoption.
Satisfy required branch reviews through the repository's existing reviewer path; never
self-approve, fabricate approval or weaken protection. If that review is absent, request
or wait for it. This is the existing merge prerequisite, not a new authorization
question for Updatecli adoption. Review the complete diff,
selected artifacts and required CI before merge; honor scoped holds and preserve
candidates already under review. Updater/template repair-code merges retain their source approval boundary. Final
platform and agent-chart releases remain with the existing release owner.

The image pipeline evaluates all tracked images; it has no CVE-only apply selector. The
planner may include other verified, compatible normal-channel updates in the same CVE
batch. Inspect every selected artifact before applying and review every resulting change
before merge. Unverified, incompatible or off-channel changes remain preview-only until
resolved. Do not copy selected pins by hand or describe the full pipeline as a targeted
apply.

## Release selection and diagnosis

The manifests and [image mapping](updatecli/updatecli-image-mapping.md) define
supported producer channels and actual chart fields. Sources include registry
lookups and shell resolvers. Normal snapshots, rebuild suffixes and commit-based
releases are valid where the producer contract supports them. Temporary candidate
images are qualification artifacts and must not become chart defaults.

Before applying, compare intended releases with the selected source results and
all changed targets, including unrelated images in the same pipeline.

| Observation | Action |
| --- | --- |
| The normal release is absent or the producer used an unsupported name | Return publication/naming work to the producer; do not hand-pin its candidate or broaden a filter to admit it. |
| A valid supported release exists but filtering, ordering or target mapping misses it | Prepare a scoped updater repair with a reproducer and regression checks. |
| Lookup/authentication or publication completion is uncertain | Establish the failing boundary before choosing a producer or updater repair. |
| The intended artifact is already adopted | Verify identity and delivery; do not create duplicate changes. |
| A different selected tag has findings | Compare actual source, packages, digest, scanner evidence and VEX applicability; a changed tag alone does not prove the source fix was lost. |

`latest-so-tag.sh` compares numeric upstream components and release increments.
The agent shell sources check that the images for their selected producer commit
exist; when unavailable, they retain the current pin. Such a successful no-change
result does not prove a new artifact was adopted or all producer checks passed.

Keep a candidate stable during full CI and independent review. Late routine
releases belong in the next batch; blocked producers need not delay ready fixes.
If the complete diff is unsuitable, resolve its scope before apply. Do not assume
there is a selective mode or manually copy a desired pin around the updater.

## Targets, documentation and versions

Each YAML target names its source with `sourceid` and its chart field with
`spec.file` plus `spec.key` or `spec.keys`. Use `dependson` for real dependencies,
not an artificial chain through every image.

The Docker-image manifest contains the final shell targets: normalize quoted tags,
regenerate platform documentation and bump the agent chart when needed. There is
no separate finalization pipeline. The platform publisher owns its prerelease;
ordinary adoption must not allocate a speculative platform version on a branch.
The [version helper](scripts/bump-chart-version/README.md) changes one selected
stable chart and regenerates docs; it does not cascade versions through dependents
or regenerate lockfiles. Local dependencies use wildcard versions.

For an assigned updater repair, inspect source/target wiring, resolver tests and a
real pipeline diff against the candidate. Validate required chart checks on the
actual head. Render/image-reference checks do not replace artifact scan evidence.
After chart publication, use a complete VEX-aware delivery scan for CVE closure.

## Environment and helpers

`updatecli/values.d/values.yaml` defines environment-variable mappings. Workflow
credentials and signing are configured by `scripts/ci/configure_git.sh`.

| Variable | Purpose |
| --- | --- |
| `TARGET_BRANCH` | Intended chart SCM branch. |
| `HELM_CHARTS_PAT` | GitHub SCM/API credential supplied through the environment. |
| `UPDATE_CLI_USER`, `UPDATE_CLI_EMAIL` | Updater commit identity. |
| `REGISTRY_USER`, `REGISTRY_PASSWORD` | Registry credentials where required. |
| `STACKSTATE_SOURCE_BRANCH` | Producer branch for the StackState pipeline. |
| `STACKPACKS_SOURCE_BRANCH`, `STACKPACKS_CONTRIB_SOURCE_BRANCH`, `STACKPACKS_SUSE_SOURCE_BRANCH` | Producer branches for the StackPack pipeline. |

The workflow's `apply` and `follow_up` jobs supply their own GitHub App credentials.
`scripts/ci/updatecli_squash_push.sh` handles signed final commits;
`scripts/ci/open_updatecli_pr.sh` reuses or opens the Docker-image adoption PR.
