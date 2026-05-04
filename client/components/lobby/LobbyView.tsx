"use client";
import { Gamepad2, ArrowLeft, User, Timer, Play, BookOpenText } from "lucide-react";
import GameModeCard, { Color } from "@/components/lobby/GameModeCard";
import Link from "next/link";
import { Slider } from "../ui/slider";
import { Label } from "../ui/label";
import { useState } from "react";
import { PlayerList } from "@/components/lobby/PlayerList";
import { SettingsPanel } from "@/components/lobby/GameSettings";
import { Button } from "@/components/ui/button";

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
    color: "pink" as Color,
  },
  {
    id: "contexto",
    title: "Kontext Strid",
    description: "Race att hitta det dålda ordet.",
    icon: "🧠",
    players: "2-12",
    color: "blue" as Color,
  },
  {
    id: "synonym",
    title: "Synonym Duell",
    description: "Ange den bästa synonymen varje runda. Det sämsta åker ut!",
    icon: "⚔️",
    players: "3-12",
    color: "green" as Color,
  },
  {
    id: "antimatch",
    title: "Anti-Matchning",
    description:
      "Tänk anorlunda! Skriv en synonym men var försiktig så det inte matchar någon annans, då får båda noll poäng!",
    icon: "🎯",
    players: "3-12",
    color: "yellow" as Color,
  },
];

export default function LobbyView() {
  // const [roomCode, setRoomCode] = useState("");
  const [gametime, setGametime] = useState(10);
  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-6">
      <div className="w-full max-w-4xl animate-slide-up">
        <div className="flex items-center flex-1 mb-6 text-center">
          <Link href="/" className="flex items-center w-full h-full">
            <button className="flex items-center gap-2 transition-colors cursor-pointer text-muted-foreground hover:text-foreground">
              <ArrowLeft className="w-4 h-4" />
              <span className="text-sm">Tillbaka</span>
            </button>
          </Link>
          {/*<h1 className="w-full text-4xl font-bold font-display text-glow-green text-neon-green">XXXX's rum</h1>*/}
          <div className="w-full" />
        </div>
        <div>
          <div className="flex flex-col sm:grid sm:grid-cols-5 gap-6">
            <SettingsPanel className={"sm:col-span-3"} />
            <PlayerList className={"sm:col-span-2"} />
          </div>
          <div className="flex flex-col sm:flex-row mt-6 gap-6">
            <Button variant="glass" size="lg" className="gap-2 flex-1 min-h-12 font-body">
              Snabb Guide
              <BookOpenText />
            </Button>

            <Button size="lg" className="gap-2 flex-1 min-h-12 font-body">
              Starta
              <Play />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

function PlayerCard() {
  return (
    <div className="flex items-center gap-2 border-2 bg-background">
      1. <User className="size-6" />
    </div>
  );
}
