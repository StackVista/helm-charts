# Purpose

This project provides an off-the shelve k8s sidecar to inject the x-request-id header coming into and
out of a k8s pod.

TODO: We might unit/it test this using the just framework used by linkerd: https://github.com/linkerd/linkerd2-proxy-init

# Overview

- `/examples`. Deployable k8s examples showing the http header injection
- `/charts/http-header-injector`. A mutating admission webhook that will inject the iptables initializer and a proxy into annotated pods, such
  that all incoming http traffic gets a `x-request-id` and responses get the same `x-request-id`.
- `./proxy`. Container definition doing the http rewriting
- `./proxy-init`. Init container definition, which rewrites ip tables to make sure all traffic goes through the proxy.

# Deploy

- the `./charts/http-header-injector` can be installed using helm `helm upgrade --install --namespace http-header-injector --create-namespace http-header-injector ./charts/http-header-injector`
- The `./examples/install.sh` can be used to install/uninstall the various examples

# Releasing

The docker images under `./proxy` and `./proxy-init` will be built on the `main` branch when there are changes, and will be tagged with
the git hash.

The helm chart is using semver and should be bumped every time the chart is changed. Builds are ran when there are changes.
