package db

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient implements DatabaseClient for AWS DynamoDB
type DynamoDBClient struct {
	client *dynamodb.Client
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClientFromConfig(cfg aws.Config) *DynamoDBClient {
	return &DynamoDBClient{
		client: dynamodb.NewFromConfig(cfg),
	}
}

func NewDynamoDBClientFromClient(client *dynamodb.Client) *DynamoDBClient {
	return &DynamoDBClient{
		client: client,
	}
}

// PutItem puts an item into DynamoDB
func (d *DynamoDBClient) PutItem(ctx context.Context, tableName string, item map[string]types.AttributeValue) error {
	_, err := d.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

// DeleteItem deletes an item from DynamoDB
func (d *DynamoDBClient) DeleteItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) error {
	_, err := d.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	return err
}

// GetItem gets an item from DynamoDB
func (d *DynamoDBClient) GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	if err != nil {
		return nil, err
	}
	return result.Item, nil
}

// QueryItems queries items from DynamoDB with support for single-table design
func (d *DynamoDBClient) QueryItems(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	// Make a copy of the input to avoid modifying the original
	queryInput := &dynamodb.QueryInput{
		TableName:                 input.TableName,
		IndexName:                 input.IndexName,
		KeyConditionExpression:    input.KeyConditionExpression,
		FilterExpression:          input.FilterExpression,
		ExpressionAttributeNames:  input.ExpressionAttributeNames,
		ExpressionAttributeValues: input.ExpressionAttributeValues,
		ExclusiveStartKey:         input.ExclusiveStartKey,
		Limit:                     input.Limit,
		ScanIndexForward:          input.ScanIndexForward,
		ReturnConsumedCapacity:    input.ReturnConsumedCapacity,
		Select:                    input.Select,
	}

	// Execute the query
	result, err := d.client.Query(ctx, queryInput)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Scan scans items from a DynamoDB table with support for single-table design
func (d *DynamoDBClient) Scan(ctx context.Context, input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	// Make a copy of the input to avoid modifying the original
	scanInput := &dynamodb.ScanInput{
		TableName:                 input.TableName,
		IndexName:                 input.IndexName,
		FilterExpression:          input.FilterExpression,
		ExpressionAttributeNames:  input.ExpressionAttributeNames,
		ExpressionAttributeValues: input.ExpressionAttributeValues,
		ExclusiveStartKey:         input.ExclusiveStartKey,
		Limit:                     input.Limit,
		ReturnConsumedCapacity:    input.ReturnConsumedCapacity,
		Select:                    input.Select,
	}

	// Execute the scan
	result, err := d.client.Scan(ctx, scanInput)
	if err != nil {
		return nil, err
	}

	return result, nil
}
