"use client";

import PhaseTransition from "../PhaseTransition";
import { HatGlasses, Users } from "lucide-react";
import { useImpostorGame } from "@/hooks/gamecontext";
import { cn } from "@/lib/utils";

export function RevealPhase() {
  const game = useImpostorGame();
  if (!game || !game.roundState) return null;

  const isImpostor = game.roundState.role === "impostor";

  return (
    <PhaseTransition phaseKey="reveal">
      <div className="w-full max-w-md text-center">
        <div className="text-5xl mb-3">{isImpostor ? <HatGlasses /> : <Users />}</div>
        <p className={cn("text-sm mb-4 uppercase tracking-wider font-display font-bold whitespace-pre-line", isImpostor ? "text-destructive animate-pulse" : "text-muted-foreground")}>{isImpostor ? "Du är impostern! \n Här är ditt ledtrådsord" : "Ditt hemliga ord"}</p>
        <div className={cn("game-card mb-6 py-10 border-2", isImpostor ? "border-destructive" : "border-game-green")}>
          <h2 className={cn("font-display text-6xl font-bold", isImpostor ? "text-destructive" : "text-game-green")}>{game.roundState.word}</h2>
        </div>
        <p className="text-muted-foreground text-sm mb-6 font-display font-semibold">{isImpostor ? "Försök lista ut vad de andra pratar om utan att bli påkommen!" : "Kom ihåg ordet! Låt inte imposters få reda på det."}</p>
      </div>
    </PhaseTransition>
  );
}
