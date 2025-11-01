import axios from 'axios';
import { AuthResponse, User, Game, Pick, Week, Team, LeaderboardEntry, Season, League } from '../types';

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

  getSeasons: async (): Promise<Season[]> => {
    const { data } = await api.get<Season[]>('/seasons');
    return data;
  },
};

// Picks
export const picksService = {
  submitPick: async (leagueId: number, gameId: number, pickedTeamId: number, pickedOverUnder: string, confidence?: number): Promise<Pick> => {
    const { data } = await api.post<Pick>('/picks', {
      league_id: leagueId,
      game_id: gameId,
      picked_team_id: pickedTeamId,
      picked_over_under: pickedOverUnder,
      confidence,
    });
    return data;
  },

  getMyPicks: async (weekId?: number, leagueId?: number): Promise<Pick[]> => {
    const params: any = {};
    if (weekId) params.week_id = weekId;
    if (leagueId) params.league_id = leagueId;
    const { data } = await api.get<Pick[]>('/picks/me', { params });
    return data;
  },

  getUserPicks: async (userId: number, weekId?: number, leagueId?: number): Promise<Pick[]> => {
    const params: any = {};
    if (weekId) params.week_id = weekId;
    if (leagueId) params.league_id = leagueId;
    const { data } = await api.get<Pick[]>(`/picks/user/${userId}`, { params });
    return data;
  },

  getWeekPicks: async (weekId: number, leagueId?: number): Promise<Pick[]> => {
    const params: any = {};
    if (leagueId) params.league_id = leagueId;
    const { data } = await api.get<Pick[]>(`/picks/week/${weekId}`, { params });
    return data;
  },
};

// Leaderboard
export const leaderboardService = {
  getLeaderboard: async (seasonId?: number, leagueId?: number): Promise<LeaderboardEntry[]> => {
    const params: any = {};
    if (seasonId) params.season_id = seasonId;
    if (leagueId) params.league_id = leagueId;
    const { data } = await api.get<LeaderboardEntry[]>('/leaderboard', { params });
    return data;
  },
};

// Leagues
export const leaguesService = {
  createLeague: async (name: string, description: string, isPublic: boolean): Promise<League> => {
    const { data } = await api.post<League>('/leagues', {
      name,
      description,
      is_public: isPublic,
    });
    return data;
  },

  getMyLeagues: async (): Promise<League[]> => {
    const { data } = await api.get<League[]>('/leagues');
    return data;
  },

  getLeague: async (id: number): Promise<League> => {
    const { data } = await api.get<League>(`/leagues/${id}`);
    return data;
  },

  updateLeague: async (id: number, name: string, description: string, isPublic: boolean): Promise<League> => {
    const { data } = await api.put<League>(`/leagues/${id}`, {
      name,
      description,
      is_public: isPublic,
    });
    return data;
  },

  joinLeague: async (code: string): Promise<League> => {
    const { data } = await api.post<League>('/leagues/join', { code });
    return data;
  },

  browsePublicLeagues: async (): Promise<League[]> => {
    const { data } = await api.get<League[]>('/leagues/browse');
    return data;
  },

  leaveLeague: async (id: number): Promise<void> => {
    await api.delete(`/leagues/${id}/leave`);
  },
};

// Admin
export const adminService = {
  // Game management
  createGame: async (weekId: number, homeTeamId: number, awayTeamId: number, gameTime: string, homeSpread: number, total: number): Promise<Game> => {
    const { data } = await api.post<Game>('/admin/games', {
      week_id: weekId,
      home_team_id: homeTeamId,
      away_team_id: awayTeamId,
      game_time: gameTime,
      home_spread: homeSpread,
      total,
    });
    return data;
  },

  updateGame: async (gameId: number, weekId: number, homeTeamId: number, awayTeamId: number, gameTime: string, homeSpread: number, total: number): Promise<Game> => {
    const { data } = await api.put<Game>(`/admin/games/${gameId}`, {
      week_id: weekId,
      home_team_id: homeTeamId,
      away_team_id: awayTeamId,
      game_time: gameTime,
      home_spread: homeSpread,
      total,
    });
    return data;
  },

  deleteGame: async (gameId: number): Promise<void> => {
    await api.delete(`/admin/games/${gameId}`);
  },

  updateGameResult: async (gameId: number, homeScore: number, awayScore: number, isFinal: boolean): Promise<Game> => {
    const { data } = await api.put<Game>(`/admin/games/${gameId}/result`, {
      home_score: homeScore,
      away_score: awayScore,
      is_final: isFinal,
    });
    return data;
  },

  // Week management
  createWeek: async (seasonId: number, weekNumber: number, name: string): Promise<Week> => {
    const { data } = await api.post<Week>('/admin/weeks', {
      season_id: seasonId,
      week_number: weekNumber,
      name,
    });
    return data;
  },

  updateWeek: async (weekId: number, seasonId: number, weekNumber: number, name: string): Promise<Week> => {
    const { data } = await api.put<Week>(`/admin/weeks/${weekId}`, {
      season_id: seasonId,
      week_number: weekNumber,
      name,
    });
    return data;
  },

  openWeekForPicks: async (weekId: number, pickDeadline: string): Promise<Week> => {
    const { data } = await api.put<Week>(`/admin/weeks/${weekId}/open`, {
      pick_deadline: pickDeadline,
    });
    return data;
  },

  lockWeek: async (weekId: number): Promise<Week> => {
    const { data } = await api.put<Week>(`/admin/weeks/${weekId}/lock`);
    return data;
  },

  completeWeek: async (weekId: number): Promise<Week> => {
    const { data } = await api.put<Week>(`/admin/weeks/${weekId}/complete`);
    return data;
  },

  // Season management
  createSeason: async (leagueId: number, year: number, name: string, isActive: boolean) => {
    const { data } = await api.post('/admin/seasons', {
      league_id: leagueId,
      year,
      name,
      is_active: isActive,
    });
    return data;
  },
};

export default api;
