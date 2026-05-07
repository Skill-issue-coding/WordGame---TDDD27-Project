"use client";

import { Crown, Users } from "lucide-react";
import { User } from "@/lib/game/types";
import { useGameContext } from "@/hooks/gamecontext";
import { cn } from "@/lib/utils";

interface PlayerListProps {
  className?: string;
}

function PlayerAvatar({ name, color }: { name: string; color: string }) {
  //const initials = name.split(" ").map(n => n[0]).join("").toUpperCase();
  const split = name.split(" ");
  return (
    <div className="w-10 h-10 rounded-full flex items-center justify-center font-body font-bold text-white border-2 border-card shrink-0" style={{ backgroundColor: color, boxShadow: `0 3px 0 0 ${color}88` }}>
      {/*style={{ backgroundColor: p.color, boxShadow: `0 3px 0 0 ${p.color}88` }} */}
      {split.length >= 2 ? `${split[0][0]}${split[1][0]}` : name[0]}
    </div>
  );
}

// TODO: function PlayerCard({ username, isHost, userId }: User & { isHost: boolean })
function PlayerCard({ player, isHost }: { player: User; isHost: boolean }) {
  //const { user } = useGameContext();

  //if (!user) return null;

  return (
    <div className="flex items-center px-4 py-3 gap-4 transition-all rounded-lg bg-muted/50 border-2 border-border hover:border-primary font-display">
      <PlayerAvatar name={player.username} color={player.background} />
      <div className="flex-1 min-w-0">
        <div className="flex items-center">
          <span className="text-base truncate text-foreground font-bold">
            {player.username}
            {isHost && <span className="text-xs text-muted-foreground ml-2"> (Du)</span>}
          </span>
        </div>
      </div>
      {isHost && (
        <div className="flex items-center p-2 rounded-md bg-game-yellow/20">
          <Crown className="size-6 text-game-yellow shrink-0 stroke-3 " />
        </div>
      )}
    </div>
  );
}

function EmptySlot() {
  return (
    <div className="flex items-center px-4 py-3 border border-dashed gap-4 rounded-lg border-border/90 opacity-90 bg-muted/5">
      <div className="w-10 h-10 rounded-full border border-dashed border-border/50 shrink-0 bg-muted/10 flex items-center justify-center text-muted-foreground/30">?</div>
      <div className="flex-1">
        <span className="text-sm italic text-muted-foreground/40">Väntar på spelare...</span>
      </div>
    </div>
  );
}

export function PlayerList({ className }: PlayerListProps) {
  const { lobbyState } = useGameContext();
  const MAX_SLOTS = 12;
  const playerCount = Object.values(lobbyState?.users || {}).length;
  const emptySlotsCount = Math.max(0, MAX_SLOTS - playerCount);

  return (
    <div className={cn("game-card flex-1 p-6 border shadow-sm rounded-2xl bg-card border-border", className)}>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <h2 className="font-display text-sm font-bold text-muted-foreground uppercase tracking-wider flex items-center gap-2">
            <Users className="w-4 h-4" />
            Spelare
          </h2>
          <span className="font-display font-bold text-sm px-2.5 py-0.5 rounded-full bg-muted text-foreground tabular-nums">{playerCount}/12</span>
        </div>
        <div className="flex flex-col gap-2 max-h-105 overflow-y-auto custom-scrollbar">
          {lobbyState?.users && Object.values(lobbyState.users).map((player) => <PlayerCard key={player.user_id} player={player} isHost={player.user_id === lobbyState?.host} />)}
          {Array.from({ length: emptySlotsCount }).map((_, index) => (
            <EmptySlot key={`empty-${index}`} />
          ))}
        </div>
      </div>
    </div>
  );
}
