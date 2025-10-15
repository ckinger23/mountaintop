# System Architecture

## High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│                         Frontend                             │
│                     (React + TypeScript)                     │
│                                                              │
│  ┌──────────┐  ┌──────────┐  ┌────────────┐  ┌──────────┐ │
│  │  Login   │  │   Make   │  │Leaderboard │  │  Admin   │ │
│  │  Page    │  │  Picks   │  │   Page     │  │  Panel   │ │
│  └──────────┘  └──────────┘  └────────────┘  └──────────┘ │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              API Service Layer (Axios)               │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │          Auth Context (JWT Token Management)         │  │
│  └──────────────────────────────────────────────────────┘  │
└──────────────────────┬───────────────────────────────────────┘
                       │ HTTP/REST API
                       │ (localhost:3000 → proxy → localhost:8080)
┌──────────────────────▼───────────────────────────────────────┐
│                         Backend                              │
│                        (Go + Chi)                            │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │                   API Routes                          │  │
│  │  /api/auth/* | /api/games/* | /api/picks/*          │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                  │
│  ┌────────────────────────┼────────────────────────────┐   │
│  │      Middleware Layer  │                            │   │
│  │  ┌──────────┐  ┌──────▼──────┐  ┌──────────────┐  │   │
│  │  │  CORS    │  │     JWT     │  │    Admin     │  │   │
│  │  │ Handler  │  │    Auth     │  │  Middleware  │  │   │
│  │  └──────────┘  └─────────────┘  └──────────────┘  │   │
│  └──────────────────────┬──────────────────────────────┘   │
│                         │                                   │
│  ┌────────────────────────────────────────────────────┐    │
│  │              Handler Layer                          │    │
│  │  ┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐  │    │
│  │  │  Auth  │  │ Games  │  │ Picks  │  │ Admin  │  │    │
│  │  └────────┘  └────────┘  └────────┘  └────────┘  │    │
│  └────────────────────┬────────────────────────────────┘    │
│                       │                                     │
│  ┌────────────────────▼────────────────────────────────┐   │
│  │                 GORM (ORM)                           │   │
│  └────────────────────┬────────────────────────────────┘   │
└───────────────────────┼──────────────────────────────────────┘
                        │
┌───────────────────────▼──────────────────────────────────────┐
│                      Database                                │
│                     (SQLite/PostgreSQL)                      │
│                                                              │
│  ┌─────────┐  ┌────────┐  ┌────────┐  ┌────────┐          │
│  │  Users  │  │ Seasons│  │  Teams │  │  Weeks │          │
│  └─────────┘  └────────┘  └────────┘  └────────┘          │
│  ┌─────────┐  ┌────────┐                                   │
│  │  Games  │  │  Picks │                                   │
│  └─────────┘  └────────┘                                   │
└──────────────────────────────────────────────────────────────┘
```

## Request Flow Examples

### 1. User Makes a Pick

```
User Browser
    │
    │ 1. Click team to pick
    ▼
React Component (MakePicks.tsx)
    │
    │ 2. Call picksService.submitPick()
    ▼
API Service (src/services/api.ts)
    │
    │ 3. POST /api/picks with JWT token
    ▼
Backend Router (cmd/api/main.go)
    │
    │ 4. Route to handler
    ▼
Auth Middleware (internal/middleware/auth.go)
    │
    │ 5. Validate JWT, extract user
    ▼
Pick Handler (internal/handlers/picks.go)
    │
    │ 6. Validate pick (not locked, valid team)
    │ 7. Create/update pick in database
    ▼
Database (GORM)
    │
    │ 8. Insert/Update pick record
    ▼
Response
    │
    │ 9. Return pick with relationships
    ▼
Frontend updates UI
```

### 2. Admin Enters Game Result

```
Admin User
    │
    │ 1. Enter scores, mark final
    ▼
Admin Component (Admin.tsx)
    │
    │ 2. Call adminService.updateGameResult()
    ▼
API Service
    │
    │ 3. PUT /api/admin/games/:id/result
    ▼
Backend Router
    │
    ▼
Auth Middleware → Admin Middleware
    │
    │ 4. Validate JWT + admin status
    ▼
Game Handler (internal/handlers/games.go)
    │
    │ 5. Update game result
    │ 6. Determine winner
    │ 7. Trigger calculatePickResults()
    ▼
Pick Calculation
    │
    │ 8. Update all picks for this game
    │ 9. Mark correct/incorrect
    │ 10. Calculate points
    ▼
Database updates
    │
    ▼
Response → Frontend shows success
    │
    ▼
Leaderboard automatically reflects new scores
```

## Data Flow

### Authentication Flow
1. User registers/logs in
2. Backend generates JWT token (24hr expiration)
3. Frontend stores token in localStorage
4. All subsequent requests include token in Authorization header
5. Backend validates token on protected routes

### Pick Submission Flow
1. Frontend loads current week's games
2. User selects teams
3. Frontend checks lock time (before sending)
4. Backend validates:
   - User is authenticated
   - Game hasn't started
   - Week isn't locked
   - Team is valid for game
5. Pick saved to database
6. Frontend shows confirmation

### Scoring Flow
1. Admin enters game result
2. Backend updates game record
3. If game is final:
   - Query all picks for that game
   - Compare picked_team_id to winner_team_id
   - Set is_correct and points_earned
   - Save updated picks
4. Leaderboard queries recalculate automatically

## Security Layers

### Frontend
- Route guards (ProtectedRoute component)
- Auth context checks user state
- Admin UI only shown to admins

### Backend
- JWT validation middleware
- Admin-only routes protected by AdminMiddleware
- Password hashing (bcrypt)
- Input validation on all handlers

### Database
- Unique constraints (email, username)
- Foreign key relationships
- Soft deletes (DeletedAt)

## Scaling Considerations

### Current (MVP)
- SQLite database (single file)
- Single server instance
- Synchronous request handling

### Future (Production)
- PostgreSQL (horizontal scaling)
- Load balancer with multiple Go instances
- Redis for session management
- Background workers for email notifications
- CDN for static frontend assets
