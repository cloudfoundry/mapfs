#!/bin/bash

docker run -it -v \
/Users/pivotal/workspace/mapfs-release/src:/go/src/ \
-w /go/src/code.cloudfoundry.org/mapfs \
--privileged \
cfpersi/mapfs-tests \
ginkgo -r .
