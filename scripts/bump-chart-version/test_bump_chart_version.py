#!/usr/bin/env python3
"""
Unit tests for bump_chart_version.py

Usage:
    pytest test_bump_chart_version.py -v
    # or
    python -m pytest test_bump_chart_version.py -v
"""

import os
import shutil
import tempfile
from pathlib import Path
from typing import Dict, List, Optional

import pytest
import yaml

from bump_chart_version import BumpChartVersion


class ChartFixtures:
    """Helper class to create test chart fixtures."""

    def __init__(self, base_dir: Path):
        self.base_dir = base_dir
        self.charts_dir = base_dir / "stable"
        self.charts_dir.mkdir(parents=True, exist_ok=True)

    def create_chart(
        self,
        name: str,
        version: str,
        local_deps: Optional[List[str]] = None,
        remote_deps: Optional[List[str]] = None,
        extra_dep_fields: Optional[Dict[str, Dict]] = None,
    ) -> Path:
        """
        Create a test chart with optional dependencies.

        Args:
            name: Chart name
            version: Chart version
            local_deps: List of local dependency names (file:// repository)
            remote_deps: List of remote dependency names (https:// repository)
            extra_dep_fields: Dict mapping dep name to extra fields (e.g., tags, condition)
        """
        chart_dir = self.charts_dir / name
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_data = {
            "apiVersion": "v2",
            "name": name,
            "version": version,
            "description": f"Test chart {name}",
        }

        dependencies = []

        if local_deps:
            for dep in local_deps:
                dep_entry = {
                    "name": dep,
                    "repository": f"file://../{dep}/",
                    "version": "*",
                }
                if extra_dep_fields and dep in extra_dep_fields:
                    dep_entry.update(extra_dep_fields[dep])
                dependencies.append(dep_entry)

        if remote_deps:
            for dep in remote_deps:
                dep_entry = {
                    "name": dep,
                    "repository": "https://helm.example.io",
                    "version": "1.0.0",
                }
                dependencies.append(dep_entry)

        if dependencies:
            chart_data["dependencies"] = dependencies

        chart_yaml_path = chart_dir / "Chart.yaml"
        with open(chart_yaml_path, "w") as f:
            yaml.dump(chart_data, f, default_flow_style=False, sort_keys=False)

        return chart_dir

    def get_version(self, chart_name: str) -> str:
        """Get the version from a chart's Chart.yaml."""
        chart_yaml_path = self.charts_dir / chart_name / "Chart.yaml"
        with open(chart_yaml_path, "r") as f:
            data = yaml.safe_load(f)
        return data.get("version", "")

    def get_dependency_version(self, chart_name: str, dep_name: str) -> Optional[str]:
        """Get a specific dependency's version from a chart's Chart.yaml."""
        chart_yaml_path = self.charts_dir / chart_name / "Chart.yaml"
        with open(chart_yaml_path, "r") as f:
            data = yaml.safe_load(f)

        for dep in data.get("dependencies", []):
            if dep.get("name") == dep_name:
                return dep.get("version")
        return None


@pytest.fixture
def temp_charts_dir():
    """Create a temporary directory for test charts."""
    temp_dir = tempfile.mkdtemp(prefix="helm-charts-test-")
    yield Path(temp_dir)
    shutil.rmtree(temp_dir)


@pytest.fixture
def fixtures(temp_charts_dir):
    """Create test fixtures with common chart structures."""
    f = ChartFixtures(temp_charts_dir)

    # Create test charts
    f.create_chart("common", "0.4.27")
    f.create_chart("library-a", "1.2.3")
    f.create_chart("library-b", "2.0.0-stackstate.5")
    f.create_chart("app-simple", "1.0.0", local_deps=["common"])
    f.create_chart(
        "app-with-suffix", "3.6.9-suse-observability.8", local_deps=["common"]
    )
    f.create_chart("app-multi-dep", "5.0.0", local_deps=["common", "library-a"])
    f.create_chart("app-transitive", "1.0.0", local_deps=["app-simple"])
    f.create_chart("app-deep-transitive", "2.0.0", local_deps=["app-transitive"])
    f.create_chart("app-no-deps", "1.0.0")
    f.create_chart("app-remote-dep", "1.0.0", remote_deps=["remote-chart"])

    return f


