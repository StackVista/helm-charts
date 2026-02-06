#!/usr/bin/env python3
"""
bump_chart_version.py - Bump a Helm chart version and update all dependents

This script:
1. Bumps the version of a specified Helm chart (major, minor, or patch)
2. Finds all charts that depend on it (including transitive dependencies)
3. Performs a patch version bump on all dependent charts
4. Updates dependency versions in Chart.yaml files
5. Regenerates Chart.lock files for all affected charts

Alternatively, with --check mode:
- Validates that all dependent charts have the correct dependency version
- Returns exit code 1 if any dependency version is outdated
- Does not modify any files

Usage:
    ./bump_chart_version.py [OPTIONS] <chart-name>

Examples:
    ./bump_chart_version.py common                    # Patch bump common
    ./bump_chart_version.py -t minor elasticsearch    # Minor bump elasticsearch
    ./bump_chart_version.py -d -v common              # Dry run with verbose output
    ./bump_chart_version.py --skip hbase common       # Bump common but skip hbase version
    ./bump_chart_version.py --check common            # Check if dependents have correct versions
"""

import argparse
import os
import re
import subprocess
import sys
from collections import defaultdict
from pathlib import Path
from typing import Dict, List, Optional, Set, Tuple

import yaml


# Colors for output
class Colors:
    RED = "\033[0;31m"
    GREEN = "\033[0;32m"
    YELLOW = "\033[1;33m"
    BLUE = "\033[0;34m"
    NC = "\033[0m"  # No Color


