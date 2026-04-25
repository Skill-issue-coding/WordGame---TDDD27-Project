"use client";

import { Crown } from "lucide-react";
import { User } from "@/lib/game/types";
import { useGameContext } from "@/hooks/gamecontext";
import { cn } from "@/lib/utils";

// TODO: Temporary
interface MockPlayer {
  userId: string;
  username: string;
  isHost: boolean;
}

// TODO: Temporary
const MOCK_PLAYERS: MockPlayer[] = [
  { userId: "1", username: "Alice", isHost: true },
  { userId: "2", username: "Bob", isHost: false },
  { userId: "3", username: "Charlie", isHost: false },
  { userId: "4", username: "David", isHost: false },
  { userId: "5", username: "Emma", isHost: false },
  { userId: "6", username: "Fanny", isHost: false },
  { userId: "7", username: "George", isHost: false },
  { userId: "8", username: "Hannes", isHost: false },
  { userId: "9", username: "Ingrid", isHost: false },
  { userId: "10", username: "John", isHost: false },
  //{ userId: "11", username: "Klara", isHost: false },
  //{ userId: "12", username: "Leo", isHost: false },
];

interface PlayerListProps {
  className?: string;
}

function PlayerAvatar({ name }: { name: string }) {
  //const initials = name.split(" ").map(n => n[0]).join("").toUpperCase();
  const split = name.split(" ");
  return (
    <div className="flex items-center justify-center w-10 h-10 text-lg font-bold rounded-full shrink-0 bg-primary/30 text-tile-foreground">
      {split.length >= 2 ? `${split[0][0]}${split[1][0]}` : name[0]}
    </div>
  );
}

// TODO: function PlayerCard({ username, isHost, userId }: User & { isHost: boolean })
function PlayerCard({ player }: { player: MockPlayer }) {
  //const { user } = useGameContext();

  //if (!user) return null;

  return (
    <div className="flex items-center px-4 py-3 border gap-4 transition-all rounded-xl bg-card border-border hover:border-primary/30">
      <PlayerAvatar name={player.username} /> {/* username */}
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
        <div className="flex items-center p-2 rounded-md bg-neon-yellow/15">
          <Crown className="size-6 text-neon-yellow shrink-0 stroke-3 " />
        </div>
      )}
    </div>
  );
}

function EmptySlot() {
  return (
    <div className="flex items-center px-4 py-3 border border-dashed gap-4 rounded-xl border-border/90 opacity-90 bg-muted/5">
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
    <div className={cn("flex-1 p-6 border shadow-sm rounded-2xl bg-card border-border", className)}>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <p className="text-base font-bold text-foreground">Spelare</p>
          <span className="text-xs text-muted-foreground">({MOCK_PLAYERS.length}/12 anslutna)</span>{" "}
          {/* teamAPlayers */}
        </div>
        {/* teamAPlayers */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
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
