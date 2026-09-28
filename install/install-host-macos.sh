#!/bin/sh
# Build the daemon and register the Chrome native messaging host for this user.
set -eu
cd "$(dirname "$0")/.."
mkdir -p daemon/bin
( cd daemon && go build -o bin/cloakwork ./cmd/cloakwork )
exec ./daemon/bin/cloakwork install-host
