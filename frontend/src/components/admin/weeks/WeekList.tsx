import { Week } from '../../../types';

interface WeekListProps {
  weeks: Week[];
  currentWeek: Week | null;
}

export default function WeekList({ weeks, currentWeek }: WeekListProps) {
  if (weeks.length === 0) {
    return (
      <div className="text-center text-gray-500 py-8">
        No weeks found. Create one to get started.
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {weeks.map((week) => (
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
      ))}
    </div>
  );
}
