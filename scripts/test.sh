#!/bin/bash

docker run -it -v \
/Users/pivotal/workspace/mapfs-release/src:/go/src/ \
-w /go/src/mapfs \
--privileged \
cfpersi/mapfs-tests \
ginkgo  -r .
