"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Gamepad2, ArrowRight, ArrowLeft, User } from "lucide-react";
import { Button } from "@/components/ui/button";
import GameModeCard from "@/components/lobby/GameModeCard";
import { cn } from "@/lib/utils";
import CodeDisplay from "@/components/lobby/CodeDisplay";
import Link from "next/link";
import { Slider } from "../ui/slider";
import { Label } from "../ui/label";

// 12 distinct player colors
const PLAYER_COLORS = [
  "#00e5a0",
  "#7c5cfc",
  "#ff5c8a",
  "#ffa726",
  "#42a5f5",
  "#ab47bc",
  "#ef5350",
  "#26c6da",
  "#fdd835",
  "#43a047",
  "#3f51b5",
  "#8d6e63",
];

const gamemodes = [
  {
    id: "imposter",
    title: "Hitta Impostern",
    description: "En spelare får ett unikt ord. Hitta imposter innan det är försent!",
    icon: "🕵️",
    players: "4-12",
    color: "pink" as const,
  },
  {
    id: "contexto",
    title: "Kontext Strid",
    description: "Race att hitta det dålda ordet.",
    icon: "🧠",
    players: "2-12",
    color: "blue" as const,
  },
  {
    id: "synonym",
    title: "Synonym Duell",
    description: "Ange den bästa synonymen varje runda. Det sämsta åker ut!",
    icon: "⚔️",
    players: "3-12",
    color: "green" as const,
  },
  {
    id: "antimatch",
    title: "Anti-Matchning",
    description: "Tänk anorlunda! Skriv en synonym men var försiktig så det inte matchar någon annans, då får båda noll poäng!",
    icon: "🎯",
    players: "3-12",
    color: "yellow" as const,
  },
];

export default function LobbyView() {
  const [roomCode, setRoomCode] = useState("");
  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-6">
      <div className="w-full max-w-3xl animate-slide-up">
        <div className="flex items-center flex-1 mb-12 text-center">
          <Link href="/" className="flex items-center w-full h-full">
            <button className="flex items-center gap-2 transition-colors cursor-pointer text-muted-foreground hover:text-foreground">
              <ArrowLeft className="w-4 h-4" />
              <span className="text-sm">Tillbaka</span>
            </button>
          </Link>
          <h1 className="w-full text-4xl font-bold font-display text-glow-green text-neon-green">XXXX's rum</h1>
          <div className="w-full" />
        </div>
        <div>
          <CodeDisplay code="XXXX-XXXX" />
          <div className="mt-6 text-left border-2 game-card">
            <h2 className="mb-4 text-sm font-semibold tracking-wider font-body text-muted-foreground">(1/12 spelare)</h2>
            <div className="grid grid-cols-2 gap-x-4">
              {Array.from({ length: 12 }).map((_, i) => (
                <PlayerCard key={i} />
              ))}
            </div>
          </div>
        </div>

        <div className="flex flex-1 w-full h-full grid-cols-1 gap-4 mt-6 sm:grid-cols-2">
          <div className="flex flex-col w-full gap-2 p-4 border-2 flex-5 bg-game-surface">
            <div className="flex justify-center">
              <h2 className="text-2xl font-bold font-display text-glow-green text-neon-green">Spel inställningar</h2>
            </div>
            <div className="flex flex-1">
              <div className="flex flex-col w-full gap-3">
                <Label>Tid slider</Label>
                <Slider className="" />
              </div>
            </div>
          </div>
          <div className="flex flex-col gap-4 flex-3">
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

function PlayerCard() {
  return (
    <div className="flex items-center gap-2 border-2 bg-background">
      1. <User className="size-6" />
    </div>
  );
}
