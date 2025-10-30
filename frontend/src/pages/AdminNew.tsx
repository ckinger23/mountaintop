import { useState, useEffect } from 'react';
import { gamesService, adminService } from '../services/api';
import { Game, Team, Season, Week } from '../types';
import { useAuth } from '../hooks/useAuth';

// Game components
import GameCard from '../components/admin/games/GameCard';
import GameResultForm from '../components/admin/games/GameResultForm';
import CreateGameModal from '../components/admin/games/CreateGameModal';
import EditGameModal from '../components/admin/games/EditGameModal';

// Week components
import CreateWeekModal from '../components/admin/weeks/CreateWeekModal';

// Season components
import CreateSeasonModal from '../components/admin/seasons/CreateSeasonModal';

// Combined Seasons & Weeks component
import SeasonsWeeksManager from '../components/admin/SeasonsWeeksManager';

type Tab = 'results' | 'games' | 'seasons-weeks';

export default function AdminNew() {
  const [activeTab, setActiveTab] = useState<Tab>('results');
  const [games, setGames] = useState<Game[]>([]);
  const [teams, setTeams] = useState<Team[]>([]);
  const [seasons, setSeasons] = useState<Season[]>([]);
  const [weeks, setWeeks] = useState<Week[]>([]);
  const [currentWeek, setCurrentWeek] = useState<Week | null>(null);
  const [selectedWeekId, setSelectedWeekId] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState<number | null>(null);
  const { user } = useAuth();

  // Modal states
  const [createGameModal, setCreateGameModal] = useState(false);
  const [editGameModal, setEditGameModal] = useState<Game | null>(null);
  const [createWeekModal, setCreateWeekModal] = useState(false);
  const [createSeasonModal, setCreateSeasonModal] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [teamsData, seasonsData] = await Promise.all([
        gamesService.getTeams(),
        gamesService.getSeasons(),
      ]);
      setTeams(teamsData);
      setSeasons(seasonsData);

      try {
        const week = await gamesService.getCurrentWeek();
        setCurrentWeek(week);
        setSelectedWeekId(week.id);

        // Load all weeks for the active season
        if (week.season_id) {
          const weeksData = await gamesService.getWeeks(week.season_id);
          setWeeks(weeksData);
        }

        const gamesData = await gamesService.getGames(week.id);
        setGames(gamesData);
      } catch (error) {
        console.error('No current week found:', error);
      }
    } catch (error) {
      console.error('Error loading data:', error);
    } finally {
      setLoading(false);
    }
  };

  const loadGames = async () => {
    if (!selectedWeekId) return;
    try {
      const gamesData = await gamesService.getGames(selectedWeekId);
      setGames(gamesData);
    } catch (error) {
      console.error('Error loading games:', error);
    }
  };

  const handleWeekChange = async (weekId: number) => {
    setSelectedWeekId(weekId);
    try {
      const gamesData = await gamesService.getGames(weekId);
      setGames(gamesData);

      // Update currentWeek for the modals
      const week = weeks.find(w => w.id === weekId);
      if (week) {
        setCurrentWeek(week);
      }
    } catch (error) {
      console.error('Error loading games for week:', error);
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

  const handleDeleteGame = async (gameId: number) => {
    if (!confirm('Are you sure you want to delete this game?')) return;
    try {
      await adminService.deleteGame(gameId);
      alert('Game deleted successfully!');
      loadGames();
    } catch (error: any) {
      alert(error.response?.data || 'Failed to delete game');
    }
  };

  const handleCreateGameSuccess = () => {
    setCreateGameModal(false);
    loadGames();
  };

  const handleEditGameSuccess = () => {
    setEditGameModal(null);
    loadGames();
  };

  const handleCreateWeekSuccess = () => {
    setCreateWeekModal(false);
    loadData();
  };

  const handleCreateSeasonSuccess = () => {
    setCreateSeasonModal(false);
    loadData();
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
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-3xl font-bold">Admin Dashboard</h1>

        {/* Week Selector */}
        {weeks.length > 0 && (
          <div className="flex items-center gap-3">
            <label className="text-sm font-medium text-gray-700">Week:</label>
            <select
              value={selectedWeekId || ''}
              onChange={(e) => handleWeekChange(parseInt(e.target.value))}
              className="px-3 py-2 border rounded-md bg-white text-sm"
            >
              {weeks.map((week) => (
                <option key={week.id} value={week.id}>
                  {week.name} {week.id === currentWeek?.id ? '(Latest)' : ''}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="border-b mb-6">
        <nav className="flex space-x-8">
          <button
            onClick={() => setActiveTab('results')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'results'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Enter Results
          </button>
          <button
            onClick={() => setActiveTab('games')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'games'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Manage Games
          </button>
          <button
            onClick={() => setActiveTab('seasons-weeks')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'seasons-weeks'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Seasons & Weeks
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'results' && (
        <div className="space-y-4">
          {games.length === 0 ? (
            <div className="text-center text-gray-500 py-8">
              No games found for the current week.
            </div>
          ) : (
            games.map((game) => (
              <GameResultForm
                key={game.id}
                game={game}
                onUpdate={handleScoreUpdate}
                updating={updating === game.id}
              />
            ))
          )}
        </div>
      )}

      {activeTab === 'games' && (
        <div>
          <div className="mb-4">
            <button
              onClick={() => setCreateGameModal(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
            >
              Create New Game
            </button>
          </div>
          <div className="space-y-4">
            {games.map((game) => (
              <GameCard
                key={game.id}
                game={game}
                onEdit={() => setEditGameModal(game)}
                onDelete={() => handleDeleteGame(game.id)}
              />
            ))}
          </div>
        </div>
      )}

      {activeTab === 'seasons-weeks' && (
        <div>
          <div className="mb-4 flex gap-3">
            <button
              onClick={() => setCreateSeasonModal(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
            >
              Create New Season
            </button>
            <button
              onClick={() => setCreateWeekModal(true)}
              className="bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors"
            >
              Create New Week
            </button>
          </div>
          <SeasonsWeeksManager
            seasons={seasons}
            weeks={weeks}
            onWeekUpdate={loadData}
          />
        </div>
      )}

      {/* Modals */}
      <CreateGameModal
        isOpen={createGameModal}
        onClose={() => setCreateGameModal(false)}
        teams={teams}
        currentWeek={currentWeek}
        onSuccess={handleCreateGameSuccess}
      />

      {editGameModal && (
        <EditGameModal
          isOpen={true}
          onClose={() => setEditGameModal(null)}
          game={editGameModal}
          teams={teams}
          currentWeek={currentWeek}
          onSuccess={handleEditGameSuccess}
        />
      )}

      <CreateWeekModal
        isOpen={createWeekModal}
        onClose={() => setCreateWeekModal(false)}
        seasons={seasons}
        onSuccess={handleCreateWeekSuccess}
      />

      <CreateSeasonModal
        isOpen={createSeasonModal}
        onClose={() => setCreateSeasonModal(false)}
        onSuccess={handleCreateSeasonSuccess}
      />
    </div>
  );
}
