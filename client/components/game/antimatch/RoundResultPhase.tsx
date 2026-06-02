"use client";

import { PlayerAvatar } from "@/components/lobby/PlayerList";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ArrowRight } from "lucide-react";

// Mock Data
const MOCK_TARGET = "kärlek";
const MOCK_RESULTS = [
  { id: "2", name: "Oskar", color: "#4fd1c5", word: "romantik", points: 98, isDuplicate: false, rank: 1 },
  { id: "3", name: "Astrid", color: "#fbd38d", word: "rosor", points: 91, isDuplicate: false, rank: 2 },
  { id: "1", name: "Du", color: "#8b5cf6", word: "—", points: 0, isDuplicate: false, rank: 3 },
  { id: "4", name: "Saga", color: "#10b981", word: "hjärta", points: 0, isDuplicate: true, rank: 4 },
  { id: "5", name: "Nils", color: "#f6ad55", word: "hjärta", points: 0, isDuplicate: true, rank: 5 },
];

export function RoundResultPhase() {
  return (
    <div className="flex flex-col items-center w-full max-w-2xl mx-auto mt-8 relative z-10">
      <div className="text-center mb-8">
        <h2 className="font-display text-3xl font-extrabold mb-2">Resultat till: "{MOCK_TARGET}"</h2>
        <p className="font-body font-semibold text-muted-foreground">"hjärta" x2</p>
      </div>

      <div className="w-full flex flex-col gap-3 mb-8">
        {MOCK_RESULTS.map((r) => {
          const isWinner = r.rank === 1;
          const isDupe = r.isDuplicate;

          return (
            <div
              key={r.id}
              className={cn(
                "flex items-center gap-4 p-3 rounded-xl border-2 border-border transition-all bg-card", // bg-card ligger alltid i botten
                isDupe
                  ? "border-game-red bg-linear-to-r from-game-red/20 to-game-red/20"
                  : isWinner
                    ? "border-game-green bg-linear-to-r from-game-green/20 to-game-green/20"
                    : "border-border",
              )}>
              <div className="w-6 text-center font-display font-extrabold text-xl">{r.rank}</div>

              <PlayerAvatar name={r.name} color={r.color} className="w-9 h-9 border-3 font-display font-bold" />

              <div className="flex-1 min-w-0">
                <div className="font-display font-bold text-xs text-muted-foreground">{r.name}</div>
                <div className="font-display font-extrabold text-xl truncate flex items-center gap-2">
                  "{r.word}"
                  {isDupe && (
                    <span className="text-[10px] font-display font-extrabold text-game-red uppercase tracking-wider">
                      · Dubblett
                    </span>
                  )}
                </div>
              </div>

              <div className={cn("font-display font-extrabold text-xl", isDupe ? "text-game-red" : "text-game-green")}>
                {isDupe ? "0p" : `+${r.points}p`}
              </div>
            </div>
          );
        })}
      </div>

      {/*
      <Button size="lg" className="flex-1 min-h-12 font-body transition-all w-full">
        Nästa runda
        <ArrowRight />
      </Button>
      */}
      <Button size="lg" className="flex-1 min-h-12 font-body transition-all w-full">
        Slutresultat
        <ArrowRight />
      </Button>
    </div>
  );
}
