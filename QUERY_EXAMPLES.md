# DynamoDB Query Examples for Football Picking League

This document provides specific query examples for each required capability using the optimized single-table design.

## Table Structure Overview

**Main Table**: `FootballLeague`
- **PK** (Partition Key): String
- **SK** (Sort Key): String

**Global Secondary Indexes**:
- **GSI1-EntityLookup**: GSI1_PK (hash) + GSI1_SK (range)
- **GSI2-LeagueWeek**: GSI2_PK (hash) + GSI2_SK (range)  
- **GSI3-UserQueries**: GSI3_PK (hash) + GSI3_SK (range)
- **GSI4-GameQueries**: GSI4_PK (hash) + GSI4_SK (range)

## User Queries

### 1. Get User by ID
```go
key := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
    "SK": &types.AttributeValueMemberS{Value: "PROFILE"},
}
result, err := dbClient.GetItem(ctx, tableName, key)
```

### 2. Get User by Username
```go
// Step 1: Look up user ID from username
lookupKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "USERNAME#" + username},
    "SK": &types.AttributeValueMemberS{Value: "LOOKUP"},
}
lookupResult, err := dbClient.GetItem(ctx, tableName, lookupKey)

// Step 2: Get user data using the found user ID
userKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
    "SK": &types.AttributeValueMemberS{Value: "PROFILE"},
}
userResult, err := dbClient.GetItem(ctx, tableName, userKey)
```

### 3. Get User by Email
```go
// Step 1: Look up user ID from email
lookupKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "EMAIL#" + email},
    "SK": &types.AttributeValueMemberS{Value: "LOOKUP"},
}
lookupResult, err := dbClient.GetItem(ctx, tableName, lookupKey)

// Step 2: Get user data using the found user ID
userKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
    "SK": &types.AttributeValueMemberS{Value: "PROFILE"},
}
userResult, err := dbClient.GetItem(ctx, tableName, userKey)
```

### 4. Get All Users
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI1-EntityLookup"),
    KeyConditionExpression: aws.String("GSI1_PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "USER"},
    },
}
result, err := dbClient.Query(ctx, input)
```

## Conference Queries

### 1. Get Conference by ID
```go
key := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "CONFERENCE#" + conferenceID},
    "SK": &types.AttributeValueMemberS{Value: "METADATA"},
}
result, err := dbClient.GetItem(ctx, tableName, key)
```

### 2. Get Conference by Name
```go
// Step 1: Look up conference ID from name
lookupKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "CONFERENCE_NAME#" + name},
    "SK": &types.AttributeValueMemberS{Value: "LOOKUP"},
}
lookupResult, err := dbClient.GetItem(ctx, tableName, lookupKey)

// Step 2: Get conference data using the found conference ID
conferenceKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "CONFERENCE#" + conferenceID},
    "SK": &types.AttributeValueMemberS{Value: "METADATA"},
}
conferenceResult, err := dbClient.GetItem(ctx, tableName, conferenceKey)
```

### 3. Get All Conferences
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI1-EntityLookup"),
    KeyConditionExpression: aws.String("GSI1_PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "CONFERENCE"},
    },
}
result, err := dbClient.Query(ctx, input)
```

## Team Queries

### 1. Get All Teams
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI1-EntityLookup"),
    KeyConditionExpression: aws.String("GSI1_PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "TEAM"},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 2. Get All Teams in a Specific Conference
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "CONFERENCE#" + conferenceID},
        ":sk": &types.AttributeValueMemberS{Value: "TEAM#"},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 3. Get Team by ID
```go
key := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
    "SK": &types.AttributeValueMemberS{Value: "METADATA"},
}
result, err := dbClient.GetItem(ctx, tableName, key)
```

### 4. Get Team by Name
```go
// Step 1: Look up team ID from name
lookupKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "TEAM_NAME#" + name},
    "SK": &types.AttributeValueMemberS{Value: "LOOKUP"},
}
lookupResult, err := dbClient.GetItem(ctx, tableName, lookupKey)

// Step 2: Get team data using the found team ID
teamKey := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "TEAM#" + teamID},
    "SK": &types.AttributeValueMemberS{Value: "METADATA"},
}
teamResult, err := dbClient.GetItem(ctx, tableName, teamKey)
```

