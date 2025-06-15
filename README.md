# Football Picking League

A web application for managing and playing in football game prediction leagues.

## Features

- League management system
- Weekly game selection by league admins
- Player predictions submission
- Real-time scoring and leaderboards
- Season statistics tracking

## Tech Stack

- Backend: Go
- Frontend: React
- Hosting: AWS

## Project Structure

```
football-picking-league/
├── backend/         # Go backend service
├── frontend/        # React frontend application
└── docs/           # Documentation
```

## Getting Started

### Prerequisites

- Go 1.20+
- Node.js 18+
- AWS CLI configured

### Installation

1. Clone the repository
2. Set up backend:
   ```bash
cd backend
go mod tidy
go run main.go
```

3. Set up frontend:
   ```bash
cd frontend
npm install
npm start
```

## Development

### Backend Development

The backend runs on port 8080 by default.

### Frontend Development

The frontend runs on port 3000 by default.

## License

MIT
