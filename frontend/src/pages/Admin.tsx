import { useState, useEffect } from 'react';
import toast from 'react-hot-toast';
import { gamesService, adminService } from '../services/api';
import { Game, Team, Season, Week, WeekStatus } from '../types';
import { useAuth } from '../hooks/useAuth';
import { useLeague } from '../hooks/useLeague';
import { ChevronDown, ChevronRight, Plus, Edit, Save } from 'lucide-react';

// Game components
import GameResultForm from '../components/admin/games/GameResultForm';
import CreateGameModal from '../components/admin/games/CreateGameModal';
import EditGameModal from '../components/admin/games/EditGameModal';

// Week components
import CreateWeekModal from '../components/admin/weeks/CreateWeekModal';
import UpdateWeekModal from '../components/admin/weeks/UpdateWeekModal';

// Season components
import CreateSeasonModal from '../components/admin/seasons/CreateSeasonModal';

export default function Admin() {
  const { currentLeague } = useLeague();
  const [seasons, setSeasons] = useState<Season[]>([]);
  const [weeksBySeason, setWeeksBySeason] = useState<Record<number, Week[]>>({});
  const [gamesByWeek, setGamesByWeek] = useState<Record<number, Game[]>>({});
  const [teams, setTeams] = useState<Team[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<number | null>(null);
  const { user } = useAuth();

  // Collapsible state
  const [expandedSeasons, setExpandedSeasons] = useState<Set<number>>(new Set());
  const [expandedWeeks, setExpandedWeeks] = useState<Set<number>>(new Set());

  // Filter toggles
  const [showOnlyActiveSeasons, setShowOnlyActiveSeasons] = useState(false);
  const [showOnlyActiveWeeks, setShowOnlyActiveWeeks] = useState(true);

  // Modal states
  const [createSeasonModal, setCreateSeasonModal] = useState(false);
  const [createWeekModal, setCreateWeekModal] = useState<{ seasonId: number } | null>(null);
  const [updateWeekModal, setUpdateWeekModal] = useState<Week | null>(null);
  const [createGameModal, setCreateGameModal] = useState<{ weekId: number; week: Week } | null>(null);
  const [editGameModal, setEditGameModal] = useState<Game | null>(null);

  useEffect(() => {
    loadData();
  }, [currentLeague]);

  const loadData = async () => {
    if (!currentLeague) {
      setLoading(false);
      return;
    }

    try {
      const [teamsData, seasonsData] = await Promise.all([
        gamesService.getTeams(),
        gamesService.getSeasons(),
      ]);
      setTeams(teamsData);

      // Filter seasons by current league
      const leagueSeasons = seasonsData.filter(s => s.league_id === currentLeague.id);
      setSeasons(leagueSeasons);

      // Load weeks for league seasons only
      const weeksPromises = leagueSeasons.map(async (season) => {
        const weeks = await gamesService.getWeeks(season.id);
        return { seasonId: season.id, weeks };
      });

      const weeksResults = await Promise.all(weeksPromises);
      const weeksMap: Record<number, Week[]> = {};
      weeksResults.forEach(({ seasonId, weeks }) => {
        weeksMap[seasonId] = weeks;
      });
      setWeeksBySeason(weeksMap);

      // Auto-expand seasons with active weeks and expand those weeks
      const newExpandedSeasons = new Set<number>();
      const newExpandedWeeks = new Set<number>();

      leagueSeasons.forEach((season) => {
        const weeks = weeksMap[season.id] || [];
        const hasActiveWeeks = weeks.some(
          (w) => w.status === 'creating' || w.status === 'picking' || w.status === 'scoring'
        );

        if (hasActiveWeeks || season.is_active) {
          newExpandedSeasons.add(season.id);

          // Auto-expand creating/scoring weeks (prioritized)
          weeks.forEach((week) => {
            if (week.status === 'creating' || week.status === 'scoring') {
              newExpandedWeeks.add(week.id);
            }
          });
        }
      });

      setExpandedSeasons(newExpandedSeasons);
      setExpandedWeeks(newExpandedWeeks);

      // Load games for auto-expanded weeks
      const gamesPromises = Array.from(newExpandedWeeks).map(async (weekId) => {
        const games = await gamesService.getGames(weekId);
        return { weekId, games };
      });

      const gamesResults = await Promise.all(gamesPromises);
      const gamesMap: Record<number, Game[]> = {};
      gamesResults.forEach(({ weekId, games }) => {
        gamesMap[weekId] = games;
      });
      setGamesByWeek(gamesMap);
    } catch (error) {
      console.error('Error loading data:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadGamesForWeek = async (weekId: number) => {
    try {
      const games = await gamesService.getGames(weekId);
      setGamesByWeek((prev) => ({ ...prev, [weekId]: games }));
    } catch (error) {
      console.error('Error loading games:', error);
    }
  };

  const toggleSeason = async (seasonId: number) => {
    const newExpanded = new Set(expandedSeasons);
    if (newExpanded.has(seasonId)) {
      newExpanded.delete(seasonId);
    } else {
      newExpanded.add(seasonId);
    }
    setExpandedSeasons(newExpanded);
  };

  const toggleWeek = async (weekId: number) => {
    const newExpanded = new Set(expandedWeeks);
    if (newExpanded.has(weekId)) {
      newExpanded.delete(weekId);
    } else {
      newExpanded.add(weekId);
      // Load games when expanding
      if (!gamesByWeek[weekId]) {
        await loadGamesForWeek(weekId);
      }
    }
    setExpandedWeeks(newExpanded);
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
      toast.success('Game result updated successfully!');
      // Reload games for the week
      const game = Object.values(gamesByWeek)
        .flat()
        .find((g) => g.id === gameId);
      if (game) {
        await loadGamesForWeek(game.week_id);
      }
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to update game');
    } finally {
      setUpdating(null);
    }
  };

  const handleDeleteGame = async (gameId: number, weekId: number) => {
    if (!confirm('Are you sure you want to delete this game?')) return;
    try {
      await adminService.deleteGame(gameId);
      toast.success('Game deleted successfully!');
      await loadGamesForWeek(weekId);
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to delete game');
    }
  };

  const getStatusBadge = (status: WeekStatus) => {
    const styles: Record<WeekStatus, string> = {
      creating: 'bg-yellow-100 text-yellow-800',
      picking: 'bg-blue-100 text-blue-800',
      scoring: 'bg-purple-100 text-purple-800',
      finished: 'bg-gray-100 text-gray-600',
    };
    return (
      <span className={`px-2 py-1 rounded-full text-xs font-medium ${styles[status]}`}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </span>
    );
  };

  const getWeekActions = (week: Week) => {
    switch (week.status) {
      case 'creating':
        return (
          <div className="flex gap-2">
            <button
              onClick={() => setCreateGameModal({ weekId: week.id, week })}
              className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition-colors"
            >
              <Plus size={16} /> Add Game
            </button>
            <button
              onClick={() => setUpdateWeekModal(week)}
              className="flex items-center gap-1 px-3 py-1.5 bg-gray-600 text-white text-sm rounded hover:bg-gray-700 transition-colors"
            >
              <Edit size={16} /> Update Week
            </button>
          </div>
        );
      case 'picking':
        return (
          <button
            onClick={() => setUpdateWeekModal(week)}
            className="flex items-center gap-1 px-3 py-1.5 bg-gray-600 text-white text-sm rounded hover:bg-gray-700 transition-colors"
          >
            <Edit size={16} /> Update Week
          </button>
        );
      case 'scoring':
        return (
          <button
            onClick={() => setUpdateWeekModal(week)}
            className="flex items-center gap-1 px-3 py-1.5 bg-green-600 text-white text-sm rounded hover:bg-green-700 transition-colors"
          >
            <Save size={16} /> Complete Week
          </button>
        );
      case 'finished':
        return null;
      default:
        return null;
    }
  };

  const getGameActions = (game: Game, week: Week) => {
    if (week.status === 'creating') {
      return (
        <div className="flex gap-2">
          <button
            onClick={() => setEditGameModal(game)}
            className="px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition-colors"
          >
            Edit
          </button>
          <button
            onClick={() => handleDeleteGame(game.id, week.id)}
            className="px-3 py-1.5 bg-red-600 text-white text-sm rounded hover:bg-red-700 transition-colors"
          >
            Delete
          </button>
        </div>
      );
    }

    if (week.status === 'scoring') {
      return (
        <GameResultForm
          game={game}
          onUpdate={handleScoreUpdate}
          updating={updating === game.id}
        />
      );
    }

    return null;
  };

  const filteredSeasons = showOnlyActiveSeasons
    ? seasons.filter((s) => s.is_active)
    : seasons;

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

  if (!currentLeague) {
    return (
      <div className="max-w-4xl mx-auto p-6">
        <h1 className="text-3xl font-bold mb-6">Admin Dashboard</h1>
        <div className="text-center py-8 text-gray-500">
          Please select or create a league to manage seasons, weeks, and games
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto p-6">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-4">Admin Dashboard</h1>

        {/* Filter Toggles */}
        <div className="flex items-center gap-6 mb-4">
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={showOnlyActiveSeasons}
              onChange={(e) => setShowOnlyActiveSeasons(e.target.checked)}
              className="rounded"
            />
            <span>Show only active seasons</span>
          </label>
          <label className="flex items-center gap-2 text-sm">
            <input
              type="checkbox"
              checked={showOnlyActiveWeeks}
              onChange={(e) => setShowOnlyActiveWeeks(e.target.checked)}
              className="rounded"
            />
            <span>Show only active weeks</span>
          </label>
        </div>

        {/* Create Season Button */}
        <button
          onClick={() => setCreateSeasonModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors"
        >
          <Plus size={18} /> Create Season
        </button>
      </div>

      {/* Hierarchical View */}
      <div className="space-y-4">
        {filteredSeasons.map((season) => {
          const weeks = weeksBySeason[season.id] || [];
          const filteredWeeks = showOnlyActiveWeeks
            ? weeks.filter((w) => w.status !== 'finished')
            : weeks;

          return (
            <div key={season.id} className="border rounded-lg bg-white shadow-sm">
              {/* Season Header */}
              <div
                className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
                onClick={() => toggleSeason(season.id)}
              >
                <div className="flex items-center gap-3">
                  {expandedSeasons.has(season.id) ? (
                    <ChevronDown size={20} />
                  ) : (
                    <ChevronRight size={20} />
                  )}
                  <h2 className="text-xl font-semibold">
                    {season.name} ({season.year})
                  </h2>
                  {season.is_active && (
                    <span className="px-2 py-1 bg-green-100 text-green-800 rounded-full text-xs font-medium">
                      Active
                    </span>
                  )}
                </div>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    setCreateWeekModal({ seasonId: season.id });
                  }}
                  className="flex items-center gap-1 px-3 py-1.5 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition-colors"
                >
                  <Plus size={16} /> Add Week
                </button>
              </div>

              {/* Weeks */}
              {expandedSeasons.has(season.id) && (
                <div className="border-t">
                  {filteredWeeks.length === 0 ? (
                    <div className="p-4 text-center text-gray-500">
                      No weeks found. Create one to get started.
                    </div>
                  ) : (
                    <div className="divide-y">
                      {filteredWeeks.map((week) => {
                        const games = gamesByWeek[week.id] || [];
                        const isExpanded = expandedWeeks.has(week.id);

                        return (
                          <div key={week.id} className="bg-gray-50">
                            {/* Week Header */}
                            <div className="flex items-center justify-between p-4">
                              <div
                                className="flex items-center gap-3 cursor-pointer flex-1"
                                onClick={() => toggleWeek(week.id)}
                              >
                                {isExpanded ? (
                                  <ChevronDown size={18} />
                                ) : (
                                  <ChevronRight size={18} />
                                )}
                                <h3 className="text-lg font-medium">{week.name}</h3>
                                {getStatusBadge(week.status)}
                                {week.pick_deadline && (
                                  <span className="text-sm text-gray-600">
                                    Deadline: {new Date(week.pick_deadline).toLocaleString()}
                                  </span>
                                )}
                              </div>
                              {getWeekActions(week)}
                            </div>

                            {/* Games */}
                            {isExpanded && (
                              <div className="px-4 pb-4 space-y-3">
                                {games.length === 0 ? (
                                  <div className="text-center text-gray-500 py-4 bg-white rounded border">
                                    No games yet.
                                    {week.status === 'creating' && ' Click "Add Game" to create one.'}
                                  </div>
                                ) : (
                                  games.map((game) => (
                                    <div
                                      key={game.id}
                                      className="bg-white p-4 rounded-lg border shadow-sm"
                                    >
                                      <div className="flex items-center justify-between">
                                        <div className="flex-1">
                                          <div className="flex items-center gap-4 mb-2">
                                            <div className="flex items-center gap-2">
                                              <span className="font-medium">{game.away_team.name}</span>
                                              <span className="text-gray-500">@</span>
                                              <span className="font-medium">{game.home_team.name}</span>
                                            </div>
                                            <span className="text-sm text-gray-600">
                                              Spread: {game.home_spread > 0 ? '+' : ''}
                                              {game.home_spread}
                                            </span>
                                            <span className="text-sm text-gray-600">
                                              Total: {game.total}
                                            </span>
                                          </div>
                                          <div className="text-sm text-gray-500">
                                            {new Date(game.game_time).toLocaleString()}
                                          </div>
                                          {game.is_final && (
                                            <div className="mt-2 text-sm font-medium">
                                              Final Score: {game.away_team.abbreviation} {game.away_score} -{' '}
                                              {game.home_team.abbreviation} {game.home_score}
                                            </div>
                                          )}
                                        </div>
                                        <div>{getGameActions(game, week)}</div>
                                      </div>
                                    </div>
                                  ))
                                )}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Modals */}
      <CreateSeasonModal
        isOpen={createSeasonModal}
        onClose={() => setCreateSeasonModal(false)}
        onSuccess={() => {
          setCreateSeasonModal(false);
          loadData();
        }}
      />

      {createWeekModal && (
        <CreateWeekModal
          isOpen={true}
          onClose={() => setCreateWeekModal(null)}
          seasons={seasons}
          defaultSeasonId={createWeekModal.seasonId}
          onSuccess={() => {
            setCreateWeekModal(null);
            loadData();
          }}
        />
      )}

      {updateWeekModal && (
        <UpdateWeekModal
          isOpen={true}
          onClose={() => setUpdateWeekModal(null)}
          week={updateWeekModal}
          onSuccess={() => {
            setUpdateWeekModal(null);
            loadData();
          }}
        />
      )}

      {createGameModal && (
        <CreateGameModal
          isOpen={true}
          onClose={() => setCreateGameModal(null)}
          teams={teams}
          currentWeek={createGameModal.week}
          onSuccess={() => {
            setCreateGameModal(null);
            if (createGameModal) {
              loadGamesForWeek(createGameModal.weekId);
            }
          }}
        />
      )}

      {editGameModal && (
        <EditGameModal
          isOpen={true}
          onClose={() => setEditGameModal(null)}
          game={editGameModal}
          teams={teams}
          currentWeek={
            Object.values(weeksBySeason)
              .flat()
              .find((w) => w.id === editGameModal.week_id) || null
          }
          onSuccess={() => {
            if (editGameModal) {
              loadGamesForWeek(editGameModal.week_id);
            }
            setEditGameModal(null);
          }}
        />
      )}
    </div>
  );
}
