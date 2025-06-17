# Database Design Documentation

## Overview
This document describes the database schema and access patterns for the Football Picking League application. The application uses DynamoDB with a single-table design pattern for efficient data access and scalability.

## Table Structure

### Primary Key Structure
- **PK (Partition Key)**: Composite key with pattern `ENTITY#<entity_type>#<id>`
- **SK (Sort Key)**: Used for range operations and GSI overloading

### Global Secondary Indexes (GSI)

#### GSI-EntityType
- **GSI-PK**: `entity_type` (e.g., "TEAM", "USER", "GAME")
- **GSI-SK**: `id`
- **Purpose**: Query entities by their type

#### GSI-UserPicks
- **GSI-PK**: `user_id`
- **GSI-SK**: `week`
- **Purpose**: Get all picks for a specific user and week

#### GSI-GamePicks
- **GSI-PK**: `game_id`
- **GSI-SK**: `pick` (home/away)
- **Purpose**: Get all picks for a specific game

## Data Models

### Team
```
PK: ENTITY#TEAM#{teamId}
SK: META
entity_type: "TEAM"
id: string
name: string
conference_id: string
wins: number
losses: number
created_at: timestamp
updated_at: timestamp
```

### User
```
PK: ENTITY#USER#{userId}
SK: META
entity_type: "USER"
id: string
username: string
email: string
role: string (admin/player)
created_at: timestamp
updated_at: timestamp
```

### Game
```
PK: ENTITY#GAME#{gameId}
SK: META
entity_type: "GAME"
id: string
league_id: string
week: number
home_team_id: string
away_team_id: string
game_date: timestamp
status: string (pending/in_progress/completed)
winner: string (home/away)
created_at: timestamp
updated_at: timestamp
```

### Pick
```
PK: ENTITY#PICK#{pickId}
SK: USER#{userId}#GAME#{gameId}
entity_type: "PICK"
id: string
user_id: string
game_id: string
week: number
pick: string (home/away)
status: string (pending/correct/incorrect)
points: number
created_at: timestamp
updated_at: timestamp
```

## Access Patterns

1. **Get Team by ID**
   - Query PK = `ENTITY#TEAM#{teamId}`, SK = `META`

2. **List All Teams**
   - Query GSI-EntityType where entity_type = "TEAM"

3. **Get User's Picks for Week**
   - Query GSI-UserPicks where user_id = X and begins_with(week, Y)

4. **Get All Picks for Game**
   - Query GSI-GamePicks where game_id = X

5. **Get Game Details**
   - Query PK = `ENTITY#GAME#{gameId}`, SK = `META`

## Environment Variables

- `DYNAMODB_TABLE`: Name of the DynamoDB table (default: "football-picking-league")
- `ENV`: Environment ("local" for LocalStack, otherwise uses AWS)

## Local Development

For local development, the application uses an in-memory implementation of the DynamoDB client that mimics the behavior of the real DynamoDB service. This is automatically selected when `ENV=local`.
