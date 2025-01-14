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

## Testing the Helm charts

The Helm chart repository supports testing Helm charts using the [Terratest](https://terrates.gruntwork.io/) library. In order to run tests for a chart, you can invoke the following command from the root of the repository:

```shell
$ go test ./stable/<chart>/test/...
```

To run integration tests put them in an `itest` directory and run:
```shell
$ go test ./stable/<chart>/itest/...
```

The test-set for a chart is in the `stable/<chart>/itest` directory.

You are encouraged to add more tests when working on the Helm charts ;).

## Chart build scripting

We use gawk (instead of awk) for consistent local and container scripting.

```shell
$ brew install gawk
```

```shell
$ apk add gawk
```

## Design docs

We list various design docs here that accompany the charts:

- SUSE Observability router maintenance mode: https://lucid.app/lucidchart/1dd84a54-2908-4e33-9b5d-21705c461dab/edit
