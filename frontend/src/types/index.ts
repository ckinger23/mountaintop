export interface User {
  id: number;
  username: string;
  email: string;
  display_name: string;
  is_admin: boolean;
  created_at: string;
}

export interface Team {
  id: number;
  name: string;
  abbreviation: string;
  logo_url?: string;
  conference: string;
}

export interface Season {
  id: number;
  year: number;
  name: string;
  is_active: boolean;
}

export interface Week {
  id: number;
  season_id: number;
  week_number: number;
  name: string;
  lock_time: string;
  season?: Season;
}

export interface Game {
  id: number;
  week_id: number;
  home_team_id: number;
  away_team_id: number;
  game_time: string;
  home_spread: number;
  total: number;
  is_final: boolean;
  home_score?: number;
  away_score?: number;
  winner_team_id?: number;
  home_team: Team;
  away_team: Team;
  week?: Week;
}

export interface Pick {
  id: number;
  user_id: number;
  game_id: number;
  picked_team_id: number;
  picked_over_under: string; // "over" or "under"
  confidence?: number;
  spread_correct?: boolean;
  over_under_correct?: boolean;
  points_earned: number;
  game: Game;
  picked_team: Team;
  user?: User;
}

export interface LeaderboardEntry {
  user_id: number;
  username: string;
  display_name: string;
  total_points: number;
  correct_picks: number;
  total_picks: number;
  win_pct: number;
}

export interface AuthResponse {
  token: string;
  user: User;
}
