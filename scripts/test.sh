#!/bin/bash
source ./scripts/flags.sh $@

if [[ "$func" == *" "* ]]; then
    echo "Please provide a single function to test"
    exit 1
fi

if [ ! -f "build/$func" ]; then
    echo "No built function found for $func"
    exit 2
fi

if [ ! -f "cmd/$func/.env" ]; then
    echo "No env file found for $func at cmd/$func/.env"
    exit 3
fi

docker run --rm -it -v "$PWD":/var/task:ro,delegated --env-file cmd/$func/.env -e ENV=$env -e AWS_LAMBDA_FUNCTION_INVOKED_ARN=function:$func:$env lambci/lambda:go1.x build/$func $event