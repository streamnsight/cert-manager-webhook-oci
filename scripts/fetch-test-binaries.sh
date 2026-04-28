#!/bin/sh
set -eu

REPO_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
BIN_DIR="${REPO_ROOT}/kubebuilder/bin"
K8S_VERSION="${K8S_VERSION:-1.34.x}"
SETUP_ENVTEST_BIN_DIR=$(mktemp -d)

cleanup() {
	rm -rf "${SETUP_ENVTEST_BIN_DIR}"
}

trap cleanup EXIT

mkdir -p "${BIN_DIR}"

GOTOOLCHAIN=auto GOBIN="${SETUP_ENVTEST_BIN_DIR}" go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
"${SETUP_ENVTEST_BIN_DIR}/setup-envtest" use "${K8S_VERSION}" --bin-dir "${BIN_DIR}" -p path >/dev/null

printf 'Fetched envtest assets for Kubernetes %s into %s\n' "${K8S_VERSION}" "${BIN_DIR}"
