import { useState, useEffect } from 'react';
import type { User } from '../types';
import axios from 'axios';

const Leaderboard: React.FC = () => {
  const [leaderboardData, setLeaderboardData] = useState<{
    users: User[];
    totalUsers: number;
  } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchLeaderboard = async () => {
      try {
        const response = await axios.get('/api/leaderboard');
        // Ensure we're working with an array
        const users = Array.isArray(response.data) ? response.data : [];
        setLeaderboardData({
          users: users.map((user: any) => ({
            id: user.id,
            username: user.username || 'Unknown',
            totalPoints: user.totalPoints || 0,
            weeklyPoints: user.weeklyPoints || 0,
            rank: user.rank || users.indexOf(user) + 1
          })),
          totalUsers: users.length
        });
        setLoading(false);
      } catch (err) {
        console.error('Error fetching leaderboard:', err);
        setError('Failed to load leaderboard');
        setLoading(false);
      }
    };

    fetchLeaderboard();
  }, []);

  if (loading) return (
    <div className="placeholder">
      <div className="placeholder-content">
        <div className="placeholder-icon">⏳</div>
        <p>Loading leaderboard...</p>
      </div>
    </div>
  );

  if (error) return (
    <div className="placeholder error">
      <div className="placeholder-content">
        <div className="placeholder-icon">❌</div>
        <p>Failed to load leaderboard</p>
        <p className="placeholder-subtext">Please try refreshing the page</p>
      </div>
    </div>
  );

  if (!leaderboardData || leaderboardData.users.length === 0) return (
    <div className="placeholder">
      <div className="placeholder-content">
        <div className="placeholder-icon">📊</div>
        <p>No leaderboard data available</p>
        <p className="placeholder-subtext">Check back when the first week's games are played</p>
      </div>
    </div>
  );

  return (
    <div className="leaderboard">
      <h2>Leaderboard</h2>
      <p className="total-users">Total Users: {leaderboardData.totalUsers}</p>
      <table>
        <thead>
          <tr>
            <th>Rank</th>
            <th>Username</th>
            <th>Weekly Points</th>
            <th>Total Points</th>
          </tr>
        </thead>
        <tbody>
          {leaderboardData.users.map((user) => (
            <tr key={user.id}>
              <td>{user.rank}</td>
              <td>{user.username}</td>
              <td>{user.weeklyPoints}</td>
              <td>{user.totalPoints}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Leaderboard;
