# Quick Start Guide

Get up and running in 5 minutes!

## Terminal 1 - Backend

```bash
cd backend
go mod download
go run cmd/api/main.go
```

You should see:
```
Database connection established
Running database migrations...
Migrations completed successfully
Seeding database with initial data...
Database seeded successfully
Server starting on port 8080...
```

## Terminal 2 - Frontend

```bash
cd frontend
npm install
npm run dev
```

You should see:
```
VITE v5.0.8  ready in XXX ms

➜  Local:   http://localhost:3000/
```

## Create Your First User

1. Open http://localhost:3000 in your browser
2. Click "Register"
3. Fill out the form:
   - Username: `admin`
   - Email: `admin@example.com`
   - Password: `password123`
   - Display Name: `Admin User`
4. Click "Register"

## Make Yourself an Admin

```bash
# In a new terminal
sqlite3 backend/cfb-picks.db
```

```sql
UPDATE users SET is_admin = 1 WHERE email = 'admin@example.com';
.quit
```

Refresh your browser and you should now see an "Admin" link in the navigation!

## What's Next?

### As Admin, you need to:
1. Create a Season (currently manual via database or add admin UI)
2. Create Weeks for that season
3. Create Games for each week

### For now, here's a quick database setup:

```bash
sqlite3 backend/cfb-picks.db
```

```sql
-- Create a season
INSERT INTO seasons (created_at, updated_at, year, name, is_active) 
VALUES (datetime('now'), datetime('now'), 2025, '2025 Season', 1);

-- Create a week (adjust season_id if needed)
INSERT INTO weeks (created_at, updated_at, season_id, week_number, name, lock_time) 
VALUES (datetime('now'), datetime('now'), 1, 1, 'Week 1', datetime('now', '+7 days'));

-- Create a game (adjust week_id and team IDs as needed)
-- Teams 1-8 were seeded: Alabama, Georgia, Ohio State, Michigan, Texas, USC, Oregon, Penn State
INSERT INTO games (created_at, updated_at, week_id, home_team_id, away_team_id, game_time, home_spread, is_final) 
VALUES (datetime('now'), datetime('now'), 1, 1, 2, datetime('now', '+5 days'), -3.5, 0);

INSERT INTO games (created_at, updated_at, week_id, home_team_id, away_team_id, game_time, home_spread, is_final) 
VALUES (datetime('now'), datetime('now'), 1, 3, 4, datetime('now', '+5 days'), -7.0, 0);

.quit
```

Now refresh your browser and you should see games to pick!

## Testing the Full Flow

1. **Make Picks**: Click on a team to select them, then "Save Pick"
2. **View Leaderboard**: See current standings (will be empty until games are scored)
3. **Admin Panel**: Enter game results:
   - Enter scores for both teams
   - Check "Mark as Final"
   - Click "Update Result"
4. **Leaderboard Updates**: Picks are automatically scored and leaderboard updates!

## Troubleshooting

### Backend won't start
- Make sure port 8080 is available
- Check Go version: `go version` (need 1.21+)

### Frontend won't start
- Check Node version: `node --version` (need 18+)
- Try deleting `node_modules` and running `npm install` again

### Can't see games
- Make sure you created a season, week, and games in the database
- Check that the week's lock_time is in the future

### Picks won't submit
- Make sure the game hasn't started yet
- Check browser console for errors (F12)

## Next Steps

- Add more teams to the database
- Create more weeks and games
- Invite friends to join your pool!
- Consider building admin UI pages to avoid manual database work

Enjoy your picks pool! 🏈
