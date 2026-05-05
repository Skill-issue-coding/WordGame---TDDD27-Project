"use client";

import { cn } from "@/lib/utils";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Slider } from "@/components/ui/slider";
import { Info, Timer, RefreshCw, Check } from "lucide-react";
import { useState } from "react";
import CodeDisplay from "@/components/lobby/CodeDisplay";
import { GAME_MODES, getMode, type GameModeId } from "@/lib/game/gameModes";

interface SettingsPanelProps {
  className?: string;
}

/*
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
        className="data-[state=on]:border-neon-pink hover:border-neon-pink/60 data-[state=on]:border-game-red cursor-pointer">
        <span className="text-game-red">🕵️ Hitta Impostern</span>
      </ToggleGroupItem>
      <ToggleGroupItem
        value="bottom"
        aria-label="Kontext battle"
        className="data-[state=on]:border-neon-blue hover:border-neon-blue/60 data-[state=on]:border-game-blue cursor-pointer">
        <span className="text-game-blue">🧠 Kontext Strid</span>
      </ToggleGroupItem>
      <ToggleGroupItem
        value="left"
        aria-label="Synonym duel"
        className="data-[state=on]:border-neon-green hover:border-neon-green/60 data-[state=on]:border-game-green cursor-pointer">
        <span className="text-game-green">⚔️ Synonym Duell</span>
      </ToggleGroupItem>
      <ToggleGroupItem
        value="right"
        aria-label="Anti-match"
        className="data-[state=on]:border-neon-yellow hover:border-neon-yellow/60 data-[state=on]:border-game-yellow cursor-pointer">
        <span className="text-game-yellow">🎯 Anti-matchning</span>
      </ToggleGroupItem>
    </ToggleGroup>
  );
}
*/

function GameMode() {
  const [selectedMode, setSelectedMode] = useState<GameModeId>("impostor");
  const mode = getMode(selectedMode);

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {GAME_MODES.map((m) => {
          const active = selectedMode === m.id;
          return (
            <button
              key={m.id}
              onClick={() => setSelectedMode(m.id)} // TODO: Add host checks
              className={cn(
                "relative text-left rounded-lg border-2 p-3 transition-all flex items-center gap-3",
                active
                  ? `bg-card border-current ${m.textClass} shadow-md`
                  : "bg-muted/40 border-border hover:border-muted-foreground/40",
                "cursor-pointer opacity-80",
              )}>
              <div
                className={cn(
                  "w-10 h-10 rounded-xl flex items-center justify-center text-2xl shrink-0",
                  `bg-game-${m.color}`,
                )}>
                {m.icon}
              </div>
              <div className="flex-1 min-w-0">
                <div
                  className={cn("font-display font-bold text-sm truncate", active ? m.textClass : "text-foreground")}>
                  {m.title}
                </div>
                <div className="text-xs text-muted-foreground truncate">{m.players}</div>
              </div>
              {active && (
                <div
                  className={cn(
                    "w-6 h-6 rounded-full flex items-center justify-center text-white shrink-0",
                    `bg-game-${m.color}`,
                  )}>
                  <Check className="w-4 h-4" />
                </div>
              )}
            </button>
          );
        })}
      </div>

      <div className="rounded-lg bg-muted/40 border-2 border-border p-3">
        <p className="text-sm text-muted-foreground leading-snug">{mode.description}</p>
      </div>
    </>
  );
}

function GameSettings() {
  const [roundTime, setRoundTime] = useState(30);
  const [gameRounds, setGameRounds] = useState(3);
  return (
    <div className="flex flex-col sm:flex-row w-full gap-8">
      <div className="flex flex-col gap-2.5 flex-1">
        <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
          <Timer className="w-4 h-4 text-muted-foreground font-display" />
          Betänketid
        </div>
        <div className="flex items-center gap-4">
          <Slider
            name="game-time-slider"
            id="game-time-slider"
            min={10}
            max={60}
            step={5}
            value={[roundTime]}
            onValueChange={([v]) => setRoundTime(v)}
            className="flex-1"
          />
          <span className=" text-sm font-bold text-right text-foreground tabular-nums">{roundTime} s</span>
        </div>
      </div>
      <div className="flex flex-col gap-2.5 flex-1">
        <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
          <RefreshCw className="w-4 h-4 text-muted-foreground font-display" />
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
    <div className={cn("game-card flex-1 p-6 border shadow-sm rounded-2xl bg-card border-border", className)}>
      <div className="flex flex-col gap-4">
        <CodeDisplay code="XXXX-XXXX" />
        <div className="flex items-center justify-between">
          <p className="font-display text-sm font-bold text-muted-foreground uppercase tracking-wider">Spelläge</p>
          {/*<Info className="text-muted-foreground" />*/}
        </div>
        <GameMode />
        <div className="flex items-center justify-between">
          <p className="font-display text-sm font-bold text-muted-foreground uppercase tracking-wider">
            Spelinställningar
          </p>
          {/*<Info className="text-muted-foreground" />*/}
        </div>
        <GameSettings />
      </div>
    </div>
  );
}
