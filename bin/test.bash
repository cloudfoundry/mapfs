#!/bin/bash

set -eu
set -o pipefail

THIS_FILE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
pushd "${BUILD_ROOT_DIR}"
"${THIS_FILE_DIR}/run_fstest"
popd

# shellcheck disable=SC2068
# Double-quoting array expansion here causes ginkgo to fail
# Tee output to a log file but exclude component/test logs from stdout so
# concourse output is not overloaded
go run github.com/onsi/ginkgo/v2/ginkgo ${@} | tee /tmp/simulation-output.log | grep -v '{"timestamp"'