## League Queries

### 1. Get League by ID
```go
key := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "LEAGUE#" + leagueID},
    "SK": &types.AttributeValueMemberS{Value: "METADATA"},
}
result, err := dbClient.GetItem(ctx, tableName, key)
```

### 2. Get All Leagues
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI1-EntityLookup"),
    KeyConditionExpression: aws.String("GSI1_PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "LEAGUE"},
    },
}
result, err := dbClient.Query(ctx, input)
```

## Game Queries

### 1. Get All Games for a Week
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI2-LeagueWeek"),
    KeyConditionExpression: aws.String("GSI2_PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "LEAGUE#" + leagueID + "#WEEK#" + weekNum},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 2. Get All Games for a Date
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    KeyConditionExpression: aws.String("PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "DATE#" + date + "#LEAGUE#" + leagueID},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 3. Get Game by ID
```go
key := map[string]types.AttributeValue{
    "PK": &types.AttributeValueMemberS{Value: "GAME#" + gameID},
    "SK": &types.AttributeValueMemberS{Value: "METADATA"},
}
result, err := dbClient.GetItem(ctx, tableName, key)
```

## Pick Queries

### 1. Get All Picks for a League and Season
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI1-EntityLookup"),
    KeyConditionExpression: aws.String("GSI1_PK = :pk AND begins_with(GSI1_SK, :sk)"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "PICK"},
        ":sk": &types.AttributeValueMemberS{Value: leagueID + "#" + season},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 2. Get All Picks for a League and Season by Week
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI2-LeagueWeek"),
    KeyConditionExpression: aws.String("GSI2_PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "LEAGUE#" + leagueID + "#WEEK#" + weekNum},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 3. Get All Picks for a League, Season, and User
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    KeyConditionExpression: aws.String("PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "PICK#" + leagueID + "#" + season + "#" + userID},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 4. Get All Picks for a League, Season, and User by Week
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI3-UserQueries"),
    KeyConditionExpression: aws.String("GSI3_PK = :pk AND begins_with(GSI3_SK, :sk)"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "USER#" + userID + "#" + season},
        ":sk": &types.AttributeValueMemberS{Value: "WEEK#" + weekNum},
    },
}
result, err := dbClient.Query(ctx, input)
```

### 5. Get All Picks for a Specific Game
```go
input := &dynamodb.QueryInput{
    TableName: aws.String(tableName),
    IndexName: aws.String("GSI4-GameQueries"),
    KeyConditionExpression: aws.String("GSI4_PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "GAME#" + gameID},
    },
}
result, err := dbClient.Query(ctx, input)
```

## Leaderboard Queries (Computed from Picks)

### 1. Calculate Season Leaderboard
```go
// Get all picks for the league and season
picks := getAllPicksForLeagueAndSeason(leagueID, season)

// Group by user and calculate scores
userScores := make(map[string]int)
for _, pick := range picks {
    userScores[pick.UserID] += pick.Points
}

// Sort and return leaderboard
```

### 2. Calculate Weekly Leaderboard
```go
// Get all picks for the specific week
picks := getAllPicksForLeagueSeasonAndWeek(leagueID, season, week)

// Group by user and calculate weekly scores
userScores := make(map[string]int)
for _, pick := range picks {
    userScores[pick.UserID] += pick.Points
}

// Sort and return weekly leaderboard
```

## CRUD Operations

### Create Operations
Each entity type requires:
1. Main entity record with proper PK/SK
2. GSI attributes for efficient queries
3. Lookup records for alternate key access (username, email, team name, conference name)

### Update Operations
1. Update main entity record
2. Update lookup records if alternate keys change
3. Use conditional updates to prevent race conditions

### Delete Operations
1. Delete main entity record
2. Delete associated lookup records
3. Consider cascade deletes for related data

## Performance Notes

- **GetItem** operations: O(1) - fastest for direct key access
- **Query** operations: Efficient for range queries using GSIs
- **Scan** operations: Avoid - we've eliminated all scan operations with this design
- **Hot partitions**: Distribute load across different PK patterns
- **Batch operations**: Use BatchGetItem and BatchWriteItem for multiple operations