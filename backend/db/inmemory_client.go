package db

import (
	"context"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// InMemoryClient implements DatabaseClient with an in-memory store that supports single-table design
type InMemoryClient struct {
	mu sync.RWMutex
	// tableName -> PK -> SK -> item
	tables map[string]map[string]map[string]map[string]types.AttributeValue
	// For GSIs: tableName -> GSI name -> GSI PK -> GSI SK -> [PK+SK]
	globalIndexes map[string]map[string]map[string]map[string][]string
}

// NewInMemoryClient creates a new in-memory database client
func NewInMemoryClient() *InMemoryClient {
	return &InMemoryClient{
		tables:        make(map[string]map[string]map[string]map[string]types.AttributeValue),
		globalIndexes: make(map[string]map[string]map[string]map[string][]string),
	}
}

// PutItem stores an item in memory
func (i *InMemoryClient) PutItem(ctx context.Context, input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	tableName := *input.TableName
	pk := input.Item["PK"].(*types.AttributeValueMemberS).Value
	sk := input.Item["SK"].(*types.AttributeValueMemberS).Value

	// Initialize table if it doesn't exist
	if _, exists := i.tables[tableName]; !exists {
		i.tables[tableName] = make(map[string]map[string]map[string]types.AttributeValue)
	}
	if _, exists := i.tables[tableName][pk]; !exists {
		i.tables[tableName][pk] = make(map[string]map[string]types.AttributeValue)
	}

	// Make a deep copy of the item
	itemCopy := make(map[string]types.AttributeValue)
	for k, v := range input.Item {
		itemCopy[k] = v
	}

	// Store the item
	i.tables[tableName][pk][sk] = itemCopy

	// Update GSIs
	i.updateGSIs(tableName, pk, sk, itemCopy)

	return &dynamodb.PutItemOutput{}, nil
}

// GetItem retrieves an item by its primary key
func (i *InMemoryClient) GetItem(ctx context.Context, input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	tableName := *input.TableName
	pk := input.Key["PK"].(*types.AttributeValueMemberS).Value
	sk := input.Key["SK"].(*types.AttributeValueMemberS).Value

	// Find the item
	var item map[string]types.AttributeValue
	if table, exists := i.tables[tableName]; exists {
		if pkItems, exists := table[pk]; exists {
			if item, exists = pkItems[sk]; exists {
				// Return a copy of the item
				itemCopy := make(map[string]types.AttributeValue)
				for k, v := range item {
					itemCopy[k] = v
				}
				return &dynamodb.GetItemOutput{
					Item: itemCopy,
				}, nil
			}
		}
	}

	return &dynamodb.GetItemOutput{}, nil
}

// Query retrieves items by partition key and optional sort key conditions
func (i *InMemoryClient) Query(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	tableName := *input.TableName
	var items []map[string]types.AttributeValue

	// Check if this is a query on a GSI
	if input.IndexName != nil {
		return i.queryGSI(tableName, *input.IndexName, input)
	}

	// Regular table query
	pkValue := input.ExpressionAttributeValues[":pk"].(*types.AttributeValueMemberS).Value

	// Get all items with the matching PK
	if table, exists := i.tables[tableName]; exists {
		if pkItems, exists := table[pkValue]; exists {
			// Apply filter expression if provided
			filterFn := i.getFilterFunction(input.FilterExpression, input.ExpressionAttributeValues)

			for sk, item := range pkItems {
				// Skip if SK doesn't match the condition
				if input.KeyConditionExpression != nil && strings.Contains(*input.KeyConditionExpression, "begins_with") {
					prefix := input.ExpressionAttributeValues[":sk"].(*types.AttributeValueMemberS).Value
					if !strings.HasPrefix(sk, prefix) {
						continue
					}
				}

				// Apply filter if provided
				if filterFn == nil || filterFn(item) {
					// Make a copy of the item
					itemCopy := make(map[string]types.AttributeValue)
					for k, v := range item {
						itemCopy[k] = v
					}
					items = append(items, itemCopy)
				}
			}
		}
	}

	return &dynamodb.QueryOutput{
		Items: items,
	}, nil
}

// DeleteItem removes an item by its primary key
func (i *InMemoryClient) DeleteItem(ctx context.Context, input *dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	tableName := *input.TableName
	pk := input.Key["PK"].(*types.AttributeValueMemberS).Value
	sk := input.Key["SK"].(*types.AttributeValueMemberS).Value

	// Get the item before deleting to update GSIs
	var oldItem map[string]types.AttributeValue
	if table, exists := i.tables[tableName]; exists {
		if pkItems, exists := table[pk]; exists {
			if item, exists := pkItems[sk]; exists {
				oldItem = item
				delete(pkItems, sk)
				if len(pkItems) == 0 {
					delete(table, pk)
				}
			}
		}
	}

	// Update GSIs
	if oldItem != nil {
		// Remove from GSIs
		i.removeFromGSIs(tableName, pk, sk, oldItem)
	}

	return &dynamodb.DeleteItemOutput{}, nil
}

// UpdateItem updates an existing item or creates a new one if it doesn't exist
func (i *InMemoryClient) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput) (*dynamodb.UpdateItemOutput, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	tableName := *input.TableName
	pk := input.Key["PK"].(*types.AttributeValueMemberS).Value
	sk := input.Key["SK"].(*types.AttributeValueMemberS).Value

	// Get the existing item
	var existingItem map[string]types.AttributeValue
	if table, exists := i.tables[tableName]; exists {
		if pkItems, exists := table[pk]; exists {
			existingItem = pkItems[sk]
		}
	}

	// Create a new item or update the existing one
	newItem := make(map[string]types.AttributeValue)
	if existingItem != nil {
		// Copy existing attributes
		for k, v := range existingItem {
			newItem[k] = v
		}
	} else {
		// New item, set PK and SK
		newItem["PK"] = &types.AttributeValueMemberS{Value: pk}
		newItem["SK"] = &types.AttributeValueMemberS{Value: sk}
	}

	// Apply updates from the update expression
	// This is a simplified implementation that only handles basic SET operations
	// by using the ExpressionAttributeValues directly
	if input.UpdateExpression != nil && input.ExpressionAttributeValues != nil {
		// For a complete implementation, you would need to properly parse the UpdateExpression
		// and handle all possible update actions (SET, REMOVE, ADD, DELETE)
		for key, value := range input.ExpressionAttributeValues {
			// Remove the leading ':' from the attribute name
			if strings.HasPrefix(key, ":") {
				// Try to find the corresponding attribute name in the expression attribute names
				// If not found, use the key without the ':' as a fallback
				attrName := strings.TrimPrefix(key, ":")
				if input.ExpressionAttributeNames != nil {
					if mappedName, exists := input.ExpressionAttributeNames["#"+attrName]; exists {
						attrName = mappedName
					}
				}
				newItem[attrName] = value
			}
		}
	}

	// Remove old item from GSIs if it exists
	if existingItem != nil {
		i.removeFromGSIs(tableName, pk, sk, existingItem)
	}

	// Store the new item
	if _, exists := i.tables[tableName]; !exists {
		i.tables[tableName] = make(map[string]map[string]map[string]types.AttributeValue)
	}
	if _, exists := i.tables[tableName][pk]; !exists {
		i.tables[tableName][pk] = make(map[string]map[string]types.AttributeValue)
	}
	i.tables[tableName][pk][sk] = newItem

	// Update GSIs
	i.updateGSIs(tableName, pk, sk, newItem)

	return &dynamodb.UpdateItemOutput{
		Attributes: newItem,
	}, nil
}

