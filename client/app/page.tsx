"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Gamepad2 } from "lucide-react";
import { Button } from "@/components/ui/button";

const gamemodes = [
  {
    id: "imposter",
    title: "Imposter Word",
    description: "One player get a different word. Find the imposter before it is to late!",
    icon: "🕵️",
    players: "4+ {User}",
    color: "red" as const,
  },
  {
    id: "contexto",
    title: "Context Battle",
    description: "Race to guess the hidden word.",
    icon: "🧠",
    players: "2-8 {User}",
    color: "blue" as const,
  },
  {
    id: "synonym",
    title: "Synonym Showdown",
    description: "Submit the best synonym each round. The worst gets eliminated!",
    icon: "⚔️",
    players: "3-8 {User}",
    color: "green" as const,
  },
  {
    id: "antimatch",
    title: "Anti-Match",
    description: "Think different! Write a synonym but be careful if your word matches another players, you both score zero!",
    icon: "🎯",
    players: "3-8 {User}",
    color: "yellow" as const,
  },
];

export default function Home() {
  const [roomCode, setRoomCode] = useState("");
  const [selectMode, setSelectMode] = useState<string | null>(null);

  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6">
      <div className="w-full max-w-2xl animate-slide-up">
        {/* Logga */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center gap-3 mb-3">
            <Gamepad2 className="w-10 h-10 text-neon-green" />
            <h1 className="font-display text-5xl font-blod text-glow-green text-neon-green">WordArena</h1>
          </div>
          <p className="text-muted-foreground text-lg">Fast-paced multiplayer word games</p>
        </div>

        {/* Join room */}
        <div className="card-game mb-8">
          <h2 className="font-display text-lg font-semibold mb-4 text-foreground">Join a Room</h2>
          <div className="flex gap-3">
            <Input placeholder="XXXX-XXXX" value={roomCode} onChange={(e) => setRoomCode(e.target.value)} className="font-display text-xl tracking-widest text-center bg-muted/50 border-border h-12" maxLength={9} />
            <Button /* variant="noenGreen" */ size="lg" /* onClick={handleJoin} */ disabled={roomCode.replace("-", "").length !== 8}>
              Join
            </Button>
          </div>
        </div>

        {/* Create Room */}
        <div>
          <h2 className="font-display text-lg font-semibold mb-4 text-foreground">Create a Room</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/*{gamemodes.map((mode) => (
              <GameModeCard
                key={mode.id}
                {...mode}
                onClick={() => handleCreate(mode.id)}
              />
            ))}*/}
          </div>
        </div>
      </div>
    </div>
  );
}
