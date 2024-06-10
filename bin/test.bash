#!/bin/bash

set -eu
set -o pipefail

./scripts/ci/run_fstest

# shellcheck disable=SC2068
# Double-quoting array expansion here causes ginkgo to fail
# Tee output to a log file but exclude component/test logs from stdout so
# concourse output is not overloaded
go run github.com/onsi/ginkgo/v2/ginkgo ${@} | tee /tmp/simulation-output.log | grep -v '{"timestamp"'
