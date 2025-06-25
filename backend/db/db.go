package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
)

// DatabaseClient defines the interface for database operations
type DatabaseClient interface {
	PutItem(ctx context.Context, tableName string, item map[string]types.AttributeValue) error
	GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error)
	DeleteItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) error
	QueryItems(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error)
	Scan(ctx context.Context, input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error)
}

type resolverV2 struct{}

func (*resolverV2) ResolveEndpoint(ctx context.Context, params dynamodb.EndpointParameters) (smithyendpoints.Endpoint, error) {
	fmt.Printf("The endpoint provided in config is %s\n", *params.Endpoint)
	return dynamodb.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}

// NewDBClient creates a new database client based on the environment
func NewDBClient(cfg aws.Config) DatabaseClient {
	// Use LocalStack for local development
	if os.Getenv("ENV") == "local" {
		client := dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
			options.BaseEndpoint = aws.String("http://localhost:4566")
			options.EndpointResolverV2 = &resolverV2{}
		})
		return NewDynamoDBClientFromClient(client)
	}

	log.Println("Using AWS DynamoDB for production")
	return NewDynamoDBClientFromConfig(cfg)
}
