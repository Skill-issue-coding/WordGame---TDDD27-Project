"use client";

import { PlayerAvatar } from "@/components/lobby/PlayerList";
import { Button } from "@/components/ui/button";
import { House } from "lucide-react";

// Mock Data
const MOCK_LEADERBOARD = [
  { id: "1", name: "Du", color: "#8b5cf6", totalPoints: 232, rank: 1 },
  { id: "4", name: "Saga", color: "#10b981", totalPoints: 184, rank: 2 },
  { id: "3", name: "Astrid", color: "#fbd38d", totalPoints: 141, rank: 3 },
  { id: "5", name: "Nils", color: "#f6ad55", totalPoints: 114, rank: 4 },
  { id: "2", name: "Oskar", color: "#4fd1c5", totalPoints: 65, rank: 5 },
];

export function FinalScorePhase() {
  const [first, second, third, ...rest] = MOCK_LEADERBOARD;

  return (
    <div className="flex flex-col items-center w-full max-w-2xl mx-auto mt-4 relative z-10 py-8">
      <div className="text-center mb-12">
        <p className="font-display text-xs font-bold text-muted-foreground uppercase tracking-widest mb-2">
          Anti-Match · Slutresultat
        </p>
        <h1 className="font-display text-5xl font-extrabold text-primary tracking-tight">Du vann!</h1>
      </div>

      <div className="flex items-end justify-center gap-2 w-full mb-8 h-64">
        <div className="flex flex-col items-center w-1/3">
          <div className="flex flex-col items-center mb-3">
            <PlayerAvatar
              name={second.name}
              color={second.color}
              className="w-12 h-12 border-4 font-display font-bold"
            />
            <div className="font-display font-bold mt-1">{second.name}</div>
            <div className="font-body text-xs font-semibold text-muted-foreground">{second.totalPoints} p</div>
          </div>
          <div className="w-full h-32 bg-game-blue game-card rounded-xl flex justify-center items-center pt-4">
            <span className="font-display font-extrabold text-4xl text-white">2</span>
          </div>
        </div>

        <div className="flex flex-col items-center w-1/3">
          <div className="flex flex-col items-center mb-3">
            <PlayerAvatar name={first.name} color={first.color} className="w-14 h-14 border-4 font-display font-bold" />
            <div className="font-display font-extrabold mt-1">{first.name}</div>
            <div className="font-body text-xs font-semibold text-muted-foreground">{first.totalPoints} p</div>
          </div>
          <div className="w-full h-44 bg-game-yellow game-card rounded-xl flex justify-center items-center pt-4 z-10 -mx-2">
            <span className="font-display font-extrabold text-5xl text-white drop-shadow-sm">1</span>
          </div>
        </div>

        <div className="flex flex-col items-center w-1/3">
          <div className="flex flex-col items-center mb-3">
            <PlayerAvatar name={third.name} color={third.color} className="w-12 h-12 border-4 font-display font-bold" />
            <div className="font-display font-bold mt-1">{third.name}</div>
            <div className="font-body text-xs font-semibold text-muted-foreground">{third.totalPoints} p</div>
          </div>
          <div className="w-full h-24 bg-game-orange game-card rounded-xl flex justify-center items-center pt-3">
            <span className="font-display font-extrabold text-3xl text-white">3</span>
          </div>
        </div>
      </div>

      <div className="w-full flex flex-col gap-2 mb-8 bg-card game-card rounded-2xl p-2 border-2 border-border shadow-sm">
        {rest.map((r) => (
          <div key={r.id} className="flex items-center gap-3 py-2">
            <div className="w-4 text-center font-display font-bold text-muted-foreground">{r.rank}</div>
            <PlayerAvatar name={r.name} color={r.color} />
            <div className="flex-1 font-display font-bold">{r.name}</div>
            <div className="font-display font-bold text-muted-foreground">{r.totalPoints} p</div>
          </div>
        ))}
      </div>

      <div className="flex flex-row gap-4 w-full">
        <Button size="lg" className="flex-1 min-h-12 font-body transition-all">
          Tillbaka till lobbyn
          <House />
        </Button>
      </div>
    </div>
  );
}
