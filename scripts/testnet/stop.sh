#!/usr/bin/env bash

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
  stopProxy
fi

if [ $USE_TXGEN -eq 1 ]; then
  stopTxGen
fi

stopValidators
stopObservers
stopSeednode

if [ $USETMUX -eq 1 ] && [ $KEEPOPEN -eq 0 ]
then
  tmux kill-session -t "dharitri-tools"
  tmux kill-session -t "dharitri-nodes"
fi
