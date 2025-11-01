import { useState } from 'react';
import toast from 'react-hot-toast';
import { adminService } from '../../../services/api';
import { useLeague } from '../../../hooks/useLeague';
import Modal from '../../Modal';

interface CreateSeasonModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess: () => void;
}

export default function CreateSeasonModal({
  isOpen,
  onClose,
  onSuccess
}: CreateSeasonModalProps) {
  const { currentLeague } = useLeague();
  const [year, setYear] = useState(new Date().getFullYear().toString());
  const [name, setName] = useState('');
  const [isActive, setIsActive] = useState(true);
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!currentLeague) {
      toast.error('Please select a league first');
      return;
    }

    setSubmitting(true);
    try {
      await adminService.createSeason(currentLeague.id, parseInt(year), name, isActive);
      toast.success('Season created successfully!');
      onSuccess();
    } catch (error: any) {
      toast.error(error.response?.data || 'Failed to create season');
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