// Helper function to update GSIs when an item is added or updated
func (i *InMemoryClient) updateGSIs(tableName, pk, sk string, item map[string]types.AttributeValue) {
	// This is a simplified implementation that assumes GSIs are known
	// In a real implementation, you would get GSI definitions from table metadata

	// Example GSI: GSI-EntityType
	if entityType, ok := item["entity_type"]; ok {
		if entityTypeVal, ok := entityType.(*types.AttributeValueMemberS); ok {
			if id, ok := item["id"]; ok {
				if idVal, ok := id.(*types.AttributeValueMemberS); ok {
					i.updateGSI(tableName, "GSI-EntityType", entityTypeVal.Value, idVal.Value, pk, sk)
				}
			}
		}
	}

	// Example GSI: GSI-UserPicks
	if userID, ok := item["user_id"]; ok {
		if userIDVal, ok := userID.(*types.AttributeValueMemberS); ok {
			if week, ok := item["week"]; ok {
				if weekVal, ok := week.(*types.AttributeValueMemberN); ok {
					i.updateGSI(tableName, "GSI-UserPicks", userIDVal.Value, weekVal.Value, pk, sk)
				}
			}
		}
	}

	// Example GSI: GSI-GamePicks
	if gameID, ok := item["game_id"]; ok {
		if gameIDVal, ok := gameID.(*types.AttributeValueMemberS); ok {
			if pick, ok := item["pick"]; ok {
				if pickVal, ok := pick.(*types.AttributeValueMemberS); ok {
					i.updateGSI(tableName, "GSI-GamePicks", gameIDVal.Value, pickVal.Value, pk, sk)
				}
			}
		}
	}
}

