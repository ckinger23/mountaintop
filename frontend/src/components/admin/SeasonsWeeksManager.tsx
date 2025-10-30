import { useState } from 'react';
import { Season, Week, WeekStatus } from '../../types';
import { adminService } from '../../services/api';

interface SeasonsWeeksManagerProps {
  seasons: Season[];
  weeks: Week[];
  onWeekUpdate: () => void;
}

export default function SeasonsWeeksManager({ seasons, weeks, onWeekUpdate }: SeasonsWeeksManagerProps) {
  const [expandedSeasons, setExpandedSeasons] = useState<Set<number>>(new Set([seasons.find(s => s.is_active)?.id || seasons[0]?.id]));
  const [pickDeadlineInput, setPickDeadlineInput] = useState<{ [key: number]: string }>({});
  const [loading, setLoading] = useState<number | null>(null);

  const toggleSeason = (seasonId: number) => {
    const newExpanded = new Set(expandedSeasons);
    if (newExpanded.has(seasonId)) {
      newExpanded.delete(seasonId);
    } else {
      newExpanded.add(seasonId);
    }
    setExpandedSeasons(newExpanded);
  };

  const getStatusBadge = (status: WeekStatus) => {
    const badges = {
      creating: { bg: 'bg-gray-100', text: 'text-gray-800', label: 'Creating' },
      picking: { bg: 'bg-green-100', text: 'text-green-800', label: 'Picking' },
      scoring: { bg: 'bg-yellow-100', text: 'text-yellow-800', label: 'Scoring' },
      finished: { bg: 'bg-blue-100', text: 'text-blue-800', label: 'Finished' },
    };
    const badge = badges[status];
    return (
      <span className={`text-xs ${badge.bg} ${badge.text} px-2 py-1 rounded font-medium`}>
        {badge.label}
      </span>
    );
  };

  const handleOpenForPicks = async (weekId: number) => {
    const pickDeadline = pickDeadlineInput[weekId];
    if (!pickDeadline) {
      alert('Please enter a pick deadline');
      return;
    }

    try {
      setLoading(weekId);
      await adminService.openWeekForPicks(weekId, pickDeadline);
      alert('Week opened for picks!');
      setPickDeadlineInput(prev => ({ ...prev, [weekId]: '' }));
      onWeekUpdate();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to open week for picks');
    } finally {
      setLoading(null);
    }
  };

  const handleLockWeek = async (weekId: number) => {
    if (!confirm('Are you sure you want to lock this week? Users will no longer be able to submit picks.')) {
      return;
    }

    try {
      setLoading(weekId);
      await adminService.lockWeek(weekId);
      alert('Week locked!');
      onWeekUpdate();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to lock week');
    } finally {
      setLoading(null);
    }
  };

  const handleCompleteWeek = async (weekId: number) => {
    if (!confirm('Are you sure you want to complete this week? Make sure all game results have been entered.')) {
      return;
    }

    try {
      setLoading(weekId);
      await adminService.completeWeek(weekId);
      alert('Week completed!');
      onWeekUpdate();
    } catch (error: any) {
      alert(error.response?.data?.error || 'Failed to complete week');
    } finally {
      setLoading(null);
    }
  };

  const getSeasonWeeks = (seasonId: number) => {
    return weeks.filter(w => w.season_id === seasonId).sort((a, b) => a.week_number - b.week_number);
  };

  const getSeasonStatus = (seasonWeeks: Week[]) => {
    if (seasonWeeks.length === 0) return null;
    const hasActiveWeek = seasonWeeks.some(w => w.status === 'picking' || w.status === 'scoring');
    const allFinished = seasonWeeks.every(w => w.status === 'finished');

    if (hasActiveWeek) return 'active';
    if (allFinished && seasonWeeks.length > 0) return 'completed';
    return null;
  };

  return (
    <div className="space-y-4">
      {seasons.length === 0 ? (
        <div className="text-center text-gray-500 py-8">
          No seasons found. Create one to get started.
        </div>
      ) : (
        seasons
          .sort((a, b) => b.year - a.year)
          .map((season) => {
            const seasonWeeks = getSeasonWeeks(season.id);
            const seasonStatus = getSeasonStatus(seasonWeeks);
            const isExpanded = expandedSeasons.has(season.id);

            return (
              <div key={season.id} className="bg-white border rounded-lg overflow-hidden">
                {/* Season Header */}
                <div
                  className="p-4 cursor-pointer hover:bg-gray-50 transition-colors"
                  onClick={() => toggleSeason(season.id)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <svg
                        className={`w-5 h-5 text-gray-400 transition-transform ${
                          isExpanded ? 'transform rotate-90' : ''
                        }`}
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                      </svg>
                      <div>
                        <div className="font-semibold text-lg">{season.name}</div>
                        <div className="text-sm text-gray-500">
                          {seasonWeeks.length} week{seasonWeeks.length !== 1 ? 's' : ''}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {season.is_active && (
                        <span className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded font-medium">
                          Active
                        </span>
                      )}
                      {seasonStatus === 'completed' && (
                        <span className="text-xs bg-gray-100 text-gray-800 px-2 py-1 rounded font-medium">
                          Completed
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* Weeks List */}
                {isExpanded && (
                  <div className="border-t bg-gray-50">
                    {seasonWeeks.length === 0 ? (
                      <div className="p-4 text-center text-gray-500 text-sm">
                        No weeks created yet for this season.
                      </div>
                    ) : (
                      <div className="divide-y">
                        {seasonWeeks.map((week) => (
                          <div key={week.id} className="p-4 bg-white hover:bg-gray-50 transition-colors">
                            <div className="flex items-start justify-between">
                              <div className="flex-1">
                                <div className="flex items-center gap-2 mb-2">
                                  <span className="font-medium">{week.name}</span>
                                  {getStatusBadge(week.status)}
                                </div>
                                {week.pick_deadline && (
                                  <div className="text-sm text-gray-600">
                                    Pick Deadline: {new Date(week.pick_deadline).toLocaleString()}
                                  </div>
                                )}
                              </div>

                              {/* Action Buttons */}
                              <div className="ml-4">
                                {week.status === 'creating' && (
                                  <div className="flex items-center gap-2">
                                    <input
                                      type="datetime-local"
                                      value={pickDeadlineInput[week.id] || ''}
                                      onChange={(e) =>
                                        setPickDeadlineInput(prev => ({ ...prev, [week.id]: e.target.value }))
                                      }
                                      className="text-sm px-2 py-1 border rounded"
                                      onClick={(e) => e.stopPropagation()}
                                    />
                                    <button
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        handleOpenForPicks(week.id);
                                      }}
                                      disabled={loading === week.id}
                                      className="text-sm bg-blue-600 text-white px-3 py-1 rounded hover:bg-blue-700 disabled:opacity-50 whitespace-nowrap"
                                    >
                                      {loading === week.id ? 'Opening...' : 'Open for Picks'}
                                    </button>
                                  </div>
                                )}

                                {week.status === 'picking' && (
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      handleLockWeek(week.id);
                                    }}
                                    disabled={loading === week.id}
                                    className="text-sm bg-yellow-600 text-white px-3 py-1 rounded hover:bg-yellow-700 disabled:opacity-50"
                                  >
                                    {loading === week.id ? 'Locking...' : 'Lock Week'}
                                  </button>
                                )}

                                {week.status === 'scoring' && (
                                  <button
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      handleCompleteWeek(week.id);
                                    }}
                                    disabled={loading === week.id}
                                    className="text-sm bg-green-600 text-white px-3 py-1 rounded hover:bg-green-700 disabled:opacity-50"
                                  >
                                    {loading === week.id ? 'Completing...' : 'Complete Week'}
                                  </button>
                                )}

                                {week.status === 'finished' && (
                                  <span className="text-sm text-gray-500 italic">Complete</span>
                                )}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>
            );
          })
      )}
    </div>
  );
}