class BumpChartVersion:
    def __init__(
        self,
        charts_dir: str,
        dry_run: bool = False,
        verbose: bool = False,
        skip_charts: Optional[Set[str]] = None,
    ):
        self.charts_dir = Path(charts_dir)
        self.dry_run = dry_run
        self.verbose = verbose
        self.skip_charts = skip_charts or set()

        # Reverse dependency graph: chart_name -> list of charts that depend on it
        self.reverse_deps: Dict[str, Set[str]] = defaultdict(set)

        # Track bumped chart versions: chart_name -> new_version
        self.bumped_versions: Dict[str, str] = {}

    def log_info(self, message: str) -> None:
        print(f"{Colors.GREEN}[INFO]{Colors.NC} {message}")

    def log_warn(self, message: str) -> None:
        print(f"{Colors.YELLOW}[WARN]{Colors.NC} {message}")

    def log_error(self, message: str) -> None:
        print(f"{Colors.RED}[ERROR]{Colors.NC} {message}", file=sys.stderr)

    def log_debug(self, message: str) -> None:
        if self.verbose:
            print(f"{Colors.BLUE}[DEBUG]{Colors.NC} {message}", file=sys.stderr)

    def get_chart_yaml_path(self, chart_name: str) -> Path:
        """Get the path to a chart's Chart.yaml file."""
        return self.charts_dir / chart_name / "Chart.yaml"

    def read_chart_yaml(self, chart_name: str) -> dict:
        """Read and parse a Chart.yaml file."""
        chart_yaml_path = self.get_chart_yaml_path(chart_name)
        with open(chart_yaml_path, "r") as f:
            return yaml.safe_load(f)

    def write_chart_yaml(self, chart_name: str, data: dict) -> None:
        """Write data to a Chart.yaml file, preserving format as much as possible."""
        chart_yaml_path = self.get_chart_yaml_path(chart_name)
        with open(chart_yaml_path, "w") as f:
            yaml.dump(
                data, f, default_flow_style=False, sort_keys=False, allow_unicode=True
            )

    def get_chart_version(self, chart_name: str) -> str:
        """Get the current version from a chart's Chart.yaml."""
        data = self.read_chart_yaml(chart_name)
        return data.get("version", "")

    def bump_version(self, version: str, bump_type: str) -> str:
        """
        Bump a semantic version string.

        For versions WITH a suffix (e.g., 5.8.0-suse-observability.4):
          - Only the suffix number is bumped: 5.8.0-suse-observability.4 -> 5.8.0-suse-observability.5
          - The base version remains unchanged regardless of bump_type

        For versions WITHOUT a suffix (e.g., 1.2.3):
          - The bump_type determines which part is incremented (major/minor/patch)
        """
        # Parse version: base version and optional suffix
        match = re.match(r"^(\d+)\.(\d+)\.(\d+)(.*)$", version)
        if not match:
            raise ValueError(f"Cannot parse version: {version}")

        major, minor, patch = (
            int(match.group(1)),
            int(match.group(2)),
            int(match.group(3)),
        )
        suffix = match.group(4)

        # For versions with a suffix, only bump the suffix number
        if suffix:
            suffix_match = re.match(r"^(-[a-zA-Z0-9-]+\.)(\d+)$", suffix)
            if suffix_match:
                suffix_prefix = suffix_match.group(1)
                suffix_num = int(suffix_match.group(2))
                new_suffix = f"{suffix_prefix}{suffix_num + 1}"
                return f"{major}.{minor}.{patch}{new_suffix}"
            # Suffix without number - keep as-is
            return f"{major}.{minor}.{patch}{suffix}"

        # For versions without a suffix, perform standard semver bump
        if bump_type == "major":
            major += 1
            minor = 0
            patch = 0
        elif bump_type == "minor":
            minor += 1
            patch = 0
        elif bump_type == "patch":
            patch += 1

        return f"{major}.{minor}.{patch}"

    def update_chart_version(self, chart_name: str, new_version: str) -> None:
        """Update the version in a chart's Chart.yaml file."""
        chart_yaml_path = self.get_chart_yaml_path(chart_name)

        if self.dry_run:
            self.log_info(
                f"[DRY-RUN] Would update {chart_yaml_path} to version {new_version}"
            )
            return

        # Read file content
        with open(chart_yaml_path, "r") as f:
            content = f.read()

        # Replace version line (preserving any quotes)
        content = re.sub(
            r'^(version:\s*)(["\']?)[\d][^\n]*$',
            rf"\g<1>{new_version}",
            content,
            flags=re.MULTILINE,
        )

        with open(chart_yaml_path, "w") as f:
            f.write(content)

        self.log_info(f"Updated {chart_yaml_path} to version {new_version}")

    def update_dependency_version(
        self, chart_name: str, dep_name: str, new_version: str
    ) -> None:
        """Update a dependency's version in a chart's Chart.yaml file."""
        chart_yaml_path = self.get_chart_yaml_path(chart_name)

        if self.dry_run:
            self.log_debug(
                f"[DRY-RUN] Would update dependency {dep_name} to version {new_version} "
                f"in {chart_yaml_path}"
            )
            return

        self.log_debug(
            f"Updating dependency {dep_name} to version {new_version} in {chart_yaml_path}"
        )

        # Read and parse the Chart.yaml
        with open(chart_yaml_path, "r") as f:
            content = f.read()

        # Use regex to find and update the dependency version
        # This handles the YAML structure where name and version may be on different lines
        lines = content.split("\n")
        in_dependencies = False
        current_dep = None
        result_lines = []

        for line in lines:
            # Check if entering dependencies section
            if line.strip() == "dependencies:":
                in_dependencies = True
                result_lines.append(line)
                continue

            # Check if leaving dependencies section (new top-level key)
            if in_dependencies and line and not line[0].isspace():
                in_dependencies = False
                current_dep = None

            if in_dependencies:
                # Check for new dependency item
                name_match = re.match(r'^(\s*-\s*name:\s*)(["\']?)([^"\'#\s]+)', line)
                if name_match:
                    current_dep = name_match.group(3)

                # Check for version line of target dependency
                if current_dep == dep_name:
                    version_match = re.match(
                        r'^(\s*version:\s*)(["\']?)([^"\'#\s]+)(["\']?)(.*)$', line
                    )
                    if version_match:
                        indent = version_match.group(1)
                        quote = version_match.group(2)
                        end_quote = version_match.group(4)
                        rest = version_match.group(5)
                        line = f"{indent}{quote}{new_version}{end_quote}{rest}"

            result_lines.append(line)

        with open(chart_yaml_path, "w") as f:
            f.write("\n".join(result_lines))

    def build_dependency_graph(self) -> None:
        """Build the reverse dependency graph by scanning all Chart.yaml files."""
        self.log_debug("Building dependency graph...")

        for chart_dir in self.charts_dir.iterdir():
            if not chart_dir.is_dir():
                continue

            chart_yaml_path = chart_dir / "Chart.yaml"
            if not chart_yaml_path.exists():
                continue

            chart_name = chart_dir.name

            try:
                with open(chart_yaml_path, "r") as f:
                    data = yaml.safe_load(f)
            except Exception as e:
                self.log_warn(f"Failed to parse {chart_yaml_path}: {e}")
                continue

            dependencies = data.get("dependencies", [])
            if not dependencies:
                continue

            for dep in dependencies:
                dep_name = dep.get("name", "")
                repository = dep.get("repository", "")

                # Only consider local file:// dependencies
                if dep_name and repository.startswith("file://"):
                    self.reverse_deps[dep_name].add(chart_name)
                    self.log_debug(f"  {chart_name} depends on {dep_name} (local)")

        self.log_debug(
            f"Dependency graph built. Found reverse dependencies for {len(self.reverse_deps)} charts."
        )

    def find_all_dependents(self, target_chart: str) -> List[str]:
        """
        Find all charts that depend on the target chart, including transitive dependencies.
        Uses breadth-first search.
        """
        visited: Set[str] = {target_chart}
        dependents: Set[str] = set()
        queue = [target_chart]

        while queue:
            current = queue.pop(0)
            self.log_debug(f"Processing dependencies for: {current}")

            direct_deps = self.reverse_deps.get(current, set())
            for dependent in direct_deps:
                if dependent != target_chart:
                    dependents.add(dependent)

                if dependent not in visited:
                    visited.add(dependent)
                    queue.append(dependent)
                    self.log_debug(f"  Queued for processing: {dependent}")

        return sorted(dependents)

    def update_chart_dependencies(self, chart_name: str) -> None:
        """Update all dependency versions in a chart that reference bumped charts."""
        chart_yaml_path = self.get_chart_yaml_path(chart_name)

        try:
            with open(chart_yaml_path, "r") as f:
                data = yaml.safe_load(f)
        except Exception:
            return

        dependencies = data.get("dependencies", [])
        updated_any = False

        for dep in dependencies:
            dep_name = dep.get("name", "")
            repository = dep.get("repository", "")

            # Only update local dependencies that were bumped
            if dep_name in self.bumped_versions and repository.startswith("file://"):
                new_version = self.bumped_versions[dep_name]
                if self.dry_run:
                    self.log_info(
                        f"[DRY-RUN] Would update dependency {dep_name} to {new_version} in {chart_name}"
                    )
                else:
                    self.log_info(
                        f"Updating dependency {dep_name} to {new_version} in {chart_name}"
                    )
                self.update_dependency_version(chart_name, dep_name, new_version)
                updated_any = True

        if not updated_any:
            self.log_debug(f"No dependency updates needed for {chart_name}")

    def helm_repo_update(self) -> None:
        """Run helm repo update to refresh repository indices."""
        if self.dry_run:
            self.log_info("[DRY-RUN] Would run helm repo update")
            return

        self.log_info("Running helm repo update...")

        try:
            result = subprocess.run(
                ["helm", "repo", "update"],
                capture_output=True,
                text=True,
            )
            if result.returncode != 0:
                self.log_warn(f"Failed to update helm repos: {result.stderr}")
            elif self.verbose:
                for line in result.stdout.strip().split("\n"):
                    if line:
                        self.log_debug(f"  helm: {line}")
        except FileNotFoundError:
            self.log_warn("helm command not found. Skipping helm repo update.")

    def regenerate_chart_lock(self, chart_name: str) -> None:
        """Regenerate Chart.lock for a chart using helm dependency update --skip-refresh."""
        chart_dir = self.charts_dir / chart_name

        if self.dry_run:
            self.log_info(f"[DRY-RUN] Would regenerate Chart.lock for {chart_name}")
            return

        self.log_info(f"Regenerating Chart.lock for {chart_name}")

        try:
            result = subprocess.run(
                ["helm", "dependency", "update", "--skip-refresh", str(chart_dir)],
                capture_output=True,
                text=True,
            )
            if result.returncode != 0:
                self.log_warn(
                    f"Failed to update dependencies for {chart_name}: {result.stderr}"
                )
            elif self.verbose:
                for line in result.stdout.strip().split("\n"):
                    if line:
                        self.log_debug(f"  helm: {line}")
        except FileNotFoundError:
            self.log_warn("helm command not found. Skipping Chart.lock regeneration.")

    def check_dependencies(self, target_chart: str) -> bool:
        """
        Check if all dependent charts have the correct dependency version for the target chart.

        Returns:
            True if all dependency versions are correct, False otherwise.
        """
        self.log_info(f"Checking dependency versions for: {target_chart}")

        # Validate target chart exists
        target_chart_yaml = self.get_chart_yaml_path(target_chart)
        if not target_chart_yaml.exists():
            self.log_error(f"Chart not found: {target_chart}")
            return False

        # Get the target chart's current version
        expected_version = self.get_chart_version(target_chart)
        self.log_info(f"Expected dependency version: {expected_version}")

        # Build dependency graph
        self.build_dependency_graph()

        # Find all charts that depend on the target
        dependents = self.find_all_dependents(target_chart)

        if not dependents:
            self.log_info("No dependent charts found.")
            return True

        self.log_info(f"Checking {len(dependents)} dependent charts...")

        # Flush stdout before checking so error messages appear in order
        sys.stdout.flush()

        all_correct = True
        for dependent in dependents:
            # Read the dependent's Chart.yaml
            dep_chart_yaml = self.get_chart_yaml_path(dependent)
            if not dep_chart_yaml.exists():
                continue

            try:
                with open(dep_chart_yaml, "r") as f:
                    data = yaml.safe_load(f)
            except Exception as e:
                self.log_warn(f"Failed to parse {dep_chart_yaml}: {e}")
                continue

            # Check if this chart has the target as a local dependency
            for dep in data.get("dependencies", []):
                dep_name = dep.get("name", "")
                repository = dep.get("repository", "")
                version = dep.get("version", "")

                if dep_name == target_chart and repository.startswith("file://"):
                    if version != expected_version:
                        self.log_error(
                            f"Chart '{dependent}' has outdated dependency on '{target_chart}': "
                            f"{version} (expected: {expected_version})"
                        )
                        all_correct = False
                    else:
                        self.log_debug(
                            f"Chart '{dependent}' has correct dependency version: {version}"
                        )

        if all_correct:
            self.log_info("All dependency versions are correct.")
        else:
            self.log_error(
                f"Some dependency versions are outdated. "
                f"Run: python3 scripts/bump-chart-version/bump_chart_version.py {target_chart}"
            )

        return all_correct

    def run(self, target_chart: str, bump_type: str = "patch") -> None:
        """Main execution flow."""
        self.log_info(f"Bumping chart: {target_chart} ({bump_type} bump)")
        self.log_info(f"Charts directory: {self.charts_dir}")

        if self.dry_run:
            self.log_warn("DRY-RUN mode enabled - no changes will be made")

        # Run helm repo update once at the start
        self.helm_repo_update()

        # Validate target chart exists
        target_chart_yaml = self.get_chart_yaml_path(target_chart)
        if not target_chart_yaml.exists():
            self.log_error(
                f"Chart not found: {target_chart} (expected at {target_chart_yaml})"
            )
            sys.exit(1)

        # Get current version and calculate new version
        current_version = self.get_chart_version(target_chart)
        new_version = self.bump_version(current_version, bump_type)

        self.log_info(f"Version: {current_version} -> {new_version}")

        # Update the target chart's version and track it
        self.update_chart_version(target_chart, new_version)
        self.bumped_versions[target_chart] = new_version

        # Build the dependency graph
        self.log_info("Building dependency graph...")
        self.build_dependency_graph()

        # Find all dependent charts (including transitive)
        self.log_info("Finding dependent charts...")
        dependents = self.find_all_dependents(target_chart)

        if not dependents:
            self.log_info("No dependent charts found.")
        else:
            self.log_info("Found dependent charts:")
            for dep in dependents:
                print(f"  - {dep}")

            # Process each dependent chart - bump their versions (unless skipped)
            for dependent in dependents:
                dep_chart_yaml = self.get_chart_yaml_path(dependent)

                if not dep_chart_yaml.exists():
                    self.log_warn(f"Dependent chart not found: {dependent}")
                    continue

                # Check if this chart should be skipped
                if dependent in self.skip_charts:
                    self.log_info(f"Skipping version bump for: {dependent} (--skip)")
                    continue

                # Patch bump for dependents (only suffix bump for suffixed versions)
                dep_current_version = self.get_chart_version(dependent)
                dep_new_version = self.bump_version(dep_current_version, "patch")

                self.log_info(
                    f"Processing dependent: {dependent} ({dep_current_version} -> {dep_new_version})"
                )
                self.update_chart_version(dependent, dep_new_version)
                self.bumped_versions[dependent] = dep_new_version

            # Update dependency versions in each dependent chart's Chart.yaml
            self.log_info("Updating dependency versions in Chart.yaml files...")
            for dependent in dependents:
                dep_chart_yaml = self.get_chart_yaml_path(dependent)
                if dep_chart_yaml.exists():
                    self.update_chart_dependencies(dependent)

        # Regenerate Chart.lock files for all affected charts
        self.log_info("Regenerating Chart.lock files...")

        # First, regenerate for the target chart (if it has dependencies)
        target_data = self.read_chart_yaml(target_chart)
        if target_data.get("dependencies"):
            self.regenerate_chart_lock(target_chart)

        # Then regenerate for all dependent charts
        for dependent in dependents:
            dep_chart_dir = self.charts_dir / dependent
            if dep_chart_dir.exists():
                self.regenerate_chart_lock(dependent)

        self.log_info("Done!")

        if self.dry_run:
            self.log_warn("This was a dry run. No changes were made.")


