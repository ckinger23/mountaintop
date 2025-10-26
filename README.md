# CFB Picks System

A full-stack college football picks application built with Go and React.

## 🏗️ Architecture

### Backend (Go)
- **Framework**: Chi router
- **Database**: SQLite (easily swappable to PostgreSQL)
- **ORM**: GORM
- **Authentication**: JWT tokens
- **API**: RESTful endpoints

### Frontend (React + TypeScript)
- **Framework**: React 18 with Vite
- **Routing**: React Router v6
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **State Management**: React Context (for auth)

## 📁 Project Structure

```
cfb-picks-system/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go              # Main application entry point
│   ├── internal/
│   │   ├── database/
│   │   │   └── database.go          # Database connection & migrations
│   │   ├── handlers/
│   │   │   ├── auth.go              # Authentication handlers
│   │   │   ├── games.go             # Game & week handlers
│   │   │   └── picks.go             # Pick submission handlers
│   │   ├── middleware/
│   │   │   └── auth.go              # JWT middleware
│   │   └── models/
│   │       └── models.go            # Database models
│   ├── migrations/                   # SQL migrations (optional)
│   └── go.mod                        # Go dependencies
└── frontend/
    ├── src/
    │   ├── components/               # Reusable components
    │   ├── hooks/
    │   │   └── useAuth.tsx          # Authentication hook
    │   ├── pages/
    │   │   ├── Login.tsx            # Login page
    │   │   ├── MakePicks.tsx        # Pick submission page
    │   │   ├── Leaderboard.tsx      # Standings page
    │   │   └── Admin.tsx            # Admin controls
    │   ├── services/
    │   │   └── api.ts               # API service layer
    │   ├── types/
    │   │   └── index.ts             # TypeScript types
    │   ├── App.tsx                   # Main app component
    │   ├── main.tsx                  # React entry point
    │   └── index.css                 # Global styles
    └── package.json                  # Node dependencies
```

## 🚀 Getting Started

### Prerequisites
- Go 1.21+ ([download](https://go.dev/dl/))
- Node.js 18+ and npm ([download](https://nodejs.org/))

### Backend Setup

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install Go dependencies:
```bash
go mod download
```

3. Run the server:
```bash
go run cmd/api/main.go
```

The API will start on `http://localhost:8080`

**Important:** On first run, the database will be created and seeded with sample teams.

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install Node dependencies:
```bash
npm install
```

3. Start the development server:
```bash
npm run dev
```

The frontend will start on `http://localhost:3000`

## 🔑 Key Features

### Current Features
- ✅ User registration and authentication
- ✅ JWT-based authorization
- ✅ Pick submission with game locking (before game time)
- ✅ Admin panel for entering game results
- ✅ Automatic pick scoring
- ✅ Real-time leaderboard
- ✅ Weekly picks tracking
- ✅ Win percentage calculation

### Roadmap
- [ ] Email notifications for pick reminders
- [ ] Confidence pools (rank picks by confidence)
- [ ] Historical data charts
- [ ] Mobile app (React Native)
- [ ] Social features (comments, trash talk)
- [ ] Integration with ESPN API for automatic game results

## 📊 Database Schema

### Users
- Authentication and profile information
- Admin flags

### Seasons & Weeks
- Organize games by year and week
- Lock times for pick submission

### Teams
- College football teams
- Conference information

### Games
- Matchup details
- Spread information
- Final scores

### Picks
- User selections
- Correctness tracking
- Points earned

## 🔒 API Endpoints

### Public
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login
- `GET /api/teams` - List all teams
- `GET /api/leaderboard` - View standings

### Protected (Requires Auth)
- `GET /api/auth/me` - Get current user
- `GET /api/games` - List games
- `GET /api/weeks` - List weeks
- `POST /api/picks` - Submit a pick
- `GET /api/picks/me` - Get user's picks
- `GET /api/picks/user/:userId` - View another user's picks
- `GET /api/picks/week/:weekId` - View all picks for a week

### Admin Only
- `POST /api/admin/games` - Create a game
- `PUT /api/admin/games/:id/result` - Update game result

## 🎨 Frontend Pages

1. **Login/Register** - User authentication
2. **Make Picks** - Submit picks for the current week
3. **Leaderboard** - View standings and stats
4. **Admin Panel** - Enter game results (admin only)

## ⚙️ Configuration

### Backend Environment Variables
```bash
DB_PATH=./cfb-picks.db    # Database file location
PORT=8080                  # API server port
```

### Frontend Configuration
The frontend automatically proxies API requests to `http://localhost:8080` during development (configured in `vite.config.js`).

## 🧪 Testing the Application

1. Start both backend and frontend
2. Login with the pre-seeded admin user:
   - Email: `admin@example.com`
   - Password: `admin123`
3. The database automatically seeds on first run with:
   - Admin user (admin@example.com / admin123)
   - 132 college football teams
   - 2025 season with 3 weeks
   - 12 sample games (4 per week)
4. As admin, submit picks before game time
5. Enter game results in the Admin panel
6. View updated leaderboard with automatic scoring

## 🔄 Next Steps

### Immediate Enhancements
1. **Create Admin Setup UI** - Add pages for creating seasons, weeks, and games
2. **Improve Pick UI** - Add filters, search, and better mobile experience
3. **Add Notifications** - Email/push notifications for pick deadlines
4. **Stats Dashboard** - Personal stats, head-to-head records
5. **Integration** - Connect to ESPN/other APIs for automatic game data

### Production Deployment
1. **Database**: Migrate from SQLite to PostgreSQL
2. **Environment**: Use environment variables for secrets
3. **Frontend**: Build and serve static files
4. **Hosting**: Deploy to AWS, Heroku, or similar
5. **Domain**: Set up custom domain and SSL

## 📝 License

MIT License - feel free to use this for your own picks pool!

## 🤝 Contributing

This is a personal project, but suggestions and improvements are welcome!

## 📧 Contact

Questions? Reach out to the project maintainer.

---

Built with ❤️ for college football fans
