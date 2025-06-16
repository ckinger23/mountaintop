package db

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// InMemoryClient implements DatabaseClient with an in-memory store
type InMemoryClient struct {
	mu     sync.RWMutex
	tables map[string]map[string]map[string]types.AttributeValue // tableName -> itemKey -> item
}

// NewInMemoryClient creates a new in-memory database client
func NewInMemoryClient() *InMemoryClient {
	return &InMemoryClient{
		tables: make(map[string]map[string]map[string]types.AttributeValue),
	}
}

// PutItem stores an item in memory
func (i *InMemoryClient) PutItem(ctx context.Context, tableName string, item map[string]types.AttributeValue) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, exists := i.tables[tableName]; !exists {
		i.tables[tableName] = make(map[string]map[string]types.AttributeValue)
	}

	// For simplicity, assuming the key is a single field named "id"
	key := item["id"].(*types.AttributeValueMemberS).Value
	i.tables[tableName][key] = item
	return nil
}

// GetItem retrieves an item from memory
func (i *InMemoryClient) GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	table, exists := i.tables[tableName]
	if !exists {
		return nil, nil
	}

	// For simplicity, assuming the key is a single field named "id"
	keyValue := key["id"].(*types.AttributeValueMemberS).Value
	item, exists := table[keyValue]
	if !exists {
		return nil, nil
	}

	// Return a copy of the item to prevent external modifications
	itemCopy := make(map[string]types.AttributeValue, len(item))
	for k, v := range item {
		itemCopy[k] = v
	}
	return itemCopy, nil
}

// QueryItems is a basic implementation that would need to be enhanced based on your query needs
func (i *InMemoryClient) QueryItems(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	// This is a basic implementation that returns all items in the table
	// You would need to implement the actual query logic based on the input parameters
	return &dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{},
	}, nil
}

// Scan returns all items in the specified table
func (i *InMemoryClient) Scan(ctx context.Context, input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	tableName := aws.ToString(input.TableName)
	table, exists := i.tables[tableName]
	if !exists {
		return &dynamodb.ScanOutput{
			Items: []map[string]types.AttributeValue{},
		}, nil
	}

	// Convert the map of items to a slice
	items := make([]map[string]types.AttributeValue, 0, len(table))
	for _, item := range table {
		items = append(items, item)
	}

	return &dynamodb.ScanOutput{
		Items: items,
	}, nil
}
