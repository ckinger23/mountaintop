# API Testing Examples

Use these curl commands to test the API directly.

## Authentication

### Register a New User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "display_name": "Test User"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com",
    "display_name": "Test User",
    "is_admin": false
  }
}
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Get Current User
```bash
TOKEN="your-jwt-token-here"

curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

## Games & Weeks

### Get All Teams
```bash
curl -X GET http://localhost:8080/api/teams
```

### Get All Weeks
```bash
curl -X GET http://localhost:8080/api/weeks

# Filter by season
curl -X GET "http://localhost:8080/api/weeks?season_id=1"
```

### Get Current Week
```bash
curl -X GET http://localhost:8080/api/weeks/current
```

### Get All Games
```bash
curl -X GET http://localhost:8080/api/games

# Filter by week
curl -X GET "http://localhost:8080/api/games?week_id=1"
```

### Get Single Game
```bash
curl -X GET http://localhost:8080/api/games/1
```

## Picks

### Submit a Pick
```bash
TOKEN="your-jwt-token-here"

curl -X POST http://localhost:8080/api/picks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "game_id": 1,
    "picked_team_id": 2,
    "confidence": 5
  }'
```

### Get My Picks
```bash
TOKEN="your-jwt-token-here"

curl -X GET http://localhost:8080/api/picks/me \
  -H "Authorization: Bearer $TOKEN"

# Filter by week
curl -X GET "http://localhost:8080/api/picks/me?week_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

### Get Another User's Picks
```bash
TOKEN="your-jwt-token-here"

curl -X GET http://localhost:8080/api/picks/user/2 \
  -H "Authorization: Bearer $TOKEN"

# Filter by week
curl -X GET "http://localhost:8080/api/picks/user/2?week_id=1" \
  -H "Authorization: Bearer $TOKEN"
```

### Get All Picks for a Week
```bash
TOKEN="your-jwt-token-here"

curl -X GET http://localhost:8080/api/picks/week/1 \
  -H "Authorization: Bearer $TOKEN"
```

## Leaderboard

### Get Leaderboard
```bash
curl -X GET http://localhost:8080/api/leaderboard

# Filter by season
curl -X GET "http://localhost:8080/api/leaderboard?season_id=1"
```

## Admin Endpoints

### Create a Game
```bash
ADMIN_TOKEN="your-admin-jwt-token-here"

curl -X POST http://localhost:8080/api/admin/games \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "week_id": 1,
    "home_team_id": 1,
    "away_team_id": 2,
    "game_time": "2025-09-07T19:00:00Z",
    "home_spread": -3.5
  }'
```

### Update Game Result
```bash
ADMIN_TOKEN="your-admin-jwt-token-here"

curl -X PUT http://localhost:8080/api/admin/games/1/result \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "home_score": 28,
    "away_score": 21,
    "is_final": true
  }'
```

## Testing Workflow

### Complete Test Flow

1. **Register a user**:
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@example.com", "password": "pass123", "display_name": "Test"}' \
  | jq -r '.token' > token.txt
```

2. **Set token variable**:
```bash
TOKEN=$(cat token.txt)
```

3. **Get current week**:
```bash
curl -X GET http://localhost:8080/api/weeks/current \
  -H "Authorization: Bearer $TOKEN" \
  | jq
```

4. **Get games for the week**:
```bash
curl -X GET "http://localhost:8080/api/games?week_id=1" \
  -H "Authorization: Bearer $TOKEN" \
  | jq
```

5. **Submit a pick**:
```bash
curl -X POST http://localhost:8080/api/picks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"game_id": 1, "picked_team_id": 1}' \
  | jq
```

6. **View leaderboard**:
```bash
curl -X GET http://localhost:8080/api/leaderboard | jq
```

## Using Postman

Import this collection:

1. Create a new collection called "CFB Picks"
2. Add an environment variable `base_url` = `http://localhost:8080`
3. Add an environment variable `token` (will be set from login response)
4. Create requests for each endpoint above
5. In the Tests tab of login/register, add:
   ```javascript
   pm.environment.set("token", pm.response.json().token);
   ```
6. In Authorization tab for protected routes, use:
   - Type: Bearer Token
   - Token: `{{token}}`

## Response Codes

- `200` - Success
- `201` - Created (new resource)
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (not logged in)
- `403` - Forbidden (not admin)
- `404` - Not Found
- `409` - Conflict (duplicate)
- `500` - Server Error

## Common Errors

### 401 Unauthorized
- Token is missing or invalid
- Token has expired (24hr expiration)
- Solution: Login again to get a new token

### 403 Forbidden
- Trying to access admin route without admin privileges
- Solution: Make user admin in database

### 409 Conflict
- Username or email already exists
- Solution: Use different credentials

### 400 Bad Request
- Missing required fields
- Invalid data format
- Trying to pick after lock time
- Solution: Check request body matches expected format
