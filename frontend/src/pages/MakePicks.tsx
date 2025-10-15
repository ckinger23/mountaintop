import { useState, useEffect } from 'react';
import { gamesService, picksService } from '../services/api';
import { Game, Pick } from '../types';

export default function MakePicks() {
  const [games, setGames] = useState<Game[]>([]);
  const [picks, setPicks] = useState<Map<number, number>>(new Map());
  const [existingPicks, setExistingPicks] = useState<Pick[]>([]);
  const [loading, setLoading] = useState(true);
  const [currentWeekId, setCurrentWeekId] = useState<number | null>(null);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      // Get current week
      const week = await gamesService.getCurrentWeek();
      setCurrentWeekId(week.id);

      // Get games for current week
      const gamesData = await gamesService.getGames(week.id);
      setGames(gamesData);

      // Get user's existing picks
      const picksData = await picksService.getMyPicks(week.id);
      setExistingPicks(picksData);

      // Pre-populate picks map
      const picksMap = new Map<number, number>();
      picksData.forEach((pick) => {
        picksMap.set(pick.game_id, pick.picked_team_id);
      });
      setPicks(picksMap);
    } catch (error) {
      console.error('Error loading data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handlePickChange = (gameId: number, teamId: number) => {
    const newPicks = new Map(picks);
    newPicks.set(gameId, teamId);
    setPicks(newPicks);
  };

  const submitPick = async (gameId: number) => {
    const pickedTeamId = picks.get(gameId);
    if (!pickedTeamId) return;

    try {
      await picksService.submitPick(gameId, pickedTeamId);
      alert('Pick saved successfully!');
      loadData(); // Reload to show updated picks
    } catch (error: any) {
      alert(error.response?.data || 'Failed to save pick');
    }
  };

  const isPickLocked = (game: Game): boolean => {
    return new Date(game.game_time) < new Date() || game.is_final;
  };

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Make Your Picks</h1>

      {games.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          No games available for picking
        </div>
      ) : (
        <div className="space-y-4">
          {games.map((game) => {
            const locked = isPickLocked(game);
            const currentPick = picks.get(game.id);

            return (
              <div
                key={game.id}
                className={`border rounded-lg p-4 ${locked ? 'bg-gray-50' : 'bg-white'}`}
              >
                <div className="flex justify-between items-center mb-3">
                  <div className="text-sm text-gray-500">
                    {new Date(game.game_time).toLocaleString()}
                  </div>
                  {locked && (
                    <span className="text-xs bg-gray-200 px-2 py-1 rounded">
                      LOCKED
                    </span>
                  )}
                  {game.is_final && (
                    <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                      FINAL
                    </span>
                  )}
                </div>

                <div className="grid grid-cols-2 gap-4">
                  {/* Away Team */}
                  <button
                    onClick={() => !locked && handlePickChange(game.id, game.away_team_id)}
                    disabled={locked}
                    className={`p-4 rounded-lg border-2 transition-colors ${
                      currentPick === game.away_team_id
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    } ${locked ? 'cursor-not-allowed' : 'cursor-pointer'}`}
                  >
                    <div className="font-semibold">{game.away_team.name}</div>
                    <div className="text-sm text-gray-500">
                      {game.away_team.abbreviation}
                    </div>
                    {game.is_final && (
                      <div className="text-lg font-bold mt-2">{game.away_score}</div>
                    )}
                  </button>

                  {/* Home Team */}
                  <button
                    onClick={() => !locked && handlePickChange(game.id, game.home_team_id)}
                    disabled={locked}
                    className={`p-4 rounded-lg border-2 transition-colors ${
                      currentPick === game.home_team_id
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    } ${locked ? 'cursor-not-allowed' : 'cursor-pointer'}`}
                  >
                    <div className="font-semibold">{game.home_team.name}</div>
                    <div className="text-sm text-gray-500">
                      {game.home_team.abbreviation}
                      {game.home_spread !== 0 && (
                        <span className="ml-2">
                          ({game.home_spread > 0 ? '+' : ''}
                          {game.home_spread})
                        </span>
                      )}
                    </div>
                    {game.is_final && (
                      <div className="text-lg font-bold mt-2">{game.home_score}</div>
                    )}
                  </button>
                </div>

                {!locked && currentPick && (
                  <div className="mt-3">
                    <button
                      onClick={() => submitPick(game.id)}
                      className="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 transition-colors"
                    >
                      Save Pick
                    </button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
