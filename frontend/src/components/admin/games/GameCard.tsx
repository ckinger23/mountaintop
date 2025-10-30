import { Game } from '../../../types';

interface GameCardProps {
  game: Game;
  onEdit: () => void;
  onDelete: () => void;
}

export default function GameCard({ game, onEdit, onDelete }: GameCardProps) {
  return (
    <div className="bg-white border rounded-lg p-4">
      <div className="flex items-center justify-between mb-3">
        <div className="text-sm text-gray-500">
          {new Date(game.game_time).toLocaleString()}
        </div>
        <div className="flex gap-2">
          <button
            onClick={onEdit}
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            Edit
          </button>
          <button
            onClick={onDelete}
            className="text-red-600 hover:text-red-800 text-sm font-medium"
          >
            Delete
          </button>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 items-center">
        <div className="text-right">
          <div className="font-semibold">{game.away_team.name}</div>
          <div className="text-sm text-gray-500">{game.away_team.abbreviation}</div>
        </div>

        <div className="text-center text-gray-500 font-bold">@</div>

        <div className="text-left">
          <div className="font-semibold">{game.home_team.name}</div>
          <div className="text-sm text-gray-500">{game.home_team.abbreviation}</div>
        </div>
      </div>

      <div className="mt-3 text-xs text-gray-500 text-center">
        Spread: {game.home_team.abbreviation} {game.home_spread > 0 ? '+' : ''}{game.home_spread} | Total: {game.total}
      </div>
    </div>
  );
}
