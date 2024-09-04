#!/usr/bin/env bash

set -eux

export DOCKERTESTNETDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

DHARITRITESTNETSCRIPTSDIR="$(dirname "$DOCKERTESTNETDIR")/testnet"

source "$DOCKERTESTNETDIR/variables.sh"

cd ${DHARITRIDIR}

docker build -f docker/seednode/Dockerfile . -t seednode:dev

docker build -f docker/node/Dockerfile . -t node:dev

