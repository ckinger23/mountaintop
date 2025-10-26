import axios from 'axios';
import { AuthResponse, User, Game, Pick, Week, Team, LeaderboardEntry } from '../types';

const api = axios.create({
  baseURL: '/api',
});

// Add auth token to requests
// registers a function that is run before every HTTP request
api.interceptors.request.use((config) => {
  // retrieve jwt stored when user logged in
  // localstorage is a builti-in browser API from the Web Storage spec
  // Not from react, npm package, or code
  // native feature of all modern web browsers
  // persists data even after closing the browser
  // 5-10 MB storage limit
  // Synchronous (blocking)
  const token = localStorage.getItem('token');
  if (token) {
    // add token to the header before API call
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Auth
export const authService = {
  register: async (username: string, email: string, password: string, displayName: string): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/register', {
      username,
      email,
      password,
      display_name: displayName,
    });
    localStorage.setItem('token', data.token);
    return data;
  },

  login: async (email: string, password: string): Promise<AuthResponse> => {
    const { data } = await api.post<AuthResponse>('/auth/login', { email, password });
    localStorage.setItem('token', data.token);
    return data;
  },

  logout: () => {
    localStorage.removeItem('token');
  },

  getCurrentUser: async (): Promise<User> => {
    const { data } = await api.get<User>('/auth/me');
    return data;
  },

  isAuthenticated: (): boolean => {
    return !!localStorage.getItem('token');
  },
};

// Games
export const gamesService = {
  getGames: async (weekId?: number): Promise<Game[]> => {
    const params = weekId ? { week_id: weekId } : {};
    const { data } = await api.get<Game[]>('/games', { params });
    return data;
  },

  getGame: async (id: number): Promise<Game> => {
    const { data } = await api.get<Game>(`/games/${id}`);
    return data;
  },

  getWeeks: async (seasonId?: number): Promise<Week[]> => {
    const params = seasonId ? { season_id: seasonId } : {};
    const { data } = await api.get<Week[]>('/weeks', { params });
    return data;
  },

  getCurrentWeek: async (): Promise<Week> => {
    const { data } = await api.get<Week>('/weeks/current');
    return data;
  },

  getTeams: async (): Promise<Team[]> => {
    const { data } = await api.get<Team[]>('/teams');
    return data;
  },
};

// Picks
export const picksService = {
  submitPick: async (gameId: number, pickedTeamId: number, confidence?: number): Promise<Pick> => {
    const { data } = await api.post<Pick>('/picks', {
      game_id: gameId,
      picked_team_id: pickedTeamId,
      confidence,
    });
    return data;
  },

  getMyPicks: async (weekId?: number): Promise<Pick[]> => {
    const params = weekId ? { week_id: weekId } : {};
    const { data } = await api.get<Pick[]>('/picks/me', { params });
    return data;
  },

  getUserPicks: async (userId: number, weekId?: number): Promise<Pick[]> => {
    const params = weekId ? { week_id: weekId } : {};
    const { data } = await api.get<Pick[]>(`/picks/user/${userId}`, { params });
    return data;
  },

  getWeekPicks: async (weekId: number): Promise<Pick[]> => {
    const { data } = await api.get<Pick[]>(`/picks/week/${weekId}`);
    return data;
  },
};

// Leaderboard
export const leaderboardService = {
  getLeaderboard: async (seasonId?: number): Promise<LeaderboardEntry[]> => {
    const params = seasonId ? { season_id: seasonId } : {};
    const { data } = await api.get<LeaderboardEntry[]>('/leaderboard', { params });
    return data;
  },
};

// Admin
export const adminService = {
  updateGameResult: async (gameId: number, homeScore: number, awayScore: number, isFinal: boolean): Promise<Game> => {
    const { data } = await api.put<Game>(`/admin/games/${gameId}/result`, {
      home_score: homeScore,
      away_score: awayScore,
      is_final: isFinal,
    });
    return data;
  },

  createGame: async (weekId: number, homeTeamId: number, awayTeamId: number, gameTime: string, homeSpread: number): Promise<Game> => {
    const { data } = await api.post<Game>('/admin/games', {
      week_id: weekId,
      home_team_id: homeTeamId,
      away_team_id: awayTeamId,
      game_time: gameTime,
      home_spread: homeSpread,
    });
    return data;
  },
};

export default api;
