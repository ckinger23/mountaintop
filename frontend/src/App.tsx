import { useState, useEffect } from 'react';
import type { Week } from './types';
import GameList from './components/GameList';
import Leaderboard from './components/Leaderboard';
import axios from 'axios';
import './App.css'

function App() {
  const [currentWeek, setCurrentWeek] = useState<Week | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchCurrentWeek = async () => {
      try {
        const response = await axios.get('/api/weeks/current');
        setCurrentWeek(response.data);
        setLoading(false);
      } catch (err) {
        setError('Failed to load current week');
        setLoading(false);
      }
    };

    fetchCurrentWeek();
  }, []);

  if (loading) return (
    <div className="placeholder">
      <div className="placeholder-content">
        <div className="placeholder-icon">⏳</div>
        <p>Loading current week data...</p>
      </div>
    </div>
  );

  if (error) return (
    <div className="placeholder error">
      <div className="placeholder-content">
        <div className="placeholder-icon">❌</div>
        <p>Failed to load week data</p>
        <p className="placeholder-subtext">Please try refreshing the page</p>
      </div>
    </div>
  );

  if (!currentWeek) return (
    <div className="placeholder">
      <div className="placeholder-content">
        <div className="placeholder-icon">📅</div>
        <p>No current week data available</p>
        <p className="placeholder-subtext">Check back when the next week's games are scheduled</p>
      </div>
    </div>
  );

  return (
    <div className="app-container">
      <header className="app-header">
        <h1>Football Picking League</h1>
      </header>
      <main className="app-main">
        <div className="leaderboard-section">
          <Leaderboard />
        </div>
        <div className="games-section">
          <GameList week={currentWeek?.number || 0} />
        </div>
      </main>
    </div>
  );
}

export default App;
