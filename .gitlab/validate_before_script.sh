#!/usr/bin/env sh

[ "${TRACE}" ] && set -x

installDependencies() {
  apk -Uuv add bash curl groff less openssl
  curl -fSL https://git.io/get_helm.sh | bash -
}

installGitRemotes() {
  # shellcheck disable=SC2154
  git remote add helm https://oauth2:"${gitlab_api_scope_token}"@gitlab.com/stackvista/devops/helm-charts
  git fetch helm
}

configureKubectl() {
  kubectl config set-cluster k3s --server https://k3s-test:6443 --insecure-skip-tls-verify
  kubectl config set-credentials k3s --username=node --password=somesecret
  kubectl config set-context k3s --cluster=k3s --user=k3s
  kubectl config use-context k3s

  until kubectl version > /dev/null 2>&1; do
    echo "Waiting until k3s is online..."
    sleep 2
  done

  # Install Docker image pull secrets
  # shellcheck disable=SC2154
  kubectl create secret docker-registry stackstate-dockerhub-pull-secret \
    --docker-password="${docker_password}" \
    --docker-server='docker.io' \
    --docker-username="${docker_user}"

  # shellcheck disable=SC2154
  kubectl create secret docker-registry stackstate-quay-pull-secret \
    --docker-password="${quay_password}" \
    --docker-server='quay.io' \
    --docker-username="${quay_user}"
}

configureHelm() {
  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tiller
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: tiller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: tiller
    namespace: kube-system
EOF

  helm init --service-account tiller

  until helm version > /dev/null 2>&1; do
    echo "Waiting until Helm is online..."
    sleep 2
  done
}

installDependencies
installGitRemotes
configureKubectl
configureHelm