class TestVersionBumping:
    """Test version bumping logic."""

    def test_patch_bump_simple_version(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("0.4.27", "patch")
        assert result == "0.4.28"

    def test_minor_bump_simple_version(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("0.4.27", "minor")
        assert result == "0.5.0"

    def test_major_bump_simple_version(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("0.4.27", "major")
        assert result == "1.0.0"

    def test_patch_bump_with_suffix(self, fixtures):
        """For suffixed versions, only the suffix number is bumped."""
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("2.0.0-stackstate.5", "patch")
        assert result == "2.0.0-stackstate.6"

    def test_minor_bump_with_suffix(self, fixtures):
        """For suffixed versions, only the suffix number is bumped regardless of bump type."""
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("3.6.9-suse-observability.8", "minor")
        assert result == "3.6.9-suse-observability.9"

    def test_version_with_pre_suffix(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("2.7.1-pre.54", "patch")
        assert result == "2.7.1-pre.55"

    def test_version_with_snapshot_suffix(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("5.2.0-snapshot.177", "patch")
        assert result == "5.2.0-snapshot.178"

    def test_version_with_complex_suffix(self, fixtures):
        """Even with -t minor, suffixed versions only bump the suffix number."""
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        result = bumper.bump_version("8.19.4-stackstate.5", "minor")
        assert result == "8.19.4-stackstate.6"


class TestDependencyDetection:
    """Test dependency graph building and detection."""

    def test_finds_direct_dependents(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("common")

        assert "app-simple" in dependents
        assert "app-with-suffix" in dependents
        assert "app-multi-dep" in dependents

    def test_finds_transitive_dependents(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("common")

        # app-transitive depends on app-simple, which depends on common
        assert "app-transitive" in dependents
        # app-deep-transitive depends on app-transitive
        assert "app-deep-transitive" in dependents

    def test_ignores_remote_dependencies(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("app-no-deps")

        # app-no-deps has no local dependents
        assert dependents == []

    def test_no_dependents_for_leaf_chart(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("app-deep-transitive")

        assert dependents == []


class TestFileModifications:
    """Test actual file modifications."""

    def test_updates_target_chart_version(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.build_dependency_graph()

        bumper.update_chart_version("common", "0.4.28")

        assert fixtures.get_version("common") == "0.4.28"

    def test_updates_dependent_chart_versions(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("common", "patch")

        # Dependents should get a patch bump (but suffixed versions only bump suffix)
        assert fixtures.get_version("app-simple") == "1.0.1"

    def test_updates_transitive_dependent_versions(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("common", "patch")

        # Transitive dependents should also get bumped
        assert fixtures.get_version("app-transitive") == "1.0.1"

    def test_single_bump_for_multi_dependency_chart(self, fixtures):
        """Chart with multiple updated dependencies should only be bumped once."""
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("common", "patch")

        assert fixtures.get_version("app-multi-dep") == "5.0.1"

    def test_dry_run_does_not_modify_files(self, fixtures):
        original_version = fixtures.get_version("common")

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.run("common", "patch")

        assert fixtures.get_version("common") == original_version


class TestErrorHandling:
    """Test error handling."""

    def test_error_on_nonexistent_chart(self, fixtures, capsys):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)

        with pytest.raises(SystemExit) as exc_info:
            bumper.run("nonexistent-chart", "patch")

        assert exc_info.value.code == 1
        captured = capsys.readouterr()
        assert "Chart not found" in captured.err or "Chart not found" in captured.out

    def test_error_on_invalid_version(self, fixtures):
        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)

        with pytest.raises(ValueError, match="Cannot parse version"):
            bumper.bump_version("invalid-version", "patch")


class TestChartYamlParsing:
    """Test Chart.yaml parsing edge cases."""

    def test_handles_quoted_dependency_names(self, fixtures):
        # Create a chart with quoted dependency name (manually write YAML)
        chart_dir = fixtures.charts_dir / "quoted-deps"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: quoted-deps
version: 1.0.0
dependencies:
  - name: "common"
    repository: file://../common/
    version: "*"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("common")
        assert "quoted-deps" in dependents

    def test_handles_single_quoted_dependency_names(self, fixtures):
        chart_dir = fixtures.charts_dir / "single-quoted-deps"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: single-quoted-deps
version: 1.0.0
dependencies:
  - name: 'common'
    repository: file://../common/
    version: "*"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("common")
        assert "single-quoted-deps" in dependents

    def test_handles_dependency_with_alias(self, fixtures):
        fixtures.create_chart(
            "aliased-dep",
            "1.0.0",
            local_deps=["common"],
            extra_dep_fields={"common": {"alias": "my-common"}},
        )

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("common")
        assert "aliased-dep" in dependents

    def test_handles_dependency_with_condition(self, fixtures):
        fixtures.create_chart(
            "conditional-dep",
            "1.0.0",
            local_deps=["common"],
            extra_dep_fields={"common": {"condition": "common.enabled"}},
        )

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        dependents = bumper.find_all_dependents("common")
        assert "conditional-dep" in dependents


class TestDependencyVersionUpdates:
    """Test dependency version updates in Chart.yaml files."""

    def test_updates_dependency_version_in_chart_yaml(self, fixtures):
        # Create a chart with specific version number in dependency
        chart_dir = fixtures.charts_dir / "dep-consumer"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: dep-consumer
version: 1.0.0
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("common", "patch")

        dep_version = fixtures.get_dependency_version("dep-consumer", "common")
        assert dep_version == "0.4.28"

    def test_updates_multiple_dependencies_in_same_chart(self, fixtures):
        # Create a chart with multiple local dependencies
        chart_dir = fixtures.charts_dir / "multi-local-deps"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: multi-local-deps
version: 1.0.0
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
  - name: library-a
    repository: file://../library-a/
    version: "1.2.3"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("common", "patch")

        # common dependency should be updated
        assert fixtures.get_dependency_version("multi-local-deps", "common") == "0.4.28"
        # library-a dependency should NOT be updated (we only bumped common)
        assert (
            fixtures.get_dependency_version("multi-local-deps", "library-a") == "1.2.3"
        )

    def test_updates_transitive_dependency_versions(self, fixtures):
        # Set up specific versions in the transitive chain
        # app-transitive -> app-simple -> common

        # Update app-transitive to have a specific version for app-simple
        chart_dir = fixtures.charts_dir / "app-transitive"
        chart_yaml_content = """apiVersion: v2
name: app-transitive
version: 1.0.0
dependencies:
  - name: app-simple
    repository: file://../app-simple/
    version: "1.0.0"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        # Update app-simple to have a specific version for common
        chart_dir = fixtures.charts_dir / "app-simple"
        chart_yaml_content = """apiVersion: v2
name: app-simple
version: 1.0.0
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("common", "patch")

        # app-simple's dependency on common should be updated
        assert fixtures.get_dependency_version("app-simple", "common") == "0.4.28"
        # app-transitive's dependency on app-simple should be updated (app-simple was bumped)
        assert (
            fixtures.get_dependency_version("app-transitive", "app-simple") == "1.0.1"
        )

    def test_dry_run_does_not_update_dependency_versions(self, fixtures):
        # Create a chart with specific version
        chart_dir = fixtures.charts_dir / "dry-run-dep-test"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: dry-run-dep-test
version: 1.0.0
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.run("common", "patch")

        # Dependency version should NOT be updated
        assert fixtures.get_dependency_version("dry-run-dep-test", "common") == "0.4.27"

    def test_does_not_update_remote_dependency_versions(self, fixtures):
        # Create a chart with both local and remote dependencies
        chart_dir = fixtures.charts_dir / "remote-and-local-deps"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: remote-and-local-deps
version: 1.0.0
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
  - name: external-chart
    repository: https://helm.example.io
    version: "2.0.0"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("common", "patch")

        # common should be updated
        assert (
            fixtures.get_dependency_version("remote-and-local-deps", "common")
            == "0.4.28"
        )
        # external-chart should NOT be updated
        assert (
            fixtures.get_dependency_version("remote-and-local-deps", "external-chart")
            == "2.0.0"
        )


class TestMultipleDependencies:
    """Test handling of charts with multiple local dependencies (like kafka)."""

    def test_detects_second_dependency_in_chart(self, fixtures):
        """
        Test that second (and subsequent) dependencies are properly detected.
        This was a bug in the bash version where BASH_REMATCH was overwritten.
        """
        # Create a chart similar to kafka with two local dependencies
        chart_dir = fixtures.charts_dir / "kafka-like"
        chart_dir.mkdir(parents=True, exist_ok=True)

        # Create sizing chart
        fixtures.create_chart("suse-observability-sizing", "0.1.2")

        chart_yaml_content = """apiVersion: v2
name: kafka-like
version: 19.1.3-suse-observability.7
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
  - name: suse-observability-sizing
    repository: file://../suse-observability-sizing/
    version: "0.1.2"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=True)
        bumper.build_dependency_graph()

        # Should find kafka-like as dependent of suse-observability-sizing
        dependents_of_sizing = bumper.find_all_dependents("suse-observability-sizing")
        assert "kafka-like" in dependents_of_sizing

        # Should also find kafka-like as dependent of common
        dependents_of_common = bumper.find_all_dependents("common")
        assert "kafka-like" in dependents_of_common

    def test_bumps_chart_with_multiple_dependencies(self, fixtures):
        """Test that charts with multiple dependencies get properly bumped."""
        # Create sizing chart
        fixtures.create_chart("suse-observability-sizing", "0.1.2")

        # Create kafka-like chart
        chart_dir = fixtures.charts_dir / "kafka-like"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: kafka-like
version: 19.1.3-suse-observability.7
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
  - name: suse-observability-sizing
    repository: file://../suse-observability-sizing/
    version: "0.1.2"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir), dry_run=False)
        bumper.run("suse-observability-sizing", "patch")

        # kafka-like should be bumped (suffix version, so only suffix incremented)
        assert fixtures.get_version("kafka-like") == "19.1.3-suse-observability.8"

        # suse-observability-sizing dependency should be updated
        assert (
            fixtures.get_dependency_version("kafka-like", "suse-observability-sizing")
            == "0.1.3"
        )


class TestSkipCharts:
    """Test the --skip functionality to skip version bumps for specific charts."""

    def test_skip_single_chart_version_bump(self, fixtures):
        """Test that skipped charts don't get their version bumped."""
        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts={"app-simple"},
        )
        bumper.run("common", "patch")

        # app-simple should NOT be bumped (it's skipped)
        assert fixtures.get_version("app-simple") == "1.0.0"

        # But other dependents should still be bumped
        assert fixtures.get_version("app-with-suffix") == "3.6.9-suse-observability.9"

    def test_skip_chart_still_updates_dependencies(self, fixtures):
        """Test that skipped charts still get their dependency versions updated."""
        # Set up app-simple with specific common version
        chart_dir = fixtures.charts_dir / "app-simple"
        chart_yaml_content = """apiVersion: v2
name: app-simple
version: 1.0.0
dependencies:
  - name: common
    repository: file://../common/
    version: "0.4.27"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts={"app-simple"},
        )
        bumper.run("common", "patch")

        # app-simple's version should NOT change
        assert fixtures.get_version("app-simple") == "1.0.0"

        # But app-simple's dependency on common SHOULD be updated
        assert fixtures.get_dependency_version("app-simple", "common") == "0.4.28"

    def test_skip_multiple_charts(self, fixtures):
        """Test skipping multiple charts."""
        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts={"app-simple", "app-with-suffix", "app-multi-dep"},
        )
        bumper.run("common", "patch")

        # All skipped charts should keep original versions
        assert fixtures.get_version("app-simple") == "1.0.0"
        assert fixtures.get_version("app-with-suffix") == "3.6.9-suse-observability.8"
        assert fixtures.get_version("app-multi-dep") == "5.0.0"

        # Transitive dependents (not skipped) should still be bumped
        assert fixtures.get_version("app-transitive") == "1.0.1"

    def test_skip_transitive_dependent(self, fixtures):
        """Test skipping a chart that's a transitive dependent."""
        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts={"app-transitive"},
        )
        bumper.run("common", "patch")

        # app-simple should be bumped (direct dependent, not skipped)
        assert fixtures.get_version("app-simple") == "1.0.1"

        # app-transitive should NOT be bumped (skipped)
        assert fixtures.get_version("app-transitive") == "1.0.0"

        # app-deep-transitive should still be bumped (depends on app-transitive)
        assert fixtures.get_version("app-deep-transitive") == "2.0.1"

    def test_skip_with_dependency_chain(self, fixtures):
        """Test that skipped charts have dependency versions updated in the chain."""
        # Set up specific versions in dependency chain
        # app-transitive -> app-simple -> common

        chart_dir = fixtures.charts_dir / "app-transitive"
        chart_yaml_content = """apiVersion: v2
name: app-transitive
version: 1.0.0
dependencies:
  - name: app-simple
    repository: file://../app-simple/
    version: "1.0.0"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts={"app-transitive"},
        )
        bumper.run("common", "patch")

        # app-simple gets bumped
        assert fixtures.get_version("app-simple") == "1.0.1"

        # app-transitive should NOT have its version bumped
        assert fixtures.get_version("app-transitive") == "1.0.0"

        # But app-transitive's dependency on app-simple SHOULD be updated
        assert (
            fixtures.get_dependency_version("app-transitive", "app-simple") == "1.0.1"
        )

    def test_skip_empty_set(self, fixtures):
        """Test that empty skip set doesn't affect behavior."""
        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts=set(),
        )
        bumper.run("common", "patch")

        # All dependents should be bumped normally
        assert fixtures.get_version("app-simple") == "1.0.1"
        assert fixtures.get_version("app-with-suffix") == "3.6.9-suse-observability.9"

    def test_skip_nonexistent_chart(self, fixtures):
        """Test that skipping a non-existent chart doesn't cause errors."""
        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts={"nonexistent-chart"},
        )
        # Should not raise an error
        bumper.run("common", "patch")

        # All dependents should be bumped normally
        assert fixtures.get_version("app-simple") == "1.0.1"

    def test_skip_target_chart_warning(self, fixtures, capsys):
        """Test that skipping the target chart itself still bumps it (skip only affects dependents)."""
        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=False,
            skip_charts={"common"},  # Skip the target itself
        )
        bumper.run("common", "patch")

        # The target chart should still be bumped (skip only affects dependents)
        assert fixtures.get_version("common") == "0.4.28"

    def test_skip_dry_run(self, fixtures):
        """Test skip functionality in dry-run mode."""
        original_simple = fixtures.get_version("app-simple")
        original_suffix = fixtures.get_version("app-with-suffix")

        bumper = BumpChartVersion(
            str(fixtures.charts_dir),
            dry_run=True,
            skip_charts={"app-simple"},
        )
        bumper.run("common", "patch")

        # No changes should be made in dry-run mode
        assert fixtures.get_version("app-simple") == original_simple
        assert fixtures.get_version("app-with-suffix") == original_suffix


class TestCheckMode:
    """Test the --check mode for validating dependency versions."""

    def test_check_passes_when_dependencies_correct(self, fixtures):
        """Test that check passes when all dependencies have correct versions."""
        # Create a new base chart that has no other dependents
        fixtures.create_chart("check-base", "1.0.0")

        # Create a chart with correct dependency version
        chart_dir = fixtures.charts_dir / "dep-consumer"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: dep-consumer
version: 1.0.0
dependencies:
  - name: check-base
    repository: file://../check-base/
    version: "1.0.0"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir))
        result = bumper.check_dependencies("check-base")

        assert result is True

    def test_check_fails_when_dependency_outdated(self, fixtures):
        """Test that check fails when a dependency has an outdated version."""
        # Create a new base chart that has no other dependents
        fixtures.create_chart("check-base-outdated", "2.0.0")

        # Create a chart with outdated dependency version
        chart_dir = fixtures.charts_dir / "dep-consumer-outdated"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: dep-consumer-outdated
version: 1.0.0
dependencies:
  - name: check-base-outdated
    repository: file://../check-base-outdated/
    version: "1.0.0"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir))
        result = bumper.check_dependencies("check-base-outdated")

        assert result is False

    def test_check_fails_for_nonexistent_chart(self, fixtures):
        """Test that check fails for a non-existent chart."""
        bumper = BumpChartVersion(str(fixtures.charts_dir))
        result = bumper.check_dependencies("nonexistent")

        assert result is False

    def test_check_passes_when_no_dependents(self, fixtures):
        """Test that check passes when there are no dependent charts."""
        bumper = BumpChartVersion(str(fixtures.charts_dir))
        result = bumper.check_dependencies("app-no-deps")

        assert result is True

    def test_check_only_validates_local_dependencies(self, fixtures):
        """Test that check only considers local (file://) dependencies."""
        # Create a new base chart that has no pre-existing dependents
        fixtures.create_chart("check-base-remote", "1.0.0")

        # Create a chart with only a remote dependency named 'check-base-remote'
        chart_dir = fixtures.charts_dir / "remote-common-dep"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: remote-common-dep
version: 1.0.0
dependencies:
  - name: check-base-remote
    repository: https://helm.example.io
    version: "0.1.0"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir))
        result = bumper.check_dependencies("check-base-remote")

        # Should pass because remote deps aren't tracked in reverse dependency graph
        assert result is True

    def test_check_multiple_dependents_one_outdated(self, fixtures):
        """Test that check fails if even one dependent is outdated."""
        # Create a new base chart
        fixtures.create_chart("check-base-multi", "1.0.0")

        # Create two charts: one with correct version, one outdated
        for name, version in [("correct-dep", "1.0.0"), ("outdated-dep", "0.5.0")]:
            chart_dir = fixtures.charts_dir / name
            chart_dir.mkdir(parents=True, exist_ok=True)

            chart_yaml_content = f"""apiVersion: v2
name: {name}
version: 1.0.0
dependencies:
  - name: check-base-multi
    repository: file://../check-base-multi/
    version: "{version}"
"""
            with open(chart_dir / "Chart.yaml", "w") as f:
                f.write(chart_yaml_content)

        bumper = BumpChartVersion(str(fixtures.charts_dir))
        result = bumper.check_dependencies("check-base-multi")

        assert result is False

    def test_check_does_not_modify_files(self, fixtures):
        """Test that check mode never modifies any files."""
        # Create a new base chart and a dependent with outdated version
        fixtures.create_chart("check-base-nomod", "2.0.0")

        chart_dir = fixtures.charts_dir / "outdated-dep-nomod"
        chart_dir.mkdir(parents=True, exist_ok=True)

        chart_yaml_content = """apiVersion: v2
name: outdated-dep-nomod
version: 1.0.0
dependencies:
  - name: check-base-nomod
    repository: file://../check-base-nomod/
    version: "1.0.0"
"""
        with open(chart_dir / "Chart.yaml", "w") as f:
            f.write(chart_yaml_content)

        original_base_version = fixtures.get_version("check-base-nomod")
        original_dep_version = fixtures.get_version("outdated-dep-nomod")
        original_dep_ref = fixtures.get_dependency_version(
            "outdated-dep-nomod", "check-base-nomod"
        )

        bumper = BumpChartVersion(str(fixtures.charts_dir))
        bumper.check_dependencies("check-base-nomod")

        # No files should be modified
        assert fixtures.get_version("check-base-nomod") == original_base_version
        assert fixtures.get_version("outdated-dep-nomod") == original_dep_version
        assert (
            fixtures.get_dependency_version("outdated-dep-nomod", "check-base-nomod")
            == original_dep_ref
        )


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
