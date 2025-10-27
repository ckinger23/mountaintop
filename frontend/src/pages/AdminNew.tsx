import { useState, useEffect } from 'react';
import { gamesService, adminService } from '../services/api';
import { Game, Team, Season, Week } from '../types';
import { useAuth } from '../hooks/useAuth';
import Modal from '../components/Modal';

type Tab = 'results' | 'games' | 'weeks' | 'seasons';

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
            onClick={() => setActiveTab('weeks')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'weeks'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Manage Weeks
          </button>
          <button
            onClick={() => setActiveTab('seasons')}
            className={`py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
              activeTab === 'seasons'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            Manage Seasons
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

      {activeTab === 'weeks' && (
        <div>
          <div className="mb-4">
            <button
              onClick={() => setCreateWeekModal(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
            >
              Create New Week
            </button>
          </div>
          <div className="space-y-4">
            {weeks.length === 0 ? (
              <div className="text-center text-gray-500 py-8">
                No weeks found. Create one to get started.
              </div>
            ) : (
              weeks.map((week) => (
                <div key={week.id} className="bg-white border rounded-lg p-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <div className="font-semibold text-lg">{week.name}</div>
                      <div className="text-sm text-gray-500 mt-1">
                        Week {week.week_number} • Lock Time: {new Date(week.lock_time).toLocaleString()}
                      </div>
                      {week.season && (
                        <div className="text-xs text-gray-400 mt-1">
                          Season: {week.season.name}
                        </div>
                      )}
                    </div>
                    {week.id === currentWeek?.id && (
                      <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                        LATEST
                      </span>
                    )}
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      {activeTab === 'seasons' && (
        <div>
          <div className="mb-4">
            <button
              onClick={() => setCreateSeasonModal(true)}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
            >
              Create New Season
            </button>
          </div>
          <div className="space-y-4">
            {seasons.map((season) => (
              <div key={season.id} className="bg-white border rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-semibold text-lg">{season.name}</div>
                    <div className="text-sm text-gray-500">Year: {season.year}</div>
                  </div>
                  {season.is_active && (
                    <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                      ACTIVE
                    </span>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Modals */}
      <CreateGameModal
        isOpen={createGameModal}
        onClose={() => setCreateGameModal(false)}
        teams={teams}
        currentWeek={currentWeek}
        onSuccess={() => {
          setCreateGameModal(false);
          loadGames();
        }}
      />

      {editGameModal && (
        <EditGameModal
          isOpen={true}
          onClose={() => setEditGameModal(null)}
          game={editGameModal}
          teams={teams}
          currentWeek={currentWeek}
          onSuccess={() => {
            setEditGameModal(null);
            loadGames();
          }}
        />
      )}

      <CreateWeekModal
        isOpen={createWeekModal}
        onClose={() => setCreateWeekModal(false)}
        seasons={seasons}
        onSuccess={() => {
          setCreateWeekModal(false);
          loadData();
        }}
      />

      <CreateSeasonModal
        isOpen={createSeasonModal}
        onClose={() => setCreateSeasonModal(false)}
        onSuccess={() => {
          setCreateSeasonModal(false);
          loadData();
        }}
      />
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

// GameCard Component
interface GameCardProps {
  game: Game;
  onEdit: () => void;
  onDelete: () => void;
}

function GameCard({ game, onEdit, onDelete }: GameCardProps) {
  return (
    <div className="bg-white border rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="text-sm text-gray-500">
          {new Date(game.game_time).toLocaleString()}
        </div>
        <div className="flex gap-2">
          <button
            onClick={onEdit}
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            Edit
          </button>
          <button
            onClick={onDelete}
            className="text-red-600 hover:text-red-800 text-sm font-medium"
          >
            Delete
          </button>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 items-center">
        <div className="text-right">
          <div className="font-semibold">{game.away_team.name}</div>
          <div className="text-sm text-gray-500">{game.away_team.abbreviation}</div>
        </div>

        <div className="text-center text-gray-500 font-bold">@</div>

        <div className="text-left">
          <div className="font-semibold">{game.home_team.name}</div>
          <div className="text-sm text-gray-500">{game.home_team.abbreviation}</div>
        </div>
      </div>

      <div className="mt-3 text-xs text-gray-500 text-center">
        Spread: {game.home_team.abbreviation} {game.home_spread > 0 ? '+' : ''}{game.home_spread} | Total: {game.total}
      </div>
    </div>
  );
}

// Create Game Modal
interface CreateGameModalProps {
  isOpen: boolean;
  onClose: () => void;
  teams: Team[];
  currentWeek: Week | null;
  onSuccess: () => void;
}

function CreateGameModal({ isOpen, onClose, teams, currentWeek, onSuccess }: CreateGameModalProps) {
  const [homeTeamId, setHomeTeamId] = useState('');
  const [awayTeamId, setAwayTeamId] = useState('');
  const [gameTime, setGameTime] = useState('');
  const [homeSpread, setHomeSpread] = useState('');
  const [total, setTotal] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentWeek) {
      alert('No current week found. Please create a week first.');
      return;
    }

    setSubmitting(true);
    try {
      // Convert datetime-local format to ISO 8601
      const isoGameTime = new Date(gameTime).toISOString();

      await adminService.createGame(
        currentWeek.id,
        parseInt(homeTeamId),
        parseInt(awayTeamId),
        isoGameTime,
        parseFloat(homeSpread),
        parseFloat(total)
      );
      alert('Game created successfully!');
      onSuccess();
    } catch (error: any) {
      alert(error.response?.data || 'Failed to create game');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Create New Game">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Away Team
          </label>
          <select
            value={awayTeamId}
            onChange={(e) => setAwayTeamId(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          >
            <option value="">Select away team</option>
            {teams.map((team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Home Team
          </label>
          <select
            value={homeTeamId}
            onChange={(e) => setHomeTeamId(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          >
            <option value="">Select home team</option>
            {teams.map((team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Game Time
          </label>
          <input
            type="datetime-local"
            value={gameTime}
            onChange={(e) => setGameTime(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Home Spread (negative = favorite)
          </label>
          <input
            type="number"
            step="0.5"
            value={homeSpread}
            onChange={(e) => setHomeSpread(e.target.value)}
            placeholder="e.g., -7 or 3.5"
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Total (Over/Under)
          </label>
          <input
            type="number"
            step="0.5"
            value={total}
            onChange={(e) => setTotal(e.target.value)}
            placeholder="e.g., 52.5"
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div className="flex gap-3 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={submitting}
            className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-400"
          >
            {submitting ? 'Creating...' : 'Create Game'}
          </button>
        </div>
      </form>
    </Modal>
  );
}

// Edit Game Modal
interface EditGameModalProps {
  isOpen: boolean;
  onClose: () => void;
  game: Game;
  teams: Team[];
  currentWeek: Week | null;
  onSuccess: () => void;
}

function EditGameModal({ isOpen, onClose, game, teams, currentWeek, onSuccess }: EditGameModalProps) {
  const [homeTeamId, setHomeTeamId] = useState(game.home_team_id.toString());
  const [awayTeamId, setAwayTeamId] = useState(game.away_team_id.toString());
  const [gameTime, setGameTime] = useState(
    new Date(game.game_time).toISOString().slice(0, 16)
  );
  const [homeSpread, setHomeSpread] = useState(game.home_spread.toString());
  const [total, setTotal] = useState(game.total.toString());
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentWeek) {
      alert('No current week found.');
      return;
    }

    setSubmitting(true);
    try {
      // Convert datetime-local format to ISO 8601
      const isoGameTime = new Date(gameTime).toISOString();

      await adminService.updateGame(
        game.id,
        currentWeek.id,
        parseInt(homeTeamId),
        parseInt(awayTeamId),
        isoGameTime,
        parseFloat(homeSpread),
        parseFloat(total)
      );
      alert('Game updated successfully!');
      onSuccess();
    } catch (error: any) {
      alert(error.response?.data || 'Failed to update game');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Edit Game">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Away Team
          </label>
          <select
            value={awayTeamId}
            onChange={(e) => setAwayTeamId(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          >
            {teams.map((team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Home Team
          </label>
          <select
            value={homeTeamId}
            onChange={(e) => setHomeTeamId(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          >
            {teams.map((team) => (
              <option key={team.id} value={team.id}>
                {team.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Game Time
          </label>
          <input
            type="datetime-local"
            value={gameTime}
            onChange={(e) => setGameTime(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Home Spread (negative = favorite)
          </label>
          <input
            type="number"
            step="0.5"
            value={homeSpread}
            onChange={(e) => setHomeSpread(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Total (Over/Under)
          </label>
          <input
            type="number"
            step="0.5"
            value={total}
            onChange={(e) => setTotal(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div className="flex gap-3 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={submitting}
            className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-400"
          >
            {submitting ? 'Updating...' : 'Update Game'}
          </button>
        </div>
      </form>
    </Modal>
  );
}

// Create Week Modal
interface CreateWeekModalProps {
  isOpen: boolean;
  onClose: () => void;
  seasons: Season[];
  onSuccess: () => void;
}

function CreateWeekModal({ isOpen, onClose, seasons, onSuccess }: CreateWeekModalProps) {
  const [seasonId, setSeasonId] = useState('');
  const [weekNumber, setWeekNumber] = useState('');
  const [name, setName] = useState('');
  const [lockTime, setLockTime] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      // Convert datetime-local format to ISO 8601
      const isoLockTime = new Date(lockTime).toISOString();

      await adminService.createWeek(
        parseInt(seasonId),
        parseInt(weekNumber),
        name,
        isoLockTime
      );
      alert('Week created successfully!');
      onSuccess();
    } catch (error: any) {
      alert(error.response?.data || 'Failed to create week');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Create New Week">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Season
          </label>
          <select
            value={seasonId}
            onChange={(e) => setSeasonId(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          >
            <option value="">Select season</option>
            {seasons.map((season) => (
              <option key={season.id} value={season.id}>
                {season.name}
              </option>
            ))}
          </select>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Week Number
          </label>
          <input
            type="number"
            value={weekNumber}
            onChange={(e) => setWeekNumber(e.target.value)}
            placeholder="e.g., 1"
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., Week 1"
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Lock Time (when picks close)
          </label>
          <input
            type="datetime-local"
            value={lockTime}
            onChange={(e) => setLockTime(e.target.value)}
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div className="flex gap-3 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={submitting}
            className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-400"
          >
            {submitting ? 'Creating...' : 'Create Week'}
          </button>
        </div>
      </form>
    </Modal>
  );
}

// Create Season Modal
interface CreateSeasonModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

function CreateSeasonModal({ isOpen, onClose, onSuccess }: CreateSeasonModalProps) {
  const [year, setYear] = useState(new Date().getFullYear().toString());
  const [name, setName] = useState('');
  const [isActive, setIsActive] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await adminService.createSeason(parseInt(year), name, isActive);
      alert('Season created successfully!');
      onSuccess();
    } catch (error: any) {
      alert(error.response?.data || 'Failed to create season');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Create New Season">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Year
          </label>
          <input
            type="number"
            value={year}
            onChange={(e) => setYear(e.target.value)}
            placeholder="e.g., 2025"
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g., 2025 Season"
            className="w-full px-3 py-2 border rounded-md"
            required
          />
        </div>

        <div>
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={isActive}
              onChange={(e) => setIsActive(e.target.checked)}
              className="mr-2"
            />
            <span className="text-sm font-medium text-gray-700">Set as active season</span>
          </label>
        </div>

        <div className="flex gap-3 pt-4">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50 transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={submitting}
            className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors disabled:bg-gray-400"
          >
            {submitting ? 'Creating...' : 'Create Season'}
          </button>
        </div>
      </form>
    </Modal>
  );
}
