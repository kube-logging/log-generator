#!/bin/bash

set -x
set -e

# get artifactory creds
sed -i '1d' "${NYOTA_CREDENTIALS_FILE}"
. "${NYOTA_CREDENTIALS_FILE}"

export GOPRIVATE='github.cisco.com,github.com/banzaicloud'
export GOPROXY="https://proxy.golang.org,https://${artifactory_user}:${artifactory_password}@${artifactory_url},direct"
export GONOPROXY='gopkg.in,go.uber.org'

export GOPATH=$(go env GOPATH)
export PATH="${PATH}:${GOPATH}/bin"
export GOFLAGS='-mod=readonly'

echo "Check"
make check
