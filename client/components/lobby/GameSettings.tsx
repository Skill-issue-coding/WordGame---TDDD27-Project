"use client";

import { cn } from "@/lib/utils";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Slider } from "@/components/ui/slider";
import { Info, Timer, RefreshCw } from "lucide-react";
import { useState } from "react";

interface SettingsPanelProps {
  className?: string;
}

function GameMode() {
  return (
    <ToggleGroup
      type="single"
      size="lg"
      defaultValue="top"
      variant="outline"
      spacing={2}
      className="grid grid-cols-2 sm:flex sm:grid-rows-none">
      <ToggleGroupItem
        value="top"
        aria-label="Imposter game"
        className="data-[state=on]:border-neon-pink hover:border-neon-pink/60 data-[state=on]:glow-border-pink/30 cursor-pointer">
        <span className="text-neon-pink">🕵️ Hitta Impostern</span>
      </ToggleGroupItem>
      <ToggleGroupItem
        value="bottom"
        aria-label="Kontext battle"
        className="data-[state=on]:border-neon-blue hover:border-neon-blue/60 data-[state=on]:glow-border-blue/30 cursor-pointer">
        <span className="text-neon-blue">🧠 Kontext Strid</span>
      </ToggleGroupItem>
      <ToggleGroupItem
        value="left"
        aria-label="Synonym duel"
        className="data-[state=on]:border-neon-green hover:border-neon-green/60 data-[state=on]:glow-border-green/30 cursor-pointer">
        <span className="text-neon-green">⚔️ Synonym Duell</span>
      </ToggleGroupItem>
      <ToggleGroupItem
        value="right"
        aria-label="Anti-match"
        className="data-[state=on]:border-neon-yellow hover:border-neon-yellow/60 data-[state=on]:glow-border-yellow/30 cursor-pointer">
        <span className="text-neon-yellow">🎯 Anti-matchning</span>
      </ToggleGroupItem>
    </ToggleGroup>
  );
}

function GameSettings() {
  const [roundTime, setRoundTime] = useState(10);
  const [gameRounds, setGameRounds] = useState(3);
  return (
    <div className="flex flex-col sm:flex-row w-full gap-8">
      <div className="flex flex-col gap-2.5 flex-1">
        <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
          <Timer className="w-4 h-4 text-muted-foreground" />
          Betänketid
        </div>
        <div className="flex items-center gap-4">
          <Slider
            name="game-time-slider"
            id="game-time-slider"
            min={5}
            max={30}
            step={1}
            value={[roundTime]}
            onValueChange={([v]) => setRoundTime(v)}
            className="flex-1"
          />
          <span className=" text-sm font-bold text-right text-foreground tabular-nums">{roundTime} s</span>
        </div>
      </div>
      <div className="flex flex-col gap-2.5 flex-1">
        <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
          <RefreshCw className="w-4 h-4 text-muted-foreground" />
          Omgångar
        </div>
        <div className="flex items-center gap-4">
          <Slider
            name="game-time-slider"
            id="game-time-slider"
            min={1}
            max={5}
            step={1}
            value={[gameRounds]}
            onValueChange={([v]) => setGameRounds(v)}
            className="flex-1"
          />
          <span className=" text-sm font-bold text-right text-foreground tabular-nums">{gameRounds}</span>
        </div>
      </div>
    </div>
  );
}

export function SettingsPanel({ className }: SettingsPanelProps) {
  return (
    <div className={cn("flex-1 p-6 border shadow-sm rounded-2xl bg-card border-border", className)}>
      <div className="flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <p className="text-base font-bold text-foreground">Spelläge</p>
          <Info className="text-muted-foreground" />
        </div>
        <GameMode />
        <div className="flex items-center justify-between">
          <p className="text-base font-bold text-foreground">Spelinställningar</p>
          {/*<Info className="text-muted-foreground" />*/}
        </div>
        <GameSettings />
      </div>
    </div>
  );
}
