#!/usr/bin/env bash

# Usage: hack/run-e2e.sh
# Example 1: hack/run-e2e.sh (run e2e test)

# Get the OS type. 
OSTYPE=$1

set -o errexit
set -o nounset
set -o pipefail

# Install ginkgo
GO111MODULE=on go install github.com/onsi/ginkgo/v2/ginkgo@v2.0.0

# Build kusion binary according to the OS type. 
go generate ./pkg/version
if [ $OSTYPE == "windows" ]; then
    go build -o bin/kusion.exe .
else
    go build -o bin/kusion .
fi


# Run e2e
set +e
ginkgo  ./test/e2e/
TESTING_RESULT=$?


exit $TESTING_RESULT
