"use client";

import { Crown, Users } from "lucide-react";
import { User } from "@/lib/game/types";
import { useGameContext } from "@/hooks/gamecontext";
import { cn } from "@/lib/utils";

// TODO: Temporary
interface MockPlayer {
  userId: string;
  username: string;
  isHost: boolean;
  color: string;
}

// TODO: Temporary
const MOCK_PLAYERS: MockPlayer[] = [
  { userId: "1", username: "Alice", isHost: true, color: "#8b5cf6" },
  { userId: "2", username: "Bob", isHost: false, color: "#ec4899" },
  { userId: "3", username: "Charlie", isHost: false, color: "#3b82f6" },
  { userId: "4", username: "David", isHost: false, color: "#10b981" },
  { userId: "5", username: "Emma", isHost: false, color: "#f59e0b" },
  { userId: "6", username: "Fanny", isHost: false, color: "#ef4444" },
  { userId: "7", username: "George", isHost: false, color: "#06b6d4" },
  { userId: "8", username: "Hannes", isHost: false, color: "#a855f7" },
  { userId: "9", username: "Ingrid", isHost: false, color: "#8b5cf6" },
  { userId: "10", username: "John", isHost: false, color: "#f59e0b" },
  //{ userId: "11", username: "Klara", isHost: false },
  //{ userId: "12", username: "Leo", isHost: false },
];

interface PlayerListProps {
  className?: string;
}

function PlayerAvatar({ name, color }: { name: string; color: string }) {
  //const initials = name.split(" ").map(n => n[0]).join("").toUpperCase();
  const split = name.split(" ");
  return (
    <div
      className="w-10 h-10 rounded-full flex items-center justify-center font-body font-bold text-white border-2 border-card shrink-0"
      style={{ backgroundColor: color, boxShadow: `0 3px 0 0 ${color}88` }}>
      {/*style={{ backgroundColor: p.color, boxShadow: `0 3px 0 0 ${p.color}88` }} */}
      {split.length >= 2 ? `${split[0][0]}${split[1][0]}` : name[0]}
    </div>
  );
}

// TODO: function PlayerCard({ username, isHost, userId }: User & { isHost: boolean })
function PlayerCard({ player }: { player: MockPlayer }) {
  //const { user } = useGameContext();

  //if (!user) return null;

  return (
    <div className="flex items-center px-4 py-3 gap-4 transition-all rounded-lg bg-muted/50 border-2 border-border hover:border-primary font-display">
      <PlayerAvatar name={player.username} color={player.color} /> {/* username */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center">
          {/* 
            <span
              className={`text-base truncate text-foreground ${user.userId === userId ? "font-extrabold" : "font-bold"}`}>
              {user.userId === userId ? "Du" : username}
            </span>
          */}
          {/* TODO: Connect userId with "Du" */}
          <span className="text-base truncate text-foreground font-bold">
            {player.username}
            {player.isHost && <span className="text-xs text-muted-foreground ml-2"> (Du)</span>}
          </span>
        </div>
      </div>
      {/*isHost */}
      {player.isHost && (
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
      <div className="w-10 h-10 rounded-full border border-dashed border-border/50 shrink-0 bg-muted/10 flex items-center justify-center text-muted-foreground/30">
        ?
      </div>
      <div className="flex-1">
        <span className="text-sm italic text-muted-foreground/40">Väntar på spelare...</span>
      </div>
    </div>
  );
}

export function PlayerList({ className }: PlayerListProps) {
  /*
    const { gamestate, user } = useGameContext();

    if (!gamestate || !user) return;

    const teamAPlayers: User[] = [];
    const teamBPlayers: User[] = [];

    Object.keys(gamestate.players).map((key) => {
        if (gamestate.players[key].team === "a") {
        teamAPlayers.push(gamestate.players[key]);
        } else {
        teamBPlayers.push(gamestate.players[key]);
        }
    });
  */

  const MAX_SLOTS = 12;
  const emptySlotsCount = Math.max(0, MAX_SLOTS - MOCK_PLAYERS.length);

  return (
    <div className={cn("game-card flex-1 p-6 border shadow-sm rounded-2xl bg-card border-border", className)}>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <h2 className="font-display text-sm font-bold text-muted-foreground uppercase tracking-wider flex items-center gap-2">
            <Users className="w-4 h-4" />
            Players
          </h2>
          <span className="font-display font-bold text-sm px-2.5 py-0.5 rounded-full bg-muted text-foreground tabular-nums">
            {MOCK_PLAYERS.length}/12
          </span>
          {/* teamAPlayers */}
        </div>
        {/* teamAPlayers */}
        <div className="flex flex-col gap-2 max-h-100 overflow-y-auto custom-scrollbar">
          {MOCK_PLAYERS.map((player) => (
            <PlayerCard key={player.userId} player={player} /*{...player} isHost={player.userId === gamestate.host}*/ />
          ))}
          {Array.from({ length: emptySlotsCount }).map((_, index) => (
            <EmptySlot key={`empty-${index}`} />
          ))}
        </div>
      </div>
    </div>
  );
}
