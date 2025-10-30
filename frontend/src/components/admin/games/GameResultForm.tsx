import { useState } from 'react';
import { Game } from '../../../types';

interface GameResultFormProps {
  game: Game;
  onUpdate: (gameId: number, homeScore: number, awayScore: number, isFinal: boolean) => void;
  updating: boolean;
}

export default function GameResultForm({ game, onUpdate, updating }: GameResultFormProps) {
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
          <div className="text-xs text-gray-500 mt-1">
            Spread: {game.home_team.abbreviation} {game.home_spread > 0 ? '+' : ''}{game.home_spread} | Total: {game.total}
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

      {/* Show total when scores are entered */}
      {homeScore && awayScore && (
        <div className="mt-3 text-sm text-center text-gray-600">
          Total: {parseInt(homeScore) + parseInt(awayScore)}
          {game.total && (
            <span className={`ml-2 font-semibold ${
              parseInt(homeScore) + parseInt(awayScore) > game.total ? 'text-green-600' : 'text-red-600'
            }`}>
              ({parseInt(homeScore) + parseInt(awayScore) > game.total ? 'Over' : 'Under'} {game.total})
            </span>
          )}
        </div>
      )}

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
