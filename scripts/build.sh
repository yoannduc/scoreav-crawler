#!/bin/bash
curr_branch=$(git rev-parse --abbrev-ref HEAD --)

source ./scripts/flags.sh $@

# TODO Check build target extension and build accordingly

if [[ "$func" == *" "* ]]; then
    echo "Please provide a single function to build"
    exit 1
fi

if [ ! -f "cmd/$func/$func."* ]; then
    echo "No function existing for $func"
    exit 2
fi

if [ -f "build/$func" ]; then
    rm "build/$func"
fi

if [ ! -z "$version" ]; then
    git fetch --tags
    if git rev-parse "$version" >/dev/null 2>&1; then
        git checkout $version
    else
        echo "Uknown version"
        exit 3
    fi
    
fi

docker run --rm -v "$PWD":/usr/src/main -w /usr/src/main golang:latest sh -c "printenv && GOOS=linux GOARCH=amd64 go build -o build/$func cmd/$func/$func."*

if [ ! -z "$version" ]; then
    git checkout $curr_branch
fi
