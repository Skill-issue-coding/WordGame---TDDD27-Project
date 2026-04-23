import { cn } from "@/lib/utils";

interface GameModeCardProps {
  title: string;
  description: string;
  icon: string;
  players: string;
  color: "green" | "pink" | "blue" | "yellow";
  onClick: () => void;
}

const colorMap = {
  green: "border-neon-green/30 hover:border-neon-green/60 hover:glow-border-green",
  pink: "border-neon-pink/30 hover:border-neon-pink/60 hover:glow-border-pink",
  blue: "border-neon-blue/30 hover:border-neon-blue/60 hover:glow-border-blue",
  yellow: "border-neon-yellow/30 hover:border-neon-yellow/60",
};

const textColorMap = {
  green: "text-neon-green",
  pink: "text-neon-pink",
  blue: "text-neon-blue",
  yellow: "text-neon-yellow",
};

const GameModeCard = ({ title, description, icon, players, color, onClick }: GameModeCardProps) => {
  return (
    <button onClick={onClick} className={cn("game-card border-2 text-left transition-all duration-300 hover:scale-[1.02] cursor-pointer group w-full", colorMap[color])}>
      <div className="flex items-start gap-4">
        <span className="text-4xl">{icon}</span>
        <div className="flex-1 min-w-0">
          <h3 className={cn("font-display text-xl font-bold mb-1", textColorMap[color])}>{title}</h3>
          <p className="text-muted-foreground text-sm leading-relaxed mb-3">{description}</p>
          <span className="text-xs text-muted-foreground bg-muted/50 px-2 py-1 rounded-full">{players}</span>
        </div>
      </div>
    </button>
  );
};

export default GameModeCard;
