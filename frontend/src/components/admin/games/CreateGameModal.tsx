import { useState } from 'react';
import { adminService } from '../../../services/api';
import { Team, Week } from '../../../types';
import Modal from '../../Modal';

interface CreateGameModalProps {
  isOpen: boolean;
  onClose: () => void;
  teams: Team[];
  currentWeek: Week | null;
  onSuccess: () => void;
}

export default function CreateGameModal({
  isOpen,
  onClose,
  teams,
  currentWeek,
  onSuccess
}: CreateGameModalProps) {
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
