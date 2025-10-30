import { useState } from 'react';
import toast from 'react-hot-toast';
import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';
import '../../../styles/datepicker.css';
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
  const [gameTime, setGameTime] = useState<Date>(new Date(game.game_time));
  const [homeSpread, setHomeSpread] = useState(game.home_spread.toString());
  const [total, setTotal] = useState(game.total.toString());
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!currentWeek) {
      toast.error('No current week found.');
      return;
    }

    setSubmitting(true);
    try {
      // Convert Date object to ISO 8601
      const isoGameTime = gameTime.toISOString();

      await adminService.updateGame(
        game.id,
        currentWeek.id,
        parseInt(homeTeamId),
        parseInt(awayTeamId),
        isoGameTime,
        parseFloat(homeSpread),
        parseFloat(total)
      );
      toast.success('Game updated successfully!');
      onSuccess();
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to update game');
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
            <span className="ml-2 text-xs font-normal text-blue-600">
              ({Intl.DateTimeFormat().resolvedOptions().timeZone})
            </span>
          </label>
          <DatePicker
            selected={gameTime}
            onChange={(date: Date | null) => date && setGameTime(date)}
            showTimeSelect
            timeFormat="HH:mm"
            timeIntervals={15}
            dateFormat="MMMM d, yyyy h:mm aa"
            minDate={new Date()}
            className="w-full px-3 py-2 border rounded-md"
            placeholderText="Select date and time"
            wrapperClassName="w-full"
            portalId="root-portal"
          />
          <p className="text-xs text-gray-500 mt-1">
            Your local timezone
          </p>
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
