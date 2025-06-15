export interface Game {
  id: number;
  homeTeam: string;
  awayTeam: string;
  gameTime: string;
  week: number;
  homeScore?: number;
  awayScore?: number;
  isLocked: boolean;
}

export interface Pick {
  id: number;
  gameId: number;
  userId: number;
  selectedTeam: string;
  week: number;
  points?: number;
  isCorrect?: boolean;
}

export interface User {
  id: number;
  username: string;
  totalPoints: number;
  weeklyPoints: number;
  rank: number;
}

export interface Week {
  number: number;
  games: Game[];
  isCurrent: boolean;
  isLocked: boolean;
}
