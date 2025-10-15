# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Mountaintop is a college football picks pool application with a Go backend and React frontend. Users submit picks for weekly games, admins enter results, and the system automatically scores picks and maintains a leaderboard.

## Development Commands

### Backend (Go)

Start the backend server:
```bash
cd backend
go run cmd/api/main.go
```

The backend runs on `http://localhost:8080` and automatically:
- Creates SQLite database at `./cfb-picks.db` (or path in `DB_PATH` env var)
- Runs GORM migrations on startup
- Seeds initial team data if database is empty

Test database queries:
```bash
sqlite3 backend/cfb-picks.db
```

### Frontend (React + TypeScript)

Install dependencies:
```bash
cd frontend
npm install
```

Start the dev server:
```bash
npm run dev
```

Frontend runs on `http://localhost:3000` with Vite proxy forwarding `/api/*` requests to the backend.

Build for production:
```bash
npm run build
```

## Architecture

### Tech Stack
- **Backend**: Go with Chi router, GORM ORM, SQLite database, JWT authentication
- **Frontend**: React 18, TypeScript, Vite, Tailwind CSS, React Router v6, Axios

### Request Flow Pattern

All authenticated requests follow this pattern:
1. Frontend component calls service function in [frontend/src/services/api.ts](frontend/src/services/api.ts)
2. Axios interceptor adds JWT from localStorage to Authorization header
3. Request proxied to backend (dev) or direct to API (prod)
4. Chi router in [backend/cmd/api/main.go](backend/cmd/api/main.go) routes to handler
5. Middleware stack executes: Logger → Recoverer → CORS → Auth (if protected) → Admin (if admin route)
6. Handler in `backend/internal/handlers/` processes request
7. GORM interacts with database using models from [backend/internal/models/models.go](backend/internal/models/models.go)
8. Response sent back to frontend

### Database Models

Core entities defined in [backend/internal/models/models.go](backend/internal/models/models.go):
- **User**: Authentication, admin flag, relationships to picks
- **Season**: Year-based organization (e.g., 2025), has many weeks
- **Week**: Weekly grouping of games, has lock_time for pick submission deadline
- **Team**: CFB teams with name, abbreviation, logo, conference
- **Game**: Matchup between two teams, includes spread, scores, winner (when final)
- **Pick**: User's selection for a game, scored after game is final

Key relationships:
- User → Picks (one-to-many)
- Season → Weeks → Games (hierarchical)
- Game → HomeTeam, AwayTeam, Picks (foreign keys)
- Pick → User, Game, PickedTeam (composite unique index on user_id + game_id)

### API Routes

Public routes (no auth):
- `POST /api/auth/register`, `POST /api/auth/login`
- `GET /api/teams`, `GET /api/seasons`, `GET /api/leaderboard`

Protected routes (JWT required) - middleware in [backend/internal/middleware/auth.go](backend/internal/middleware/auth.go):
- `GET /api/auth/me` - current user
- `GET /api/games`, `GET /api/weeks`, `GET /api/weeks/current`
- `POST /api/picks`, `GET /api/picks/me`, `GET /api/picks/user/:userId`, `GET /api/picks/week/:weekId`

Admin routes (JWT + is_admin=true):
- `POST /api/admin/games` - create game
- `PUT /api/admin/games/:id/result` - enter scores, triggers pick scoring

### Pick Scoring System

When admin marks game as final in [backend/internal/handlers/games.go](backend/internal/handlers/games.go):
1. Determine winner based on scores
2. Query all picks for that game
3. Compare each pick's `picked_team_id` to `winner_team_id`
4. Set `is_correct` boolean and `points_earned` (typically 1 for correct, 0 for incorrect)
5. Update all picks in single transaction

Leaderboard aggregates: total_points, correct_picks, total_picks, win_pct per user.

### Frontend Structure

- **Pages** in `frontend/src/pages/`: Login, Register, MakePicks, Leaderboard, Admin
- **Auth Context**: [frontend/src/hooks/useAuth.tsx](frontend/src/hooks/useAuth.tsx) manages JWT token and user state
- **API Services**: [frontend/src/services/api.ts](frontend/src/services/api.ts) exports typed service objects (authService, gamesService, picksService, etc.)
- **Types**: [frontend/src/types/index.ts](frontend/src/types/index.ts) defines TypeScript interfaces matching Go models

Protected routes use `ProtectedRoute` component that checks authentication before rendering.

## Key Development Patterns

### Making Database Schema Changes

1. Update models in [backend/internal/models/models.go](backend/internal/models/models.go)
2. GORM AutoMigrate runs on startup in [backend/internal/database/database.go](backend/internal/database/database.go)
3. For complex migrations, add explicit SQL in `Migrate()` function
4. Restart backend to apply changes

### Adding New API Endpoints

1. Define handler function in appropriate file in `backend/internal/handlers/`
2. Register route in [backend/cmd/api/main.go](backend/cmd/api/main.go) (choose public, protected, or admin group)
3. Add corresponding service function in [frontend/src/services/api.ts](frontend/src/services/api.ts)
4. Add TypeScript type if needed in [frontend/src/types/index.ts](frontend/src/types/index.ts)

### Creating Admin User

The database doesn't have a registration flow for admins:
```sql
sqlite3 backend/cfb-picks.db
UPDATE users SET is_admin = 1 WHERE email = 'user@example.com';
```

### Setting Up a Season

Currently requires manual SQL (admin UI for this is in roadmap per README):
```sql
INSERT INTO seasons (created_at, updated_at, year, name, is_active)
VALUES (datetime('now'), datetime('now'), 2025, '2025 Season', 1);

INSERT INTO weeks (created_at, updated_at, season_id, week_number, name, lock_time)
VALUES (datetime('now'), datetime('now'), 1, 1, 'Week 1', datetime('now', '+7 days'));

INSERT INTO games (created_at, updated_at, week_id, home_team_id, away_team_id, game_time, home_spread, is_final)
VALUES (datetime('now'), datetime('now'), 1, 1, 2, datetime('now', '+5 days'), -3.5, 0);
```

Teams are seeded automatically on first run: Alabama, Georgia, Ohio State, Michigan, Texas, USC, Oregon, Penn State.

## Important Notes

- **Game Locking**: Picks cannot be submitted after `week.lock_time` or after `game.game_time` (enforced in backend handlers)
- **JWT Expiration**: Tokens expire after 24 hours (set in auth.go)
- **Unique Constraint**: One pick per user per game (enforced by DB unique index on user_id + game_id)
- **CORS**: Development allows localhost:3000 and localhost:5173 (Vite default ports)
- **Production Swap**: To use PostgreSQL, change import in [backend/internal/database/database.go](backend/internal/database/database.go) from `gorm.io/driver/sqlite` to `gorm.io/driver/postgres`

## Testing Quick Flow

1. Start backend: `cd backend && go run cmd/api/main.go`
2. Start frontend: `cd frontend && npm run dev`
3. Register user at http://localhost:3000
4. Make user admin via sqlite3
5. Create season/week/games via SQL or admin API
6. Submit picks before lock time
7. Enter game results as admin
8. View scored leaderboard
