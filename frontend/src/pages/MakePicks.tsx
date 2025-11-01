import { useState, useEffect } from 'react';
import { gamesService, picksService } from '../services/api';
import { Game } from '../types';
import { useLeague } from '../hooks/useLeague';

interface GamePick {
  teamId: number;
  overUnder: string; // "over" or "under"
}

export default function MakePicks() {
  const { currentLeague } = useLeague();
  const [games, setGames] = useState<Game[]>([]);
  const [picks, setPicks] = useState<Map<number, GamePick>>(new Map());
  const [loading, setLoading] = useState(true);

  // useEffect() hook lets you perform side effects in React functional components
  // Side effects are operations that interact with things outside the component's
  // render logic, like API calls, subscriptions, timers, DOM manipulation
  // Effect function runs after component renders
  // cleanup function (empty in this scenario) runs before the effect runs again
  // or when component unmounts
  // Dependency array controls when the effect re-runs
  // Dependency array: [] -> run once on mount
  // [a, b] -> run on mount and when a or b changes
  // No list -> run on every render (usually avoid this)
  useEffect(() => {
    // load the data from API after page renders or when league changes
    loadData();
  }, [currentLeague]);

  const loadData = async () => {
    if (!currentLeague) {
      setLoading(false);
      return;
    }

    try {
      // Get current week
      const week = await gamesService.getCurrentWeek();

      // Get games for current week
      const gamesData = await gamesService.getGames(week.id);
      setGames(gamesData);

      // Get user's existing picks for this league
      const picksData = await picksService.getMyPicks(week.id, currentLeague.id);

      // Pre-populate picks map with existing picks
      const picksMap = new Map<number, GamePick>();
      picksData.forEach((pick) => {
        picksMap.set(pick.game_id, {
          teamId: pick.picked_team_id,
          overUnder: pick.picked_over_under || 'over',
        });
      });
      setPicks(picksMap);
    } catch (error) {
      console.error('Error loading data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTeamPickChange = (gameId: number, teamId: number) => {
    const newPicks = new Map(picks);
    const existingPick = newPicks.get(gameId);
    newPicks.set(gameId, {
      teamId,
      overUnder: existingPick?.overUnder || 'over',
    });
    setPicks(newPicks);
  };

  const handleOverUnderChange = (gameId: number, overUnder: string) => {
    const newPicks = new Map(picks);
    const existingPick = newPicks.get(gameId);
    if (existingPick) {
      newPicks.set(gameId, {
        ...existingPick,
        overUnder,
      });
      setPicks(newPicks);
    }
  };

  const submitAllPicks = async () => {
    if (!currentLeague) {
      alert('Please select a league first');
      return;
    }

    // Validate all games have picks
    const unlocked = games.filter(game => !isPickLocked(game));
    const missingPicks = unlocked.filter(game => !picks.has(game.id));

    if (missingPicks.length > 0) {
      alert(`Please make picks for all games. Missing ${missingPicks.length} pick(s).`);
      return;
    }

    try {
      // Submit all picks
      const promises = Array.from(picks.entries()).map(([gameId, pick]) =>
        picksService.submitPick(currentLeague.id, gameId, pick.teamId, pick.overUnder)
      );

      await Promise.all(promises);
      alert('All picks saved successfully!');
      loadData(); // Reload to show updated picks
    } catch (error: any) {
      alert(error.response?.data || 'Failed to save picks');
    }
  };

  const isPickLocked = (game: Game): boolean => {
    return new Date(game.game_time) < new Date() || game.is_final;
  };

  if (loading) {
    return <div className="text-center py-8">Loading...</div>;
  }

  if (!currentLeague) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <h1 className="text-3xl font-bold mb-6">Make Your Picks</h1>
        <div className="text-center py-8 text-gray-500">
          Please select or create a league to start making picks
        </div>
      </div>
    );
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

                {/* Spread Pick */}
                <div className="mb-4">
                  <div className="text-sm font-medium text-gray-700 mb-2">Pick Winner (Spread)</div>
                  <div className="grid grid-cols-2 gap-4">
                    {/* Away Team */}
                    <button
                      onClick={() => !locked && handleTeamPickChange(game.id, game.away_team_id)}
                      disabled={locked}
                      className={`p-4 rounded-lg border-2 transition-colors ${
                        currentPick?.teamId === game.away_team_id
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
                      onClick={() => !locked && handleTeamPickChange(game.id, game.home_team_id)}
                      disabled={locked}
                      className={`p-4 rounded-lg border-2 transition-colors ${
                        currentPick?.teamId === game.home_team_id
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
                </div>

                {/* Over/Under Pick */}
                <div>
                  <div className="text-sm font-medium text-gray-700 mb-2">
                    Total Score: {game.total} points
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <button
                      onClick={() => !locked && currentPick && handleOverUnderChange(game.id, 'over')}
                      disabled={locked || !currentPick}
                      className={`p-3 rounded-lg border-2 transition-colors ${
                        currentPick?.overUnder === 'over'
                          ? 'border-green-500 bg-green-50'
                          : 'border-gray-200 hover:border-gray-300'
                      } ${locked || !currentPick ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'}`}
                    >
                      <div className="font-semibold text-center">Over {game.total}</div>
                    </button>
                    <button
                      onClick={() => !locked && currentPick && handleOverUnderChange(game.id, 'under')}
                      disabled={locked || !currentPick}
                      className={`p-3 rounded-lg border-2 transition-colors ${
                        currentPick?.overUnder === 'under'
                          ? 'border-green-500 bg-green-50'
                          : 'border-gray-200 hover:border-gray-300'
                      } ${locked || !currentPick ? 'cursor-not-allowed opacity-50' : 'cursor-pointer'}`}
                    >
                      <div className="font-semibold text-center">Under {game.total}</div>
                    </button>
                  </div>
                </div>
              </div>
            );
          })}

          {/* Submit All Picks Button */}
          {games.some(game => !isPickLocked(game)) && (
            <div className="sticky bottom-0 bg-white border-t pt-4 mt-6">
              <button
                onClick={submitAllPicks}
                className="w-full bg-blue-600 text-white py-3 rounded-lg hover:bg-blue-700 transition-colors font-semibold text-lg"
              >
                Submit All Picks
              </button>
              <p className="text-sm text-gray-500 text-center mt-2">
                Make sure to pick a team for every game before submitting
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
