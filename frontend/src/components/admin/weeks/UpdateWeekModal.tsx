import { useState } from 'react';
import toast from 'react-hot-toast';
import DatePicker from 'react-datepicker';
import 'react-datepicker/dist/react-datepicker.css';
import '../../../styles/datepicker.css';
import { adminService } from '../../../services/api';
import { Week } from '../../../types';
import Modal from '../../Modal';

interface UpdateWeekModalProps {
  isOpen: boolean;
  onClose: () => void;
  week: Week;
  onSuccess: () => void;
}

export default function UpdateWeekModal({
  isOpen,
  onClose,
  week,
  onSuccess
}: UpdateWeekModalProps) {
  const [pickDeadline, setPickDeadline] = useState<Date | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const handleOpenForPicks = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!pickDeadline) {
      toast.error('Please set a pick deadline');
      return;
    }

    setSubmitting(true);
    try {
      // Convert Date object to ISO 8601/RFC3339
      const isoPickDeadline = pickDeadline.toISOString();
      await adminService.openWeekForPicks(week.id, isoPickDeadline);
      toast.success('Week opened for picks!');
      onSuccess();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to open week for picks');
    } finally {
      setSubmitting(false);
    }
  };

  const handleLockWeek = async () => {
    if (!confirm('Lock this week? Users will no longer be able to submit picks.')) {
      return;
    }

    setSubmitting(true);
    try {
      await adminService.lockWeek(week.id);
      toast.success('Week locked! Enter game results to score picks.');
      onSuccess();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to lock week');
    } finally {
      setSubmitting(false);
    }
  };

  const handleCompleteWeek = async () => {
    if (!confirm('Complete this week? This will finalize all scores.')) {
      return;
    }

    setSubmitting(true);
    try {
      await adminService.completeWeek(week.id);
      toast.success('Week completed!');
      onSuccess();
    } catch (error: any) {
      toast.error(error.response?.data?.error || 'Failed to complete week');
    } finally {
      setSubmitting(false);
    }
  };

  const renderContent = () => {
    switch (week.status) {
      case 'creating':
        return (
          <form onSubmit={handleOpenForPicks} className="space-y-4">
            <p className="text-sm text-gray-600">
              This week is in "Creating" status. Once you've added all games, open it for picks.
            </p>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Pick Deadline
                <span className="ml-2 text-xs font-normal text-blue-600">
                  ({Intl.DateTimeFormat().resolvedOptions().timeZone})
                </span>
              </label>
              <DatePicker
                selected={pickDeadline}
                onChange={(date: Date | null) => setPickDeadline(date)}
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
                Times are in your local timezone and will be converted to UTC for storage
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
                className="flex-1 bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors disabled:bg-gray-400"
              >
                {submitting ? 'Opening...' : 'Open for Picks'}
              </button>
            </div>
          </form>
        );

      case 'picking':
        return (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              This week is open for picks. Users can submit picks until the deadline.
            </p>

            {week.pick_deadline && (
              <div className="p-3 bg-blue-50 rounded-lg">
                <p className="text-sm font-medium text-blue-900">Pick Deadline</p>
                <p className="text-sm text-blue-700">
                  {new Date(week.pick_deadline).toLocaleString()}
                </p>
              </div>
            )}

            <div className="flex gap-3 pt-4">
              <button
                type="button"
                onClick={onClose}
                className="flex-1 px-4 py-2 border rounded-lg hover:bg-gray-50 transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleLockWeek}
                disabled={submitting}
                className="flex-1 bg-orange-600 text-white px-4 py-2 rounded-lg hover:bg-orange-700 transition-colors disabled:bg-gray-400"
              >
                {submitting ? 'Locking...' : 'Lock Week'}
              </button>
            </div>
          </div>
        );

      case 'scoring':
        return (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              This week is locked. Enter game results to score picks, then complete the week.
            </p>

            <div className="p-3 bg-purple-50 rounded-lg">
              <p className="text-sm font-medium text-purple-900">Status</p>
              <p className="text-sm text-purple-700">
                Enter all game results, then complete the week to finalize scores.
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
                onClick={handleCompleteWeek}
                disabled={submitting}
                className="flex-1 bg-green-600 text-white px-4 py-2 rounded-lg hover:bg-green-700 transition-colors disabled:bg-gray-400"
              >
                {submitting ? 'Completing...' : 'Complete Week'}
              </button>
            </div>
          </div>
        );

      case 'finished':
        return (
          <div className="space-y-4">
            <p className="text-sm text-gray-600">
              This week is finished. All scores have been finalized.
            </p>

            <div className="p-3 bg-gray-50 rounded-lg">
              <p className="text-sm font-medium text-gray-900">Status</p>
              <p className="text-sm text-gray-700">
                Week completed - no further actions available.
              </p>
            </div>

            <div className="pt-4">
              <button
                type="button"
                onClick={onClose}
                className="w-full px-4 py-2 border rounded-lg hover:bg-gray-50 transition-colors"
              >
                Close
              </button>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={`Update ${week.name}`}>
      {renderContent()}
    </Modal>
  );
}
