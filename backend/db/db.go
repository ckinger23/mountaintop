package db

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DBClient wraps DynamoDB client and provides helper methods
type DBClient struct {
	Client *dynamodb.Client
}

// NewDBClient creates a new DBClient
func NewDBClient(cfg aws.Config) *DBClient {
	return &DBClient{
		Client: dynamodb.NewFromConfig(cfg),
	}
}

// PutItem puts an item into DynamoDB
func (d *DBClient) PutItem(ctx context.Context, tableName string, item map[string]types.AttributeValue) error {
	_, err := d.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

// GetItem gets an item from DynamoDB
func (d *DBClient) GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	result, err := d.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	})
	if err != nil {
		return nil, err
	}
	return result.Item, nil
}

// QueryItems queries items from DynamoDB
func (d *DBClient) QueryItems(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	return d.Client.Query(ctx, input)
}
