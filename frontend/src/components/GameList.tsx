import { useState, useEffect } from 'react';
import type { Game } from '../types';
import axios from 'axios';

interface GameListProps {
  week: number;
}

const GameList: React.FC<GameListProps> = ({ week }) => {
  const [gameData, setGameData] = useState<{
    games: Game[];
    weekNumber: number;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchGames = async () => {
      try {
        const response = await axios.get(`/api/games/week/${week}`);
        // Ensure we're working with an array
        const games = Array.isArray(response.data) ? response.data : [];
        setGameData({
          games: games.map((game: any) => ({
            id: game.id,
            homeTeam: game.homeTeam || 'Unknown',
            awayTeam: game.awayTeam || 'Unknown',
            gameTime: game.gameTime || 'TBD',
            week: game.week || week,
            homeScore: game.homeScore,
            awayScore: game.awayScore,
            isLocked: game.isLocked || false
          })),
          weekNumber: week
        });
        setLoading(false);
      } catch (err) {
        console.error('Error fetching games:', err);
        setError('Failed to load games');
        setLoading(false);
      }
    };

    fetchGames();
  }, [week]);

  if (loading) return (
    <div className="placeholder">
      <div className="placeholder-content">
        <div className="placeholder-icon">⏳</div>
        <p>Loading games for Week {week}...</p>
      </div>
    </div>
  );

  if (error) return (
    <div className="placeholder error">
      <div className="placeholder-content">
        <div className="placeholder-icon">❌</div>
        <p>Failed to load games</p>
        <p className="placeholder-subtext">Please try refreshing the page</p>
      </div>
    </div>
  );

  if (!gameData || gameData.games.length === 0) return (
    <div className="placeholder">
      <div className="placeholder-content">
        <div className="placeholder-icon">⚽</div>
        <p>No games scheduled for Week {week}</p>
        <p className="placeholder-subtext">Check back when the schedule is released</p>
      </div>
    </div>
  );

  return (
    <div className="game-list">
      <h2>Week {gameData.weekNumber} Games</h2>
      {gameData.games.map((game) => (
        <div key={game.id} className="game-card">
          <div className="game-info">
            <div className="team-info">
              <span className="team-name">{game.homeTeam}</span>
              {game.homeScore !== undefined && <span className="score">{game.homeScore}</span>}
            </div>
            <div className="game-time">{game.gameTime}</div>
            <div className="team-info">
              <span className="team-name">{game.awayTeam}</span>
              {game.awayScore !== undefined && <span className="score">{game.awayScore}</span>}
            </div>
          </div>
          {game.isLocked ? (
            <div className="game-status">Locked</div>
          ) : (
            <button className="pick-button">Make Pick</button>
          )}
        </div>
      ))}
    </div>
  );
};

export default GameList;
