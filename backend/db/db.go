package db

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DatabaseClient defines the interface for database operations
type DatabaseClient interface {
	PutItem(ctx context.Context, tableName string, item map[string]types.AttributeValue) error
	GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error)
	QueryItems(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
}

// NewDBClient creates a new database client based on the environment
func NewDBClient(cfg aws.Config) DatabaseClient {
	// Use LocalStack for local development
	if os.Getenv("ENV") == "local" {
		// Configure AWS SDK to use LocalStack
		localCfg := cfg.Copy()
		localCfg.EndpointResolver = aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				PartitionID:       "aws",
				URL:               "http://localhost:4566",
				SigningRegion:     region,
				HostnameImmutable: true,
			}, nil
		})

		log.Println("Using LocalStack DynamoDB for local development")
		return NewDynamoDBClient(localCfg)
	}

	log.Println("Using AWS DynamoDB for production")
	return NewDynamoDBClient(cfg)
}
