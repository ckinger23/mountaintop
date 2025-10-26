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

## Login with Pre-Seeded Admin User

The database automatically seeds an admin user on first run!

1. Open http://localhost:3000 in your browser
2. Click "Login"
3. Login with:
   - Email: `admin@example.com`
   - Password: `admin123`
4. You should see an "Admin" link in the navigation!

**Note**: If you were already logged in from a previous session and the database was reset, the app will automatically log you out when it detects the token is outdated. Just login again with the admin credentials above.

## What's Pre-Seeded?

The database automatically includes:
- **Admin user**: admin@example.com / admin123
- **132 teams**: All major college football teams
- **2025 Season**: Active season ready to go
- **3 Weeks**: Week 1, 2, and 3 with proper lock times
- **12 Games**: 4 games per week with realistic matchups

You're ready to start making picks immediately!

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
- Games are automatically seeded on first run - check the backend logs
- Check that the week's lock_time is in the future

### Picks won't submit
- Make sure the game hasn't started yet
- Check browser console for errors (F12)

## Next Steps

- Create additional users for your picks pool
- Add more weeks and games as needed
- Invite friends to join your pool!
- Consider building admin UI pages for easier season/week/game management

## Need to Reset?

To start fresh with a new database:
```bash
rm backend/cfb-picks.db
# Then restart the backend - it will auto-seed everything again
```

Enjoy your picks pool! 🏈
