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
$ brew install helm helm-docs shellcheck
```
