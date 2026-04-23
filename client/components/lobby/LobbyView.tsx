"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Gamepad2, ArrowRight, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import GameModeCard from "./GameModeCard";
import { cn } from "@/lib/utils";
import CodeDisplay from "@/components/lobby/CodeDisplay";
import Link from "next/link";

const gamemodes = [
  {
    id: "imposter",
    title: "Imposter Word",
    description: "One player get a different word. Find the imposter before it is to late!",
    icon: "🕵️",
    players: "4+ players",
    color: "pink" as const,
  },
  {
    id: "contexto",
    title: "Context Battle",
    description: "Race to guess the hidden word.",
    icon: "🧠",
    players: "2-8 players",
    color: "blue" as const,
  },
  {
    id: "synonym",
    title: "Synonym Showdown",
    description: "Submit the best synonym each round. The worst gets eliminated!",
    icon: "⚔️",
    players: "3-8 players",
    color: "green" as const,
  },
  {
    id: "antimatch",
    title: "Anti-Match",
    description: "Think different! Write a synonym but be careful if your word matches another players, you both score zero!",
    icon: "🎯",
    players: "3-8 players",
    color: "yellow" as const,
  },
];

export default function LobbyView() {
  const [roomCode, setRoomCode] = useState("");
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6">
      <div className="w-full max-w-2xl animate-slide-up">
        <div className="text-center mb-12">
          <h1 className="font-display text-4xl font-bold text-glow-green text-neon-green">XXXX's rum</h1>
        </div>

        <Link href="/">
          <button className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors mb-6 cursor-pointer">
            <ArrowLeft className="w-4 h-4" />
            <span className="text-sm">Tillbaka</span>
          </button>
        </Link>

        <CodeDisplay code="XXXX-XXXX" />

        <div className="game-card border-2 text-left mt-6">
          <h2 className="font-body text-lg font-semibold mb-4 text-foreground">X spelare anslutna</h2>
          <div className="h-6 bg-background border-2"></div>
          <div className="h-6 bg-background border-2"></div>
          <div className="h-6 bg-background border-2"></div>
          <div className="h-6 bg-background border-2"></div>
        </div>

        <div className="mt-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {gamemodes.map((mode) => (
              <GameModeCard key={mode.id} {...mode} onClick={() => console.log("click")} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function formatCode(value: string) {
  return value;
}
