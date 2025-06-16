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
	Scan(ctx context.Context, input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

// NewDBClient creates a new database client based on the environment
func NewDBClient(cfg aws.Config) DatabaseClient {
	// Use LocalStack for local development
	if os.Getenv("ENV") == "local" {
		// Create a custom endpoint resolver for LocalStack
		customResolver := dynamodb.NewDefaultEndpointResolverV2()
		_, err := customResolver.ResolveEndpoint(context.TODO(), dynamodb.EndpointParameters{
			Region: aws.String("us-east-1"),
		})
		if err != nil {
			log.Fatalf("unable to resolve endpoint, %v", err)
		}

		// Create a copy of the config and set the custom endpoint
		localCfg := cfg.Copy()
		localCfg.BaseEndpoint = aws.String("http://localhost:4566")
		localCfg.Region = "us-east-1"

		log.Println("Using LocalStack DynamoDB for local development")
		return NewDynamoDBClient(localCfg)
	}

	log.Println("Using AWS DynamoDB for production")
	return NewDynamoDBClient(cfg)
}
