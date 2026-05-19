#!/usr/bin/env python3
"""
bump_chart_version.py - Bump a Helm chart's own version

Bumps the version of a single Helm chart's Chart.yaml and regenerates its
README.md via helm-docs. Used by the updatecli pipeline to bump the
suse-observability-agent chart after docker image tag changes.

Local-only charts under local/ are pinned to "*" by their consumers, so they
do not need version bumps and are not supported by this script.

Usage:
    ./bump_chart_version.py [OPTIONS] <chart-name>

Examples:
    ./bump_chart_version.py suse-observability-agent       # Patch bump
    ./bump_chart_version.py -t minor suse-observability    # Minor bump
    ./bump_chart_version.py -d suse-observability-agent    # Dry run
"""

import argparse
import re
import subprocess
import sys
from pathlib import Path


class Colors:
    RED = "\033[0;31m"
    GREEN = "\033[0;32m"
    YELLOW = "\033[1;33m"
    NC = "\033[0m"


def log_info(message: str) -> None:
    print(f"{Colors.GREEN}[INFO]{Colors.NC} {message}")


def log_warn(message: str) -> None:
    print(f"{Colors.YELLOW}[WARN]{Colors.NC} {message}")


def log_error(message: str) -> None:
    print(f"{Colors.RED}[ERROR]{Colors.NC} {message}", file=sys.stderr)


def bump_version(version: str, bump_type: str) -> str:
    """
    Bump a semantic version string.

    For versions WITH a suffix (e.g., 5.8.0-suse-observability.4):
      - Only the suffix number is bumped: ...4 -> ...5
      - The base version is unchanged regardless of bump_type

    For versions WITHOUT a suffix (e.g., 1.2.3):
      - bump_type determines which part is incremented (major/minor/patch)
    """
    match = re.match(r"^(\d+)\.(\d+)\.(\d+)(.*)$", version)
    if not match:
        raise ValueError(f"Cannot parse version: {version}")

    major, minor, patch = int(match.group(1)), int(match.group(2)), int(match.group(3))
    suffix = match.group(4)

    if suffix:
        suffix_match = re.match(r"^(-[a-zA-Z0-9-]+\.)(\d+)$", suffix)
        if suffix_match:
            suffix_num = int(suffix_match.group(2))
            return f"{major}.{minor}.{patch}{suffix_match.group(1)}{suffix_num + 1}"
        return f"{major}.{minor}.{patch}{suffix}"

    if bump_type == "major":
        return f"{major + 1}.0.0"
    if bump_type == "minor":
        return f"{major}.{minor + 1}.0"
    return f"{major}.{minor}.{patch + 1}"


def update_chart_version(chart_yaml_path: Path, new_version: str, dry_run: bool) -> None:
    if dry_run:
        log_info(f"[DRY-RUN] Would update {chart_yaml_path} to version {new_version}")
        return

    with open(chart_yaml_path, "r") as f:
        content = f.read()

    content = re.sub(
        r'^(version:\s*)(["\']?)[\d][^\n]*$',
        rf"\g<1>{new_version}",
        content,
        flags=re.MULTILINE,
    )

    with open(chart_yaml_path, "w") as f:
        f.write(content)

    log_info(f"Updated {chart_yaml_path} to version {new_version}")


def run_helm_docs(chart_dir: Path, dry_run: bool) -> None:
    if dry_run:
        log_info(f"[DRY-RUN] Would run helm-docs for {chart_dir.name}")
        return

    log_info(f"Running helm-docs for {chart_dir.name}")
    try:
        result = subprocess.run(
            ["helm-docs", "--chart-search-root", str(chart_dir)],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            log_warn(f"helm-docs failed for {chart_dir.name}: {result.stderr}")
    except FileNotFoundError:
        log_warn("helm-docs command not found. Skipping README regeneration.")


def main():
    script_dir = Path(__file__).parent.resolve()
    repo_root = script_dir.parent.parent
    default_charts_dir = repo_root / "stable"

    parser = argparse.ArgumentParser(
        description="Bump a Helm chart's own version.",
    )
    parser.add_argument("chart_name", help="Name of the chart to bump")
    parser.add_argument(
        "-t", "--type",
        dest="bump_type",
        choices=["major", "minor", "patch"],
        default="patch",
        help="Version bump type (default: patch)",
    )
    parser.add_argument(
        "-c", "--charts-dir",
        dest="charts_dir",
        default=str(default_charts_dir),
        help=f"Directory containing charts (default: {default_charts_dir})",
    )
    parser.add_argument(
        "-d", "--dry-run",
        action="store_true",
        help="Show what would be done without making changes",
    )
    args = parser.parse_args()

    charts_dir = Path(args.charts_dir)
    if not charts_dir.is_dir():
        log_error(f"Charts directory not found: {charts_dir}")
        sys.exit(1)

    chart_dir = charts_dir / args.chart_name
    chart_yaml = chart_dir / "Chart.yaml"
    if not chart_yaml.is_file():
        log_error(f"Chart not found: {chart_yaml}")
        sys.exit(1)

    with open(chart_yaml, "r") as f:
        for line in f:
            m = re.match(r'^version:\s*["\']?([^"\'#\s]+)', line)
            if m:
                current_version = m.group(1)
                break
        else:
            log_error(f"No version field found in {chart_yaml}")
            sys.exit(1)

    new_version = bump_version(current_version, args.bump_type)
    log_info(f"Bumping {args.chart_name} {current_version} -> {new_version}")

    update_chart_version(chart_yaml, new_version, args.dry_run)
    run_helm_docs(chart_dir, args.dry_run)


if __name__ == "__main__":
    main()