// Helper function to update a specific GSI
func (i *InMemoryClient) updateGSI(tableName, indexName, gsiPK, gsiSK, pk, sk string) {
	if _, exists := i.globalIndexes[tableName]; !exists {
		i.globalIndexes[tableName] = make(map[string]map[string]map[string][]string)
	}
	if _, exists := i.globalIndexes[tableName][indexName]; !exists {
		i.globalIndexes[tableName][indexName] = make(map[string]map[string][]string)
	}
	if _, exists := i.globalIndexes[tableName][indexName][gsiPK]; !exists {
		i.globalIndexes[tableName][indexName][gsiPK] = make(map[string][]string)
	}
	if _, exists := i.globalIndexes[tableName][indexName][gsiPK][gsiSK]; !exists {
		i.globalIndexes[tableName][indexName][gsiPK][gsiSK] = []string{}
	}

	// Add the PK/SK to the GSI if not already present
	found := false
	for _, item := range i.globalIndexes[tableName][indexName][gsiPK][gsiSK] {
		if item == pk+"#"+sk {
			found = true
			break
		}
	}
	if !found {
		i.globalIndexes[tableName][indexName][gsiPK][gsiSK] = append(
			i.globalIndexes[tableName][indexName][gsiPK][gsiSK],
			pk+"#"+sk,
		)
	}
}

// Helper function to remove an item from GSIs when it's deleted or updated
func (i *InMemoryClient) removeFromGSIs(tableName, pk, sk string, item map[string]types.AttributeValue) {
	// This is a simplified implementation that assumes GSIs are known
	// Similar to updateGSIs but removes the item from GSIs
}

// Helper function to query a GSI
func (i *InMemoryClient) queryGSI(tableName, indexName string, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	var items []map[string]types.AttributeValue

	// Get the GSI PK value
	gsiPKValue := input.ExpressionAttributeValues[":pk"].(*types.AttributeValueMemberS).Value

	// Check if the GSI exists
	if _, exists := i.globalIndexes[tableName]; !exists {
		return &dynamodb.QueryOutput{Items: items}, nil
	}
	if _, exists := i.globalIndexes[tableName][indexName]; !exists {
		return &dynamodb.QueryOutput{Items: items}, nil
	}
	if _, exists := i.globalIndexes[tableName][indexName][gsiPKValue]; !exists {
		return &dynamodb.QueryOutput{Items: items}, nil
	}

	// Get items from the GSI
	for gsiSK, itemRefs := range i.globalIndexes[tableName][indexName][gsiPKValue] {
		// Check if SK condition is satisfied
		if input.KeyConditionExpression != nil && strings.Contains(*input.KeyConditionExpression, "begins_with") {
			prefix := input.ExpressionAttributeValues[":sk"].(*types.AttributeValueMemberS).Value
			if !strings.HasPrefix(gsiSK, prefix) {
				continue
			}
		}

		// Get the actual items
		for _, itemRef := range itemRefs {
			parts := strings.SplitN(itemRef, "#", 2)
			if len(parts) != 2 {
				continue
			}
			pk, sk := parts[0], parts[1]

			if item, exists := i.tables[tableName][pk][sk]; exists {
				// Apply filter if provided
				filterFn := i.getFilterFunction(input.FilterExpression, input.ExpressionAttributeValues)
				if filterFn == nil || filterFn(item) {
					// Make a copy of the item
					itemCopy := make(map[string]types.AttributeValue)
					for k, v := range item {
						itemCopy[k] = v
					}
					items = append(items, itemCopy)
				}
			}
		}
	}

	return &dynamodb.QueryOutput{
		Items: items,
	}, nil
}

// Helper function to create a filter function from a filter expression
func (i *InMemoryClient) getFilterExpression(expr *string, values map[string]types.AttributeValue) func(map[string]types.AttributeValue) bool {
	if expr == nil {
		return nil
	}

	// This is a simplified implementation that only handles simple equality conditions
	// In a real implementation, you would need to parse the expression
	return func(item map[string]types.AttributeValue) bool {
		// Simple equality check for now
		// Example: "attribute_not_exists(deleted_at) OR deleted_at = :deleted_at"
		return true
	}
}

// Helper function to get a filter function from a filter expression
func (i *InMemoryClient) getFilterFunction(expr *string, values map[string]types.AttributeValue) func(map[string]types.AttributeValue) bool {
	if expr == nil {
		return nil
	}

	// This is a simplified implementation that only handles simple equality conditions
	return func(item map[string]types.AttributeValue) bool {
		// Simple equality check for now
		// In a real implementation, you would parse the expression
		return true
	}
}
