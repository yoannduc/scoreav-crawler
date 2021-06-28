#!/bin/bash

help() {
    echo "-----------------------------------------------------------------------"
    echo "                      Available commands                              -"
    echo "-----------------------------------------------------------------------"
    echo -e -n "$BLUE"
    echo "   > build - To build a function"
    echo "   > test - To test a function locally"
    echo "   > build-test - To build then test a function locally"
    echo "   > deploy - To deploy a function to aws"
    echo "   > build-deploy - To build then deploy a function to aws"
    echo -e -n "$NORMAL"
    echo "-----------------------------------------------------------------------"
}

build() {
    ./scripts/build.sh $@
}

test() {
    ./scripts/test.sh $@
}

build-test() {
    ./scripts/build.sh $@ && ./scripts/test.sh $@
}

deploy() {
    ./scripts/deploy.sh $@
}

build-deploy() {
    ./scripts/build.sh $@ && ./scripts/deploy.sh $@
}

if declare -f "$1" > /dev/null
then
  $*
else
  echo "'$*' is not a known function name" >&2
  help
  exit 1
fi
