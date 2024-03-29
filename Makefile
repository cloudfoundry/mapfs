test:
	docker run -it -v \
    ${HOME}/workspace/mapfs-release/src:/go/src/ \
    -w /go/src/mapfs \
    --privileged \
    cfpersi/mapfs-tests \
    go run github.com/onsi/ginkgo/v2/ginkgo -r -v --flake-attempts 3 .

# Note: the fstest suite is available at https://github.com/zfsonlinux/fstest
fly-fstest:
	fly -t persi e -c scripts/ci/run_fstest.build.yml \
    -i mapfs=${HOME}/workspace/mapfs \
    -i fstest=${HOME}/workspace/fstest \
    -p

.PHONY: test fly-fstest