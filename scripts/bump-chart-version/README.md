# bump-chart-version

A Python script that bumps the version of a specified Helm chart and automatically updates all dependent charts, including transitive dependencies.

## Features

1. **Bumps the version of a specified Helm chart** - supports major, minor, or patch bumps (default: patch)
2. **Finds all dependent charts** - including transitive dependencies (e.g., if A depends on B, and B depends on C, bumping C will also bump B and A)
3. **Performs a single patch bump for dependents** - even if a chart has multiple updated dependencies, it only gets bumped once
4. **Updates dependency versions in Chart.yaml** - automatically updates the `version:` field in the `dependencies:` section to reference the new versions of bumped charts
5. **Regenerates Chart.lock files** - uses `helm dependency update` to update lock files for all affected charts
6. **Handles complex version formats** - supports versions like `1.2.3`, `8.19.4-stackstate.5`, `2.7.1-pre.54`, etc.

## Usage

```bash
# Basic usage (patch bump) - run from repository root
./scripts/bump-chart-version/bump_chart_version.py common

# Minor bump
./scripts/bump-chart-version/bump_chart_version.py -t minor elasticsearch

# Major bump
./scripts/bump-chart-version/bump_chart_version.py -t major kafka

# Dry run (see what would change without making changes)
./scripts/bump-chart-version/bump_chart_version.py -d common

# Verbose output (shows dependency graph building)
./scripts/bump-chart-version/bump_chart_version.py -d -v common

# Custom charts directory
./scripts/bump-chart-version/bump_chart_version.py -c /path/to/charts kafka

# Skip version bumps for specific charts (they still get dependency updates)
./scripts/bump-chart-version/bump_chart_version.py --skip hbase --skip kafka common
```

## Options

| Option | Description |
|--------|-------------|
| `-t, --type TYPE` | Version bump type: `major`, `minor`, `patch` (default: `patch`) |
| `-c, --charts-dir DIR` | Directory containing charts (default: `./stable`) |
| `-d, --dry-run` | Show what would be done without making changes |
| `-v, --verbose` | Enable verbose output |
| `-s, --skip CHART` | Skip version bump for CHART (can be specified multiple times). Skipped charts will still have their dependency versions updated. |
| `-h, --help` | Show help message |

## Key Design Decisions

- **Only considers local `file://` dependencies** - Remote dependencies (like from `helm.stackstate.io`) are not updated since they reference external charts that are versioned independently.

- **Builds dependency graph upfront** - The script performs a single-pass parsing of all `Chart.yaml` files to build a reverse dependency graph, making transitive dependency resolution efficient.

- **Safe by default** - Supports dry-run mode (`-d`) to preview all changes before making them. Always recommended to run with `-d` first.

- **Version suffix handling** - For versions with suffixes like `-stackstate.5` or `-suse-observability.7`, only the suffix number is incremented (e.g., `5.8.0-suse-observability.4` → `5.8.0-suse-observability.5`). The base version remains unchanged. For versions without a suffix, the standard semver bump applies.

- **Optimized helm operations** - The script runs `helm repo update` once at the start, then uses `--skip-refresh` on all subsequent `helm dependency update` commands. This significantly speeds up execution when updating many charts.

## Examples

### Example 1: Bumping the `common` chart

The `common` chart is a library chart used by many other charts. Running:

```bash
./scripts/bump-chart-version/bump_chart_version.py -d common
```

Would show output like:

```
[INFO] Bumping chart: common (patch bump)
[INFO] Version: 0.4.27 -> 0.4.28
[INFO] Found dependent charts:
  - clickhouse
  - hbase
  - kafka
  - minio
  - suse-observability
  - victoria-metrics-single
  ...
[INFO] Processing dependent: clickhouse (3.6.9-suse-observability.8 -> 3.6.9-suse-observability.9)
...
[INFO] Updating dependency versions in Chart.yaml files...
[INFO] Updating dependency common to 0.4.28 in clickhouse
...
[INFO] Regenerating Chart.lock files...
```

Note: For the dependent chart `clickhouse`, only the suffix number is incremented (`8` → `9`), while the base version (`3.6.9`) remains unchanged.

### Example 2: Transitive dependencies

If `prometheus-elasticsearch-exporter` is bumped:

```bash
./scripts/bump-chart-version/bump_chart_version.py -d prometheus-elasticsearch-exporter
```

The script will:
1. Bump `prometheus-elasticsearch-exporter`
2. Find that `elasticsearch` depends on it → bump `elasticsearch`
3. Find that `suse-observability` depends on `elasticsearch` → bump `suse-observability`
4. Update the dependency version references in each dependent chart's `Chart.yaml`
5. Regenerate Chart.lock files for all affected charts

### Example 3: Skipping specific charts

Sometimes you may want to bump a chart and its dependents, but skip the version bump for specific charts (e.g., charts that are managed separately or have manual release processes). Use the `--skip` option:

```bash
./scripts/bump-chart-version/bump_chart_version.py -d --skip suse-observability --skip hbase common
```

This will:
1. Bump `common` as usual
2. Find all charts that depend on `common`
3. **Skip version bumps** for `suse-observability` and `hbase`
4. **Still update** the dependency versions in `suse-observability` and `hbase` (so they reference the new `common` version)
5. Bump all other dependent charts normally

Output:
```
[INFO] Bumping chart: common (patch bump)
[INFO] Version: 0.4.27 -> 0.4.28
[INFO] Found dependent charts:
  - clickhouse
  - hbase
  - kafka
  - suse-observability
  ...
[INFO] Processing dependent: clickhouse (3.6.9-suse-observability.8 -> 3.6.9-suse-observability.9)
[INFO] Skipping version bump for: hbase (--skip)
[INFO] Skipping version bump for: suse-observability (--skip)
...
[INFO] Updating dependency versions in Chart.yaml files...
[INFO] Updating dependency common to 0.4.28 in hbase
[INFO] Updating dependency common to 0.4.28 in suse-observability
...
```

## Requirements

- Python 3.8+
- PyYAML (`pip install pyyaml`)
- `helm` CLI (for `helm dependency update`)

## Running Tests

The script includes a comprehensive test suite using pytest.

```bash
# Install test dependencies
pip install pytest pyyaml

# Run all tests (from repository root)
pytest scripts/bump-chart-version/test_bump_chart_version.py -v
```

The test suite covers:
- Version bumping (patch, minor, major)
- Version suffix handling (e.g., `-stackstate.5`, `-pre.54`)
- Direct dependency detection
- Transitive dependency detection
- File modification verification
- Dependency version updates in Chart.yaml
- Error handling
- Chart.yaml parsing edge cases (quoted names, aliases, conditions)
- Multiple dependencies in a single chart (the bug that affected the bash version)

## Migration from Bash Script

The Python version is a complete rewrite that:
- Fixes the `BASH_REMATCH` bug that caused some dependencies to be missed
- Uses proper YAML parsing instead of regex-based parsing
- Has better error handling and more informative error messages
- Is easier to maintain and extend

The command-line interface is identical to the bash version, so you can simply replace:
```bash
./bump-chart-version.sh [OPTIONS] <chart-name>
```
with:
```bash
./bump_chart_version.py [OPTIONS] <chart-name>
```