def main():
    # Determine default charts directory
    script_dir = Path(__file__).parent.resolve()
    repo_root = script_dir.parent.parent
    default_charts_dir = repo_root / "stable"

    parser = argparse.ArgumentParser(
        description="Bump a Helm chart version and update all dependent charts.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s common                        # Patch bump common
  %(prog)s -t minor elasticsearch        # Minor bump elasticsearch
  %(prog)s -d -v common                  # Dry run with verbose output
  %(prog)s -c /path/to/charts kafka      # Use custom charts directory
  %(prog)s --skip hbase --skip kafka common  # Skip version bumps for hbase and kafka

The script will:
1. Bump the version of the specified chart
2. Find all charts that depend on it (including transitive dependencies)
3. Perform a patch version bump on all dependent charts (once per chart)
4. Update dependency versions in Chart.yaml files
5. Regenerate Chart.lock files for all affected charts

Charts specified with --skip will have their dependency versions updated
but will NOT have their own version bumped.
        """,
    )

    parser.add_argument(
        "chart_name",
        help="Name of the chart to bump",
    )
    parser.add_argument(
        "-t",
        "--type",
        dest="bump_type",
        choices=["major", "minor", "patch"],
        default="patch",
        help="Version bump type (default: patch)",
    )
    parser.add_argument(
        "-c",
        "--charts-dir",
        dest="charts_dir",
        default=str(default_charts_dir),
        help=f"Directory containing charts (default: {default_charts_dir})",
    )
    parser.add_argument(
        "-d",
        "--dry-run",
        dest="dry_run",
        action="store_true",
        help="Show what would be done without making changes",
    )
    parser.add_argument(
        "-v",
        "--verbose",
        action="store_true",
        help="Enable verbose output",
    )
    parser.add_argument(
        "-s",
        "--skip",
        dest="skip_charts",
        action="append",
        default=[],
        metavar="CHART",
        help="Skip version bump for CHART (can be specified multiple times). "
        "Skipped charts will still have their dependency versions updated.",
    )
    parser.add_argument(
        "--check",
        action="store_true",
        help="Check if dependent charts have correct dependency versions. "
        "Returns exit code 1 if any dependency version is outdated. "
        "Does not modify any files.",
    )

    args = parser.parse_args()

    # Validate charts directory
    if not Path(args.charts_dir).is_dir():
        print(
            f"{Colors.RED}[ERROR]{Colors.NC} Charts directory not found: {args.charts_dir}",
            file=sys.stderr,
        )
        sys.exit(1)

    bumper = BumpChartVersion(
        charts_dir=args.charts_dir,
        dry_run=args.dry_run,
        verbose=args.verbose,
        skip_charts=set(args.skip_charts),
    )

    if args.check:
        # Check mode: validate dependency versions without making changes
        success = bumper.check_dependencies(args.chart_name)
        sys.exit(0 if success else 1)
    else:
        # Normal mode: bump versions
        bumper.run(args.chart_name, args.bump_type)


if __name__ == "__main__":
    main()
