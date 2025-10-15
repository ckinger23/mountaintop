import { useState, useEffect } from 'react';
import { gamesService, adminService } from '../services/api';
import { Game } from '../types';
import { useAuth } from '../hooks/useAuth';

export default function Admin() {
  const [games, setGames] = useState<Game[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<number | null>(null);
  const { user } = useAuth();

  useEffect(() => {
    loadGames();
  }, []);

  const loadGames = async () => {
    try {
      const week = await gamesService.getCurrentWeek();
      const gamesData = await gamesService.getGames(week.id);
      setGames(gamesData);
    } catch (error) {
      console.error('Error loading games:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleScoreUpdate = async (
    gameId: number,
    homeScore: number,
    awayScore: number,
    isFinal: boolean
  ) => {
    setUpdating(gameId);
    try {
      await adminService.updateGameResult(gameId, homeScore, awayScore, isFinal);
      alert('Game result updated successfully!');
      loadGames();
    } catch (error: any) {
      alert(error.response?.data || 'Failed to update game');
    } finally {
      setUpdating(null);
    }
  };

  if (!user?.is_admin) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
          You do not have permission to access this page.
        </div>
      </div>
    );
  }

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  return (
    <div className="max-w-6xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Admin - Enter Game Results</h1>

      <div className="space-y-4">
        {games.map((game) => (
          <GameResultForm
            key={game.id}
            game={game}
            onUpdate={handleScoreUpdate}
            updating={updating === game.id}
          />
        ))}
      </div>
    </div>
  );
}

interface GameResultFormProps {
  game: Game;
  onUpdate: (gameId: number, homeScore: number, awayScore: number, isFinal: boolean) => void;
  updating: boolean;
}

function GameResultForm({ game, onUpdate, updating }: GameResultFormProps) {
  const [homeScore, setHomeScore] = useState(game.home_score?.toString() || '');
  const [awayScore, setAwayScore] = useState(game.away_score?.toString() || '');
  const [isFinal, setIsFinal] = useState(game.is_final);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onUpdate(game.id, parseInt(homeScore), parseInt(awayScore), isFinal);
  };

  return (
    <form onSubmit={handleSubmit} className="bg-white border rounded-lg p-4">
      <div className="flex items-center justify-between mb-4">
        <div>
          <div className="text-sm text-gray-500">
            {new Date(game.game_time).toLocaleString()}
          </div>
        </div>
        {game.is_final && (
          <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
            FINAL
          </span>
        )}
      </div>

      <div className="grid grid-cols-3 gap-4 items-center">
        {/* Away Team */}
        <div>
          <div className="font-semibold">{game.away_team.name}</div>
          <input
            type="number"
            value={awayScore}
            onChange={(e) => setAwayScore(e.target.value)}
            placeholder="Score"
            className="mt-2 w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        {/* VS */}
        <div className="text-center text-gray-500 font-bold">VS</div>

        {/* Home Team */}
        <div>
          <div className="font-semibold">{game.home_team.name}</div>
          <input
            type="number"
            value={homeScore}
            onChange={(e) => setHomeScore(e.target.value)}
            placeholder="Score"
            className="mt-2 w-full px-3 py-2 border rounded-md"
            required
          />
        </div>
      </div>

      <div className="mt-4 flex items-center justify-between">
        <label className="flex items-center">
          <input
            type="checkbox"
            checked={isFinal}
            onChange={(e) => setIsFinal(e.target.checked)}
            className="mr-2"
          />
          <span className="text-sm">Mark as Final</span>
        </label>

        <button
          type="submit"
          disabled={updating}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-400"
        >
          {updating ? 'Updating...' : 'Update Result'}
        </button>
      </div>
    </form>
  );
}
