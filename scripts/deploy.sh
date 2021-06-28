#!/bin/bash
source ./scripts/flags.sh $@

if [[ "$func" == *" "* ]]; then
    echo "Please provide a single function to deploy"
    exit 1
fi

if [ ! -f "build/$func" ]; then
    echo "No built function found for $func"
    exit 2
fi

if [ ! -f "cmd/$func/.aws-lambda-name" ]; then
    echo "No lambda function name found for $func at cmd/$func/.aws-lambda-name"
    exit 3
fi

# Retrieve lambda name from config file and remove all trailing space
# To retrieve lambda, we want to retrieve exactly one line AFTER the match with profile
lambda=$(cat cmd/$func/.aws-lambda-name | sed -n "/\[$profile\]/{n;p;}")
lambda="${lambda#"${lambda%%[![:space:]]*}"}"
lambda="${lambda%"${lambda##*[![:space:]]}"}"

if [ -z $lambda ]; then
    echo "No lambda name found for $func in cmd/$func/.aws-lambda-name for profile $profile"
    exit 4
fi


# cd in build folder as Handler name would be build/... if we zip from root folder
cd build
zip -qr $lambda.zip $func

if [ ! -f "$lambda.zip" ]; then
    echo "Error creating $lambda.zip for $func"
    exit 5
fi

# Volume both pwd with zip + will host recap json && .aws folder with credentials for profiles
docker run --rm -v $(pwd)/../.aws:/root/.aws -v $(pwd):/aws amazon/aws-cli lambda update-function-code \
    --function-name $lambda \
    --profile $profile \
    --zip-file fileb://$lambda.zip > deploy-$lambda.json

if [ -f "./deploy-$lambda.json" ]; then
    cat deploy-$lambda.json
    rm deploy-$lambda.json
else
    echo "It seems an error occurred while uploading $func to aws"
fi

if [ -f "$lambda.zip" ]; then
    rm $lambda.zip
fi
