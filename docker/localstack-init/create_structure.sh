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

# # Load dynamo data-history data
# aws dynamodb batch-write-item \
#     --request-items "{
#         \"$AWS_DDB_TABLE_NAME\": [
#             {
#                 \"PutRequest\": {
#                     \"Item\": {
#                         \"pk\": {\"S\": \"notification\"},
#                         \"sk\": {\"S\": \"mail#scoreav#news#2021-06-26T00:24:53Z\"},
#                         \"date\": {\"S\": \"2021-06-26T00:24:53Z\"},
#                         \"last_element_notified_pk\": {\"S\": \"scoreav#news\"},
#                         \"last_element_notified_sk\": {\"S\": \"2021-06-25#http://www.scoreav.com/the-judas-knife-avec-son-single-et-son-couteau/\"},
#                         \"source\": {\"S\": \"scoreav\"},
#                         \"type\": {\"S\": \"news\"}
#                     }
#                 }
#             }
#         ]
#     }" \
#     --endpoint-url "http://$LOCALSTACK_HOSTNAME:$EDGE_PORT" \
#     --profile "$AWS_PROFILE_NAME"
