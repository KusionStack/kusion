#!/usr/bin/env bash

# Usage: hack/run-e2e.sh
# Example 1: hack/run-e2e.sh (run e2e test)

set -o errexit
set -o nounset
set -o pipefail

# Install ginkgo
GO111MODULE=on go install github.com/onsi/ginkgo/v2/ginkgo@v2.0.0

# Build kusion binary
go build -o bin/kusion ./cmd/kusionctl/kusionctl.go


# Run e2e
set +e
ginkgo  ./test/e2e/ 
TESTING_RESULT=$?


exit $TESTING_RESULT