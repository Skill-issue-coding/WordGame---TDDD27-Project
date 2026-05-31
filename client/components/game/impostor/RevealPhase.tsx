"use client";

import PhaseTransition from "../PhaseTransition";
import { HatGlasses, Users } from "lucide-react";
import { useImpostorGame } from "@/hooks/gamecontext";
import { cn } from "@/lib/utils";
import { ImpostorClientGameState } from "@/lib/game/impostor-types";

// --- Fake data toggles ---
const isImpostor = false;

// --- Fake round state ---
const fakeRoundState: ImpostorClientGameState = {
  timers: { start_time: Date.now(), ready_time: Date.now(), end_time: Date.now() + 10_000 },
  role: isImpostor ? "impostor" : "normal",
  word: "Sommarsemester",
  active_players: { "user-1": true, "user-2": true, "user-3": true, "user-4": false },
};

export function RevealPhase() {
  // const game = useImpostorGame();
  // if (!game || !game.roundState) return null;

  // const isImpostor = game.roundState.role === "impostor";
  const roundState = fakeRoundState;
  // const roundState = game.roundState;

  return (
    <PhaseTransition phaseKey="reveal">
      <div className="flex justify-center w-full mb-3">{isImpostor ? <HatGlasses className="stroke-[2.5] size-12" /> : <Users className="stroke-[2.5] size-12" />}</div>
      <p className={cn("text-sm mb-4 uppercase tracking-wider font-display font-bold whitespace-pre-line text-center", isImpostor ? "text-destructive animate-pulse" : "text-muted-foreground")}>
        {isImpostor ? "Du är impostern! \n Här är ditt ledtrådsord" : "Ditt hemliga ord"}
      </p>
      <div className={cn("game-card mb-6 py-10 border-2", isImpostor ? "border-destructive" : "border-game-green")}>
        <h2 className={cn("font-display text-6xl font-bold", isImpostor ? "text-destructive" : "text-game-green")}>{roundState.word}</h2>
      </div>
      <p className="mb-6 text-sm font-semibold text-center text-muted-foreground font-display">{isImpostor ? "Försök lista ut vad de andra pratar om utan att bli påkommen!" : "Kom ihåg ordet! Låt inte imposters få reda på det."}</p>
    </PhaseTransition>
  );
}
