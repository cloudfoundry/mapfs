#!/bin/bash

docker run --rm -it -v \
/Users/pivotal/workspace/mapfs-release/src:/go/src/ \
-w /go/src/code.cloudfoundry.org/mapfs \
golang \
bash -c "go get github.com/onsi/ginkgo/ginkgo && ginkgo -r ."