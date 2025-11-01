import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { League } from '../types';
import { leaguesService } from '../services/api';

interface LeagueContextType {
  currentLeague: League | null;
  leagues: League[];
  loading: boolean;
  setCurrentLeague: (league: League | null) => void;
  refreshLeagues: () => Promise<void>;
}

const LeagueContext = createContext<LeagueContextType | undefined>(undefined);

export const LeagueProvider = ({ children }: { children: ReactNode }) => {
  const [currentLeague, setCurrentLeagueState] = useState<League | null>(null);
  const [leagues, setLeagues] = useState<League[]>([]);
  const [loading, setLoading] = useState(true);

  // Load leagues on mount
  useEffect(() => {
    loadLeagues();
  }, []);

  // Load current league from localStorage on mount
  useEffect(() => {
    const savedLeagueId = localStorage.getItem('currentLeagueId');
    if (savedLeagueId && leagues.length > 0) {
      const league = leagues.find(l => l.id === parseInt(savedLeagueId));
      if (league) {
        setCurrentLeagueState(league);
      } else if (leagues.length > 0) {
        // If saved league not found, default to first league
        setCurrentLeagueState(leagues[0]);
        localStorage.setItem('currentLeagueId', leagues[0].id.toString());
      }
    } else if (leagues.length > 0 && !currentLeague) {
      // No saved league, default to first one
      setCurrentLeagueState(leagues[0]);
      localStorage.setItem('currentLeagueId', leagues[0].id.toString());
    }
  }, [leagues]);

  const loadLeagues = async () => {
    try {
      setLoading(true);
      const fetchedLeagues = await leaguesService.getMyLeagues();
      setLeagues(fetchedLeagues);
    } catch (error) {
      console.error('Failed to load leagues:', error);
      setLeagues([]);
    } finally {
      setLoading(false);
    }
  };

  const setCurrentLeague = (league: League | null) => {
    setCurrentLeagueState(league);
    if (league) {
      localStorage.setItem('currentLeagueId', league.id.toString());
    } else {
      localStorage.removeItem('currentLeagueId');
    }
  };

  const refreshLeagues = async () => {
    await loadLeagues();
  };

  return (
    <LeagueContext.Provider
      value={{
        currentLeague,
        leagues,
        loading,
        setCurrentLeague,
        refreshLeagues,
      }}
    >
      {children}
    </LeagueContext.Provider>
  );
};

export const useLeague = () => {
  const context = useContext(LeagueContext);
  if (context === undefined) {
    throw new Error('useLeague must be used within a LeagueProvider');
  }
  return context;
};
