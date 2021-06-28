#!/bin/bash

# Create local sns
aws sns create-topic --name "$AWS_SNS_TOPIC_NAME" --endpoint-url "http://$LOCALSTACK_HOSTNAME:$EDGE_PORT" --profile "$AWS_PROFILE_NAME"

# Create the single dynamodb table
aws dynamodb create-table \
    --table-name $AWS_DDB_TABLE_NAME \
    --attribute-definitions AttributeName=pk,AttributeType=S AttributeName=sk,AttributeType=S AttributeName=id,AttributeType=S AttributeName=date,AttributeType=S \
    --key-schema AttributeName=pk,KeyType=HASH AttributeName=sk,KeyType=RANGE \
    --global-secondary-indexes '[
        {
            "IndexName": "gsi_date",
            "KeySchema": [
                {"AttributeName":"date","KeyType":"HASH"},
                {"AttributeName":"sk","KeyType":"RANGE"}
            ],
            "Projection": {
                "ProjectionType":"ALL"
            }
        },
        {
            "IndexName": "gsi_uuid",
            "KeySchema": [
                {"AttributeName":"id","KeyType":"HASH"}
            ],
            "Projection": {
                "ProjectionType":"ALL"
            }
        }
    ]' \
    --billing-mode PAY_PER_REQUEST \
    --endpoint-url "http://$LOCALSTACK_HOSTNAME:$EDGE_PORT" \
    --profile $AWS_PROFILE_NAME
