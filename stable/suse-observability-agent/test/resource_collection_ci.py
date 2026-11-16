#!/usr/bin/env python3
"""Validate the chart's canonical CRD, collector configuration and Helm lifecycle."""

import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
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


def image_binary(image, root, pull):
    if pull:
        run("docker", "pull", image)
    container = run("docker", "create", image, stdout=subprocess.PIPE).stdout.strip()
    binary = root / "agent"
    try:
        run("docker", "cp", f"{container}:/usr/bin/sts-opentelemetry-agent-collector", str(binary))
    finally:
        run("docker", "rm", container, stdout=subprocess.DEVNULL)
    binary.chmod(0o755)
    return binary


def immutable_image(image):
    digests = json.loads(run(
        "docker", "image", "inspect", image, "--format", "{{json .RepoDigests}}",
        stdout=subprocess.PIPE,
    ).stdout)
    repository = image.split("@", 1)[0]
    if ":" in repository.rsplit("/", 1)[-1]:
        repository = repository.rsplit(":", 1)[0]
    for digest in digests or []:
        if digest.startswith(repository + "@sha256:") and re.fullmatch(
                r"[^@\s]+@sha256:[a-f0-9]{64}", digest):
            return digest
    raise RuntimeError("The validated chart image has no matching immutable repository digest")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--canonical-crd", type=Path,
                        help="local candidate export; CI fetches the committed source")
    parser.add_argument("--agent-binary", type=Path,
                        help="local candidate binary; CI extracts the chart's pinned image")
    parser.add_argument("--go", default="go")
    parser.add_argument("--lifecycle-profile", choices=("native", "topology"))
    parser.add_argument("--collector-source", type=Path,
                        help="collector checkout containing the lifecycle runner")
    parser.add_argument("--agent-image", help="lifecycle image; CI uses the validated immutable digest")
    parser.add_argument("--node-image")
    parser.add_argument("--kubernetes-version")
    parser.add_argument("--artifacts", type=Path)
    args = parser.parse_args()
    if args.lifecycle_profile and not all((
            args.collector_source, args.node_image, args.kubernetes_version, args.artifacts)):
        parser.error("Lifecycle checks require collector source, node image, Kubernetes version and artifacts")
    if args.agent_image and not args.lifecycle_profile:
        parser.error("--agent-image requires --lifecycle-profile")
    repo = Path(__file__).resolve().parents[3]
    chart = repo / "stable" / "suse-observability-agent"
    imported = chart / "files/crds/resourcecollections.observability.suse.com.yaml"
    provenance = json.loads(imported.with_suffix(".provenance.json").read_text())
    source_path = "receiver/k8sresourcereceiver/resourcecollection/resourcecollection-crd.yaml"
    if provenance["repository"] != "StackVista/sts-opentelemetry-collector" or provenance["path"] != source_path:
        raise RuntimeError("Unexpected canonical ResourceCollection source")

    with tempfile.TemporaryDirectory(prefix="resourcecollection-chart-ci-") as temporary:
        root = Path(temporary)
        canonical = root / "canonical.yaml"
        if args.canonical_crd:
            canonical.write_bytes(args.canonical_crd.resolve().read_bytes())
        else:
            revision = provenance["sourceCommit"]
            if provenance["sourceState"] != "committed" or not re.fullmatch(r"[a-f0-9]{40}", revision):
                raise RuntimeError("CI requires provenance pinned to a committed collector export")
            url = f"https://raw.githubusercontent.com/StackVista/sts-opentelemetry-collector/{revision}/{source_path}"
            with urllib.request.urlopen(url, timeout=30) as response:
                canonical.write_bytes(response.read())
        exported = canonical.read_bytes()
        if hashlib.sha256(exported).hexdigest() != provenance["sha256"] or imported.read_bytes() != exported:
            raise RuntimeError("Chart CRD differs from its pinned canonical export")

        if args.lifecycle_profile:
            source = args.collector_source.resolve()
            source_crd = source / source_path
            if source_crd.read_bytes() != exported:
                raise RuntimeError("Lifecycle source CRD differs from the chart's canonical export")
            if not args.canonical_crd:
                head = run("git", "rev-parse", "HEAD", cwd=source,
                           stdout=subprocess.PIPE).stdout.strip()
                if head != provenance["sourceCommit"]:
                    raise RuntimeError("Lifecycle harness must use the pinned collector commit")
            endpoint = run(
                "docker", "context", "inspect", "--format", "{{.Endpoints.docker.Host}}",
                stdout=subprocess.PIPE,
            ).stdout.strip()
            if not endpoint.startswith("unix://"):
                raise RuntimeError("Lifecycle checks require a local Docker daemon")
            image = args.agent_image or chart_image(repo, chart)
            if not args.canonical_crd and not re.fullmatch(
                    r"[^@\s]+@sha256:[a-f0-9]{64}", image):
                raise RuntimeError("Lifecycle CI requires an immutable validated agent image")
            binary = image_binary(image, root, pull=not args.agent_image or not args.canonical_crd)
            if args.agent_binary:
                supplied = hashlib.sha256(args.agent_binary.read_bytes()).digest()
                extracted = hashlib.sha256(binary.read_bytes()).digest()
                if supplied != extracted:
                    raise RuntimeError("Candidate binary differs from the lifecycle image's binary")
        elif args.agent_binary:
            binary = args.agent_binary.resolve()
        else:
            image = chart_image(repo, chart)
            binary = image_binary(image, root, pull=True)
        if not binary.is_file() or not os.access(binary, os.X_OK):
            raise RuntimeError("An executable candidate agent binary is required")

        env = os.environ | {
            "KUBECONFIG": "/dev/null",
            "RESOURCE_COLLECTION_REQUIRE_BINARY": "true",
            "RESOURCE_COLLECTION_CANONICAL_CRD": str(canonical),
            "OTEL_AGENT_BINARY": str(binary),
        }
        if args.lifecycle_profile:
            run(
                sys.executable, str(source / "test/resourcecollection/chart.py"),
                "--chart", str(chart), "--crd", str(canonical),
                "--agent-binary", str(binary), "--agent-image", image,
                "--node-image", args.node_image, "--kubernetes-version", args.kubernetes_version,
                "--profile", args.lifecycle_profile, "--artifacts", str(args.artifacts.resolve()),
                cwd=repo, env=env,
            )
        else:
            run(
                args.go, "test", "-count=1", "-timeout=10m", "-v",
                "-run", "^TestResourceCollection", "./stable/suse-observability-agent/test/...",
                cwd=repo, env=env,
            )
            if not args.agent_binary and os.environ.get("GITHUB_OUTPUT"):
                digest = immutable_image(image)
                with open(os.environ["GITHUB_OUTPUT"], "a") as output:
                    output.write(f"agent_image={digest}\n")


if __name__ == "__main__":
    main()
