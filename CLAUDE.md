# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Quick Start
```bash
# Start all services with Docker (recommended)
docker compose up --build

# Access points:
# - Frontend: http://localhost:3000
# - Backend API: http://localhost:8080  
# - LocalStack: http://localhost:4566
```

### Manual Development
```bash
# Backend
cd backend && go run main.go

# Frontend
cd frontend && npm install && npm run dev

# Infrastructure setup
terraform init && terraform apply -auto-approve
```

### Testing
```bash
# Backend tests
make test                    # Or: go test ./... -v
cd backend && go test ./...  # Run from backend directory

# Frontend linting
cd frontend && npm run lint

# Frontend build
cd frontend && npm run build
```

### Docker Management
```bash
make localstack-up          # Start LocalStack
make localstack-down        # Stop LocalStack
make init-localstack        # Initialize with sample data
docker compose logs -f      # View all logs
docker compose logs -f backend  # View specific service logs
```

## Architecture Overview

This is a **full-stack football picking league application** with:

### Backend (Go)
- **Framework**: Gorilla Mux REST API server (port 8080)
- **Database**: Single-table DynamoDB design with LocalStack for development
- **Testing**: Comprehensive handler tests using testify with mocks
- **Key Pattern**: Interface-based design with `DatabaseClient` interface for testability

### Frontend (React + TypeScript)
- **Framework**: React 19 + TypeScript with Vite build system
- **Development**: Hot reload on port 3000 with proxy to backend
- **Status**: Basic structure exists but components need implementation

### Database Design
**Single-table DynamoDB** (`FootballLeague` table) with sophisticated access patterns:

**Entity Structure**: 
- PK format: `ENTITY#<type>#<id>`
- SK: Used for range operations and metadata
- Entity types: USER, TEAM, CONFERENCE, GAME, PICK, LEADERBOARD_ENTRY

**Global Secondary Indexes**:
- `GSI-EntityType`: Query by entity type
- `GSI-UserPicks`: Get user picks by week  
- `GSI-GamePicks`: Get all picks for specific game

### Sample Data Available
- 5 test users: Carter, Paul, Cal, Nathan, Nick
- 4 conferences: SEC, Big Ten, Big 12, ACC
- 16 teams across conferences
- 15 games spanning 5 weeks with complete pick data

## Key Files and Patterns

### Backend Structure
- `main.go`: Server entry point with routing
- `handlers/`: CRUD operations for each entity type
- `database/`: DynamoDB client interface and implementation
- `models/`: Data structures and types
- All handlers have corresponding `*_test.go` files with mocks

### Testing Approach
- **Mock Strategy**: `MockDBClient` interface for isolating database operations
- **Test Structure**: Table-driven tests with comprehensive setup/teardown
- **Coverage**: All major handlers have complete test suites
- Tests use dependency injection pattern for database clients

### Environment Configuration
- **Local Development**: Uses LocalStack with `ENV=local`
- **Test Credentials**: `AWS_ACCESS_KEY_ID=test`, `AWS_SECRET_ACCESS_KEY=test`
- **DynamoDB Endpoint**: `http://localstack:4566` (in containers) or `http://localhost:4566`

## Development Workflow

1. **Infrastructure**: Use `docker compose up --build` to start all services
2. **Database**: Terraform automatically creates tables and test data
3. **Testing**: Run `make test` before committing changes
4. **Frontend Development**: Components exist but need implementation to match backend API
5. **Data Persistence**: LocalStack data persists between restarts (use `docker compose down -v` to reset)

## Important Notes

- **Single Table Design**: All entities stored in one DynamoDB table with composite keys
- **Interface Pattern**: Database operations go through `DatabaseClient` interface for testability
- **Environment Switching**: `ENV` variable switches between LocalStack (local) and AWS (production)
- **Mock Testing**: All handlers tested with `MockDBClient` for fast, isolated tests
- **Infrastructure as Code**: Terraform manages DynamoDB schema and test data consistently