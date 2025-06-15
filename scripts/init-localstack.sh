#!/bin/bash

# Exit on error
set -e

# Create tables
create_table() {
  local table_name=$1
  local key_schema=$2
  local attribute_definitions=$3
  
  aws dynamodb create-table \
    --endpoint-url http://localhost:4566 \
    --table-name "$table_name" \
    --key-schema "$key_schema" \
    --attribute-definitions "$attribute_definitions" \
    --billing-mode PAY_PER_REQUEST
}

# Wait for LocalStack to be ready
until aws --endpoint-url=http://localhost:4566 dynamodb list-tables; do
  echo "Waiting for LocalStack to be ready..."
  sleep 2
done

# Create your DynamoDB tables here
# Example:
# create_table "Users" "AttributeName=userId,KeyType=HASH" "AttributeName=userId,AttributeType=S"
# create_table "Games" "AttributeName=gameId,KeyType=HASH" "AttributeName=gameId,AttributeType=S"

echo "LocalStack initialization complete!"

# Keep the container running
tail -f /dev/null
