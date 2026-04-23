import { cn } from "@/lib/utils";
import { Info, User } from "lucide-react";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";

export type Color = "green" | "pink" | "blue" | "yellow";

interface GameModeCardProps {
  title: string;
  description: string;
  icon: string;
  players: string;
  color: Color;
  // onClick: () => void;
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

const GameModeCard = ({ title, description, icon, players, color }: GameModeCardProps) => {
  return (
    <button
      // onClick={onClick}
      className={cn(
        "p-4 bg-game-surface border-border rounded-xl border-2 text-left transition-all duration-300 hover:scale-[1.02] cursor-pointer group w-full flex items-center",
        colorMap[color],
      )}>
      <div className="flex items-start w-full gap-4">
        <div className="flex items-center justify-between flex-1 gap-4">
          <div className="flex flex-1 gap-2">
            <span>{icon}</span>
            <h3 className={cn("font-display text-xl font-bold mb-1", textColorMap[color])}>{title}</h3>
          </div>
          <Tooltip>
            <TooltipTrigger asChild>
              <Info />
            </TooltipTrigger>
            <TooltipContent>
              <div className="flex items-center gap-1">
                <p className="flex flex-1 text-sm">{description}</p>
                <div className="flex items-center gap-1 px-4 py-1 text-xs rounded-full bg-muted/50">
                  <p className="flex text-nowrap">{players}</p>
                  <User className="size-6" />
                </div>
              </div>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </button>
  );
};

export default GameModeCard;
