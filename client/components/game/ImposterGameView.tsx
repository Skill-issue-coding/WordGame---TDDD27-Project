"use client";

import { ArrowRight, Play, BookOpenText } from "lucide-react";
import { AnimatePresence } from "framer-motion";
import PhaseTransition from "@/components/game/PhaseTransition";
import CountdownBar from "@/components/game/CountdownBar";
import Link from "next/link";
import { PlayerList } from "@/components/lobby/PlayerList";
import { SettingsPanel } from "@/components/lobby/GameSettings";
import { Button } from "@/components/ui/button";
import { useGameContext } from "@/hooks/gamecontext";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

type Phase = "reveal" | "input" | "discuss" | "vote";

const ImpostorGame = () => {
  //const { username, color } = useUser();
  const [phase, setPhase] = useState<Phase>("reveal");
  const [guess, setGuess] = useState("");
  const [voted, setVoted] = useState<number | null>(null);

  const isImpostor = true; // Temporary toggle for testing

  //const players = [{ name: username, color }, ...OTHER_NAMES.map((name, i) => ({ name, color: OTHER_COLORS[i] }))];

  return (
    <div className="pt-5">
      <AnimatePresence mode="wait">
        {phase === "reveal" && (
          <PhaseTransition phaseKey="reveal">
            <div className="w-full max-w-md text-center">
              <div className="text-5xl mb-3">{isImpostor ? "🕵️" : "🤫"}</div>
              <p
                className={cn(
                  "text-sm mb-4 uppercase tracking-wider font-display font-bold whitespace-pre-line",
                  isImpostor ? "text-destructive animate-pulse" : "text-muted-foreground",
                )}>
                {isImpostor ? "Du är impostern! \n Här är ditt ledtrådsord" : "Ditt hemliga ord"}
              </p>
              <div
                className={cn(
                  "game-card mb-6 py-10 border-2",
                  isImpostor ? "border-destructive" : "border-game-green",
                )}>
                {/* Change word color */}
                <h2
                  className={cn(
                    "font-display text-6xl font-bold",
                    isImpostor ? "text-destructive" : "text-game-green",
                  )}>
                  Äpple
                </h2>
              </div>
              <p className="text-muted-foreground text-sm mb-6 font-display font-semibold">
                {isImpostor
                  ? "Försök lista ut vad de andra pratar om utan att bli påkommen!"
                  : "Kom ihåg ordet! Låt inte imposters få reda på det."}
              </p>
              <Button size="lg" onClick={() => setPhase("input")} className="w-full">
                Jag är redo <ArrowRight />
              </Button>
            </div>
          </PhaseTransition>
        )}

        {phase === "input" && (
          <PhaseTransition phaseKey="input">
            <div className="w-full max-w-md flex flex-col items-center">
              <CountdownBar duration={5} isRunning onComplete={() => setPhase("discuss")} />
              <div className="mt-6 text-center mb-6">
                <p
                  className={cn(
                    "text-muted-foreground text-sm mb-2 font-display uppercase tracking-wider font-bold",
                    isImpostor ? "text-destructive" : "text-muted-foreground",
                  )}>
                  {isImpostor ? "Hitta på en bluff" : "Ange en ledtråd"}
                </p>
                <p className="text-xs text-muted-foreground font-display">
                  {isImpostor
                    ? "Välj ett ord som får dig att smälta in i gruppen"
                    : "Skriv ett ord relaterat till ditt hemliga ord"}
                </p>
              </div>
              <div className="flex gap-3 w-full items-center justify-center">
                <Input
                  value={guess}
                  onChange={(e) => setGuess(e.target.value)}
                  placeholder={isImpostor ? "Skriv en bluff..." : "Skriv en ledtråd..."}
                  className="bg-card border-2 h-12 text-lg font-display font-bold rounded-2xl"
                  autoFocus
                />
                <Button size="lg" onClick={() => setPhase("discuss")}>
                  Skicka
                </Button>
              </div>
              <div className="flex justify-center gap-4 mt-8">
                {/*{players.map((p) => (
                <PlayerAvatar key={p.name} name={p.name} color={p.color} size="sm" />
              ))}*/}
              </div>
            </div>
          </PhaseTransition>
        )}

        {phase === "discuss" && (
          <PhaseTransition phaseKey="discuss">
            <div className="w-full max-w-md">
              <CountdownBar duration={15} isRunning onComplete={() => setPhase("vote")} />
              <div className="mt-6 text-center mb-6">
                <h2 className="font-display text-2xl font-bold text-foreground mb-2">
                  💬 Diskussions fas
                </h2>
                <p className="text-muted-foreground text-sm font-display font-semibold">
                  Berätta — vem är misstänksam?
                </p>
              </div>
              <div className="game-card mb-6">
                <h3 className="text-sm font-display font-bold text-muted-foreground uppercase mb-3">
                  Ledtrådar
                </h3>
                <div className="space-y-2">
                  {["fruit", "red", "juice", "orchard"].map((clue, i) => (
                    <div key={i} className="flex items-center gap-3 text-sm">
                      {/*<PlayerAvatar name={players[i].name} color={players[i].color} size="sm" />*/}
                      <span className="font-display font-bold text-foreground">{clue}</span>
                    </div>
                  ))}
                </div>
              </div>
              <Button variant="neonBlue" className="w-full" onClick={() => setPhase("vote")}>
                Skippa att rösta
              </Button>
            </div>
          </PhaseTransition>
        )}

        {phase === "vote" && (
          <PhaseTransition phaseKey="vote">
            <div className="w-full max-w-md">
              <CountdownBar duration={10} isRunning />
              <div className="mt-6 text-center mb-6">
                <h2 className="font-display text-2xl font-bold text-foreground mb-2">🗳️ Rösta!</h2>
                <p className="text-muted-foreground text-sm font-display font-semibold">
                  Vem är impostern?
                </p>
              </div>
              <div className="grid grid-cols-2 gap-4 mb-6">
                {/*{players.map((p, i) => (
                <button key={p.name} onClick={() => setVoted(i)} className={`game-card flex flex-col items-center gap-2 transition-all cursor-pointer ${voted === i ? "border-game-pink" : "hover:border-muted-foreground"}`}>
                  <PlayerAvatar name={p.name} color={p.color} size="md" isVoted={voted === i} />
                </button>
              ))}*/}
              </div>
              <Button variant="neonPink" className="w-full" disabled={voted === null}>
                Bekräfta röstning
              </Button>
            </div>
          </PhaseTransition>
        )}
      </AnimatePresence>
    </div>
  );
};

export default ImpostorGame;
