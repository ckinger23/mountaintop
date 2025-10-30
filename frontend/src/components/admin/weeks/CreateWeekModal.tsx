import { useState } from 'react';
import toast from 'react-hot-toast';
import { adminService } from '../../../services/api';
import { Season } from '../../../types';
import Modal from '../../Modal';

interface CreateWeekModalProps {
  isOpen: boolean;
  onClose: () => void;
  seasons: Season[];
  defaultSeasonId?: number;
  onSuccess: () => void;
}

export default function CreateWeekModal({
  isOpen,
  onClose,
  seasons,
  defaultSeasonId,
  onSuccess
}: CreateWeekModalProps) {
  const [seasonId, setSeasonId] = useState(defaultSeasonId?.toString() || '');
  const [weekNumber, setWeekNumber] = useState('');
  const [name, setName] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      await adminService.createWeek(
        parseInt(seasonId),
        parseInt(weekNumber),
        name
      );
      toast.success('Week created successfully! Week is in "Creating" status.');
      onSuccess();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to create week');
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
          <p className="text-xs text-gray-500 mt-1">
            Week will be created in "Creating" status. Add games first, then open it for picks.
          </p>
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
