# helm-charts

StackState curated applications for Kubernetes.

## Pre-commit hooks

This Helm repository has pre-commit hooks for Helm specific needs:

* Makes sure all charts pass a `helm lint` check.
* Makes sure all shell scripts pass the `shellcheck`.
* Updates the `README.md` file of all charts based on comments in that chart's `values.yaml` file.

### Install `pre-commit` binary

The binary for `pre-commit` can be installed via Homebrew:

```shell
$ brew install pre-commit
```

### Install git hooks

After the `pre-commit` binary is installed, go to this repository's directory, and run the following commmands to install the git hook:

```shell
$ pre-commit install
$ pre-commit install-hooks
```

### Install hook dependencies

The pre-commit hooks themselves call binaries under the hood; they can be installed via the following command:

```shell
$ brew install helm shellcheck
$ brew install norwoodj/tap/helm-docs
```

NOTE: The templates for README generation are only compatible with helm-docs 0.15+.

#### Helm-docs

Our helm-docs pre-commit requires helm-docs 1.6.0 or helm-docs 1.7.0. The current brew version is 1.8.1. Because we cannot get older version from brew, we made a nix flake file that brings helm-docs into scope. Getting into the nix environment is simply:

```shell
$ nix develop
```

### Installing nix

To install Nix follow the [official installation instructions](https://nixos.org/download.html) for your platform. The script, when executed, is fully interactive and transparent in changes applied to the host system. It's worth to read the installer output to get some additional insights about Nix.

#### Configuration

Nix dev shell is using built-in feature named [Flakes](https://nixos.wiki/wiki/Flakes). At the time of writing this (`nix 2.5.0`) flakes is considered stable but it still tagged as an experimental-feature. You have to update your nix config to allow use of experimental features in case if the commands bellow throw errors for you.


```shell
mkdir -p ~/.config/nix && echo "experimental-features = nix-command flakes" >> ~/.config/nix/nix.conf
```

## Testing the Helm charts

The Helm chart repository supports testing Helm charts using the [Terratest](https://terratest.gruntwork.io/) library. In order to run tests for a chart, you can invoke the following command from the root of the repository:

```shell
$ go test ./stable/<chart>/test/...
```

To run integration tests put them in an `itest` directory and run:
```shell
$ go test ./stable/<chart>/itest/...
```

The test-set for a chart is in the `stable/<chart>/itest` directory.

You are encouraged to adding more tests when working on the Helm charts ;).

## Chart build scripting

We use gawk (instead of awk) for consistent local and container scripting.

```shell
$ brew install gawk
```

```shell
$ apk add gawk
```
