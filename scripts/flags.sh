#!/bin/bash
version=""
env="local"
profile="dev"
event='{}'

while test $# -gt 0; do
    case "$1" in
        -e | --env)
            env=$2
            shift 2
            ;;
        --env=*)
            env=${1##"--env="}
            shift
            ;;
        -v | --version)
            version=$2
            shift 2
            ;;
        --version=*)
            version=${1##"--version="}
            shift
            ;;
        -p | --profile)
            profile=$2
            shift 2
            ;;
        --profile=*)
            profile=${1##"--profile="}
            shift
            ;;
        --event)
            event=$2
            shift 2
            ;;
        --event=*)
            event=${1##"--event="}
            shift
            ;;
        *)
            break
            ;;
    esac
done

export version
export env
export profile
export event
func=$@
export func
