import { Season } from '../../../types';

interface SeasonListProps {
  seasons: Season[];
}

export default function SeasonList({ seasons }: SeasonListProps) {
  return (
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
  );
}
