# bump-chart-version

Bumps the `version:` field in a single Helm chart's `Chart.yaml` and regenerates its `README.md` via `helm-docs`.

Used by the updatecli pipeline to bump `suse-observability-agent` after docker image tag changes.

Local-only charts under `local/` are pinned to `"*"` by their consumers and do not need version bumps — this script does not operate on them.

## Usage

```bash
# Patch bump (default)
./scripts/bump-chart-version/bump_chart_version.py suse-observability-agent

# Minor bump
./scripts/bump-chart-version/bump_chart_version.py -t minor suse-observability

# Dry run
./scripts/bump-chart-version/bump_chart_version.py -d suse-observability-agent
```

## Options

| Option | Description |
|--------|-------------|
| `-t, --type TYPE` | Version bump type: `major`, `minor`, `patch` (default: `patch`) |
| `-c, --charts-dir DIR` | Directory containing charts (default: `./stable`) |
| `-d, --dry-run` | Show what would be done without making changes |
| `-h, --help` | Show help message |

## Version handling

- Versions with a numbered suffix (e.g. `5.8.0-suse-observability.4`) increment only the suffix number, regardless of `--type`.
- Plain semver versions (e.g. `1.2.3`) honor `--type`.
