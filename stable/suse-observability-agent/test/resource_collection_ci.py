#!/usr/bin/env python3
"""Validate the chart's canonical CRD and collector configuration."""

import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import urllib.request


def run(*args, **kwargs):
    return subprocess.run(args, check=True, text=True, **kwargs)


def chart_image(repo, chart):
    output = run(
        "helm", "template", "suse-observability-agent", str(chart),
        "-f", str(chart / "test/values/minimal.yaml"),
        "--set", "otel.enabled=true", "--set", "otel.prometheusScraping.enabled=false",
        "--show-only", "templates/otel/k8sresourcecollector/deployment.yaml",
        stdout=subprocess.PIPE, cwd=repo,
    ).stdout
    deployment = json.loads(run(
        "yq", "-o=json", ".", input=output, stdout=subprocess.PIPE,
    ).stdout)
    containers = deployment["spec"]["template"]["spec"]["containers"]
    return next(item["image"] for item in containers if item["name"] == "k8s-resource-collector")


def image_binary(image, root):
    run("docker", "pull", image)
    container = run("docker", "create", image, stdout=subprocess.PIPE).stdout.strip()
    binary = root / "agent"
    try:
        run("docker", "cp", f"{container}:/usr/bin/sts-opentelemetry-agent-collector", str(binary))
    finally:
        run("docker", "rm", container, stdout=subprocess.DEVNULL)
    binary.chmod(0o755)
    return binary


def image_revision(image):
    labels = json.loads(run(
        "docker", "image", "inspect", image, "--format", "{{json .Config.Labels}}",
        stdout=subprocess.PIPE,
    ).stdout) or {}
    revision = labels.get("org.opencontainers.image.revision", "")
    if (labels.get("org.opencontainers.image.source") !=
            "https://github.com/StackVista/sts-opentelemetry-collector"
            or not re.fullmatch(r"[a-f0-9]{40}", revision)):
        raise RuntimeError("Collector image must identify its canonical source commit")
    return revision


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--canonical-crd", type=Path,
                        help="local candidate export, requires --agent-binary")
    parser.add_argument("--agent-binary", type=Path,
                        help="local candidate binary; CI extracts the chart's pinned image")
    parser.add_argument("--go", default="go")
    args = parser.parse_args()
    if bool(args.canonical_crd) != bool(args.agent_binary):
        parser.error("--canonical-crd and --agent-binary must be supplied together")
    repo = Path(__file__).resolve().parents[3]
    chart = repo / "stable" / "suse-observability-agent"
    imported = chart / "files/crds/resourcecollections.observability.suse.com.yaml"
    source_path = "receiver/k8sresourcereceiver/resourcecollection/resourcecollection-crd.yaml"

    with tempfile.TemporaryDirectory(prefix="resourcecollection-chart-ci-") as temporary:
        root = Path(temporary)
        canonical = root / "canonical.yaml"
        if args.canonical_crd:
            binary = args.agent_binary.resolve()
            canonical.write_bytes(args.canonical_crd.resolve().read_bytes())
        else:
            image = chart_image(repo, chart)
            binary = image_binary(image, root)
            revision = image_revision(image)
            url = f"https://raw.githubusercontent.com/StackVista/sts-opentelemetry-collector/{revision}/{source_path}"
            with urllib.request.urlopen(url, timeout=30) as response:
                canonical.write_bytes(response.read())
        exported = canonical.read_bytes()
        if imported.read_bytes() != exported:
            raise RuntimeError("Chart CRD differs from the collector image's canonical export")
        if not binary.is_file() or not os.access(binary, os.X_OK):
            raise RuntimeError("An executable candidate agent binary is required")

        env = os.environ | {
            "KUBECONFIG": "/dev/null",
            "RESOURCE_COLLECTION_REQUIRE_BINARY": "true",
            "RESOURCE_COLLECTION_CANONICAL_CRD": str(canonical),
            "OTEL_AGENT_BINARY": str(binary),
        }
        run(
            args.go, "test", "-count=1", "-timeout=10m", "-v",
            "-run", "^TestResourceCollection", "./stable/suse-observability-agent/test/...",
            "./updatecli/updatecli.d/update-docker-images/...",
            cwd=repo, env=env,
        )


if __name__ == "__main__":
    main()
