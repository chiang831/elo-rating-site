#!/bin/bash

GITROOT=$(git rev-parse --show-toplevel)

if [ -e $GITROOT/.env ]; then
  source $GITROOT/.env/bin/activate
fi

export PATH=$GITROOT/bin:$PATH
export PYTHONPATH=$GITROOT:$PYTHONPATH

if [ $# -eq 0 ]
  then
    echo "No arguments supplied, use default config.sh"
    CONFIG_FILE=$GITROOT/bin/config.sh
else
    CONFIG_FILE=$GITROOT/bin/config_$1.sh
    echo "Use config $CONFIG_FILE"
fi

if [ -e $CONFIG_FILE ]; then
  echo "source $CONFIG_FILE"
  source $CONFIG_FILE
fi
