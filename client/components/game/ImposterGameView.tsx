"use client";

import { ArrowLeft, Play, BookOpenText } from "lucide-react";
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

type Phase = "reveal" | "input" | "discuss" | "vote";

const ImpostorGame = () => {
  //const { username, color } = useUser();
  const [phase, setPhase] = useState<Phase>("reveal");
  const [guess, setGuess] = useState("");
  const [voted, setVoted] = useState<number | null>(null);

  //const players = [{ name: username, color }, ...OTHER_NAMES.map((name, i) => ({ name, color: OTHER_COLORS[i] }))];

  return (
    <AnimatePresence mode="wait">
      {phase === "reveal" && (
        <PhaseTransition phaseKey="reveal">
          <div className="w-full max-w-md text-center">
            <div className="text-5xl mb-3">🤫</div>
            <p className="text-muted-foreground text-sm mb-4 uppercase tracking-wider font-display font-bold">Ditt hemliga ord</p>
            <div className="game-card mb-6 border-game-pink py-10">
              <h2 className="font-display text-6xl font-bold text-game-pink">Äpple</h2>
            </div>
            <p className="text-muted-foreground text-sm mb-6 font-display font-semibold">Kom ihåg ordet! Låt inte imposters få reda på det.</p>
            <Button size="lg" onClick={() => setPhase("input")} className="w-full">
              Jag är redo →
            </Button>
          </div>
        </PhaseTransition>
      )}

      {phase === "input" && (
        <PhaseTransition phaseKey="input">
          <div className="w-full max-w-md">
            <CountdownBar duration={5} isRunning onComplete={() => setPhase("discuss")} />
            <div className="mt-6 text-center mb-6">
              <p className="text-muted-foreground text-sm mb-2 font-display uppercase tracking-wider font-bold">Submit a clue</p>
              <p className="text-xs text-muted-foreground font-display">Type a word related to your secret word</p>
            </div>
            <div className="flex gap-3">
              <Input value={guess} onChange={(e) => setGuess(e.target.value)} placeholder="Type a clue..." className="bg-card border-2 h-12 text-lg font-display font-bold rounded-2xl" autoFocus />
              <Button variant="neonPink" size="lg" onClick={() => setPhase("discuss")}>
                Send
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
              <h2 className="font-display text-2xl font-bold text-foreground mb-2">💬 Discussion Phase</h2>
              <p className="text-muted-foreground text-sm font-display font-semibold">Talk it out — who seems suspicious?</p>
            </div>
            <div className="game-card mb-6">
              <h3 className="text-sm font-display font-bold text-muted-foreground uppercase mb-3">Clues</h3>
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
              Skip to Vote
            </Button>
          </div>
        </PhaseTransition>
      )}

      {phase === "vote" && (
        <PhaseTransition phaseKey="vote">
          <div className="w-full max-w-md">
            <CountdownBar duration={10} isRunning />
            <div className="mt-6 text-center mb-6">
              <h2 className="font-display text-2xl font-bold text-foreground mb-2">🗳️ Vote!</h2>
              <p className="text-muted-foreground text-sm font-display font-semibold">Who is the impostor?</p>
            </div>
            <div className="grid grid-cols-2 gap-4 mb-6">
              {/*{players.map((p, i) => (
                <button key={p.name} onClick={() => setVoted(i)} className={`game-card flex flex-col items-center gap-2 transition-all cursor-pointer ${voted === i ? "border-game-pink" : "hover:border-muted-foreground"}`}>
                  <PlayerAvatar name={p.name} color={p.color} size="md" isVoted={voted === i} />
                </button>
              ))}*/}
            </div>
            <Button variant="neonPink" className="w-full" disabled={voted === null}>
              Confirm Vote
            </Button>
          </div>
        </PhaseTransition>
      )}
    </AnimatePresence>
  );
};

export default ImpostorGame;
