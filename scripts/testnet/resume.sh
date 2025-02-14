#!/usr/bin/env bash

# Resume the paused testnet, by sending SIGCONT to all the processes of the
# testnet (seednode, observers, validators, proxy, txgen)

export DHARITRITESTNETSCRIPTSDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

if [ "$1" == "keep" ]; then
  KEEPOPEN=1
else
  KEEPOPEN=0
fi

source "$DHARITRITESTNETSCRIPTSDIR/variables.sh"
source "$DHARITRITESTNETSCRIPTSDIR/include/validators.sh"
source "$DHARITRITESTNETSCRIPTSDIR/include/observers.sh"
source "$DHARITRITESTNETSCRIPTSDIR/include/tools.sh"

if [ $USE_PROXY -eq 1 ]; then
  resumeProxy
fi

if [ $USE_TXGEN -eq 1 ]; then
  resumeTxGen
fi

resumeValidators
resumeObservers
resumeSeednode
