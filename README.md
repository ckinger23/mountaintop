# Football Picking League

A web application for managing and playing in football game prediction leagues.

## Features

- League management system
- Weekly game selection by league admins
- Player predictions submission
- Real-time scoring and leaderboards
- Season statistics tracking

## Tech Stack

- **Backend**: Go
- **Frontend**: React
- **Database**: AWS DynamoDB (LocalStack for local development)
- **Infrastructure**: Docker, AWS (production)

## Prerequisites

- [Docker](https://www.docker.com/products/docker-desktop/) (recommended)
- [Node.js](https://nodejs.org/) 18+ (for frontend development)
- [Go](https://golang.org/dl/) 1.20+ (for backend development)
- [AWS CLI](https://aws.amazon.com/cli/) (for production deployment)

## Local Development with Docker

### Quick Start

The easiest way to get started is using Docker Compose, which will set up all required services:

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/yourusername/football-picking-league.git
cd football-picking-league

# Copy the example environment file
cp backend/.env.example backend/.env

# Start all services (LocalStack, Backend, and Frontend)
docker compose up --build
```

This will start:
- LocalStack (DynamoDB) on port 4566
- Backend server on port 8080
- Frontend development server on port 3000

Access the application at: http://localhost:3000

### Managing Persistence

By default, LocalStack data is persisted between container restarts. Here's how to manage it:

- **Normal shutdown** (keeps data):
  ```bash
  docker compose down
  ```

- **Full cleanup** (deletes all data):
  ```bash
  docker compose down -v
  ```

- **View volumes**:
  ```bash
  docker volume ls
  ```

- **Inspect a volume**:
  ```bash
  docker volume inspect football-picking-league_localstack_data
  ```

### Common Docker Commands

- **Start services in detached mode**:
  ```bash
  docker compose up -d
  ```

- **View logs**:
  ```bash
  # All services
  docker compose logs -f
  
  # Specific service
  docker compose logs -f backend
  ```

- **Rebuild containers**:
  ```bash
  docker compose up --build
  ```

- **Access container shell**:
  ```bash
  docker compose exec backend sh
  ```

### Working with LocalStack

Access LocalStack's web interface at: http://localhost:4566

#### AWS CLI Examples

First, set up your environment:

```bash
# Set AWS credentials for LocalStack
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=us-east-1
```

#### DynamoDB Examples

**1. List all tables**
```bash
aws --endpoint-url=http://localhost:4566 dynamodb list-tables
```

**2. Create a new table**
```bash
aws --endpoint-url=http://localhost:4566 dynamodb create-table \
    --table-name Users \
    --attribute-definitions AttributeName=UserID,AttributeType=S \
    --key-schema AttributeName=UserID,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST
```

**3. Add an item**
```bash
aws --endpoint-url=http://localhost:4566 dynamodb put-item \
    --table-name Users \
    --item '{
        "UserID": {"S": "user1"}, 
        "Name": {"S": "John Doe"},
        "Email": {"S": "john@example.com"},
        "Score": {"N": "100"}
    }'
```

**4. Get an item**
```bash
aws --endpoint-url=http://localhost:4566 dynamodb get-item \
    --table-name Users \
    --key '{"UserID": {"S": "user1"}}'
```

**5. Scan a table**
```bash
aws --endpoint-url=http://localhost:4566 dynamodb scan \
    --table-name Users
```

**6. Query items**
```bash
# Example: Find users with score > 50
aws --endpoint-url=http://localhost:4566 dynamodb scan \
    --table-name Users \
    --filter-expression "Score > :score" \
    --expression-attribute-values '{":score": {"N": "50"}}'
```

#### Using cURL

You can also interact with DynamoDB using raw HTTP requests:

**Create Table**
```bash
curl -X POST http://localhost:4566/ \
  -H "Content-Type: application/x-amz-json-1.0" \
  -H "X-Amz-Target: DynamoDB_20120810.CreateTable" \
  -d '{
    "TableName": "Users",
    "KeySchema": [{"AttributeName": "UserID", "KeyType": "HASH"}],
    "AttributeDefinitions": [{"AttributeName": "UserID", "AttributeType": "S"}],
    "BillingMode": "PAY_PER_REQUEST"
  }'
```

**Put Item**
```bash
curl -X POST http://localhost:4566/ \
  -H "Content-Type: application/x-amz-json-1.0" \
  -H "X-Amz-Target: DynamoDB_20120810.PutItem" \
  -d '{
    "TableName": "Users",
    "Item": {
      "UserID": {"S": "user2"},
      "Name": {"S": "Jane Smith"},
      "Email": {"S": "jane@example.com"},
      "Score": {"N": "85"}
    }
  }'
```

## Manual Setup

### Backend Setup

1. Navigate to the backend directory and install dependencies:
   ```bash
   cd backend
   go mod tidy
   ```

2. Copy the example environment file and update as needed:
   ```bash
   cp .env.example .env
   ```

3. Start the backend server:
   ```bash
   go run main.go
   ```
   The backend will be available at http://localhost:8080

### Frontend Setup

1. In a new terminal, navigate to the frontend directory:
   ```bash
   cd frontend
   ```

2. Install dependencies and start the development server:
   ```bash
   npm install
   npm start
   ```
   The frontend will be available at http://localhost:3000

## Environment Configuration

### Backend Environment Variables

Create a `.env` file in the `backend` directory with the following variables:

```env
ENV=local
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_DEFAULT_REGION=us-east-1
```

## Development Workflow

- **Backend**: Runs on http://localhost:8080
- **Frontend**: Runs on http://localhost:3000
- **LocalStack (DynamoDB)**: Available at http://localhost:4566

## Troubleshooting

- If you encounter CORS issues, ensure your frontend URL is included in the `CORS_ALLOWED_ORIGINS` environment variable
- For LocalStack issues, try running `docker-compose down -v` to clear volumes and restart
- Check container logs with `docker-compose logs [service_name]`

## License

MIT
