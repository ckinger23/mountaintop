import { useState } from 'react';
import { adminService } from '../../../services/api';
import { Game, Team, Week } from '../../../types';
import Modal from '../../Modal';

interface EditGameModalProps {
  isOpen: boolean;
  onClose: () => void;
  game: Game;
  teams: Team[];
  currentWeek: Week | null;
  onSuccess: () => void;
}

export default function EditGameModal({
  isOpen,
  onClose,
  game,
  teams,
  currentWeek,
  onSuccess
}: EditGameModalProps) {
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
