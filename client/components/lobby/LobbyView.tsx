"use client";
import { Gamepad2, ArrowLeft, User, Timer } from "lucide-react";
import GameModeCard, { Color } from "@/components/lobby/GameModeCard";
import CodeDisplay from "@/components/lobby/CodeDisplay";
import Link from "next/link";
import { Slider } from "../ui/slider";
import { Label } from "../ui/label";
import { useState } from "react";
import { PlayerList } from "@/components/lobby/PlayerList";

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
          <PlayerList className={"mt-6"} />
        </div>

        <div className="flex flex-1 w-full h-full grid-cols-1 gap-4 mt-6 sm:grid-cols-2">
          <div className="flex flex-col w-full gap-2 p-4 px-8 border-2 flex-5 bg-game-surface">
            <div className="flex justify-center gap-2 pb-6">
              <h2 className="text-2xl font-bold font-display text-glow-green text-neon-green">Spel inställningar</h2>
              <Gamepad2 className="w-8 h-8 text-neon-green" />
            </div>
            <div className="flex flex-col flex-1 gap-6">
              <div className="flex items-center">
                <h3 className="font-bold font-display text-neon-green"> Gamemode: </h3>
              </div>
              <div className="flex flex-col w-full gap-3">
                <div className="flex w-full gap-2">
                  <Timer className="font-bold font-display text-neon-green size-5 aspect-square " />
                  <Label htmlFor="game-time-slider" className="font-bold font-display text-neon-green">
                    Speltid:
                  </Label>
                </div>
                <div className="flex w-full gap-6 flex-start">
                  <Slider
                    name="game-time-slider"
                    id="game-time-slider"
                    min={5}
                    max={30}
                    step={1}
                    value={[gametime]}
                    onValueChange={([v]) => setGametime(v)}
                    className="w-full"
                  />
                  <p>{gametime}min</p>
                </div>
              </div>
            </div>
          </div>
          <div className="flex flex-col gap-4 flex-4">
            {gamemodes.map((mode) => (
              <GameModeCard key={mode.id} {...mode} />
            ))}
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
