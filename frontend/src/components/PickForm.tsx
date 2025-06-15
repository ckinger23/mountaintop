import { useState } from 'react';
import type { Game } from '../types';
import axios from 'axios';

interface PickFormProps {
  game: Game;
  onPickSubmitted: () => void;
}

const PickForm: React.FC<PickFormProps> = ({ game, onPickSubmitted }) => {
  const [selectedTeam, setSelectedTeam] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedTeam) return;

    setLoading(true);
    setError(null);

    try {
      await axios.post('/api/picks', {
        gameId: game.id,
        selectedTeam,
      });
      onPickSubmitted();
    } catch (err) {
      setError('Failed to submit pick');
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="pick-form">
      <h3>Make Your Pick</h3>
      <div className="team-options">
        <label>
          <input
            type="radio"
            value={game.homeTeam}
            checked={selectedTeam === game.homeTeam}
            onChange={(e) => setSelectedTeam(e.target.value)}
          />
          {game.homeTeam}
        </label>
        <label>
          <input
            type="radio"
            value={game.awayTeam}
            checked={selectedTeam === game.awayTeam}
            onChange={(e) => setSelectedTeam(e.target.value)}
          />
          {game.awayTeam}
        </label>
      </div>
      <button type="submit" disabled={loading || !selectedTeam}>
        {loading ? 'Submitting...' : 'Submit Pick'}
      </button>
      {error && <div className="error">{error}</div>}
    </form>
  );
};

export default PickForm;
