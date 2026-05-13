"use client";

import { cn } from "@/lib/utils";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Slider } from "@/components/ui/slider";
import { Timer, RefreshCw, Check, HatGlasses } from "lucide-react";
import { useState } from "react";
import CodeDisplay from "@/components/lobby/CodeDisplay";
import { GAME_MODES, getMode, MODE_SETTINGS, ModeSetting, type GameModeId } from "@/lib/game/gameModes";
import { useEffect } from "react";
import { useLobbyContext } from "@/hooks/lobbycontext";
import { useUserContext } from "@/hooks/usercontext";

interface SettingsPanelProps {
  className?: string;
}

function GameMode({ selectedMode, onModeChange, disabled }: { selectedMode: GameModeId; onModeChange: (id: GameModeId) => void; disabled: boolean }) {
  const mode = getMode(selectedMode);

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
        {GAME_MODES.map((m) => {
          const active = selectedMode === m.id;
          return (
            <button
              key={m.id}
              disabled={disabled}
              onClick={() => onModeChange(m.id)}
              className={cn(
                "relative text-left rounded-lg border-2 p-3 transition-all flex items-center gap-3",
                active ? `bg-card border-current ${m.textClass} shadow-md` : "bg-muted/40 border-border hover:border-muted-foreground/40",
                disabled ? "cursor-not-allowed pointer-events-none" : "cursor-pointer opacity-80",
              )}>
              <div className={cn("w-10 h-10 rounded-xl flex items-center justify-center text-2xl shrink-0", `bg-game-${m.color}`)}>{m.icon}</div>
              <div className="flex-1 min-w-0">
                <div className={cn("font-display font-bold text-sm truncate", active ? m.textClass : "text-foreground")}>{m.title}</div>
                <div className="text-xs text-muted-foreground truncate">{m.players}</div>
              </div>
              {active && (
                <div className={cn("w-6 h-6 rounded-full flex items-center justify-center text-white shrink-0", `bg-game-${m.color}`)}>
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

function GameSettings({ config, disabled }: { config: ModeSetting[]; disabled: boolean }) {
  const [values, setValues] = useState<Record<string, any>>({});

  useEffect(() => {
    const defaults = config.reduce(
      (acc, setting) => {
        acc[setting.key] = setting.default;
        return acc;
      },
      {} as Record<string, any>,
    );

    setValues(defaults);
  }, [config]);

  const updateValue = (key: string, val: any) => {
    setValues((prev) => ({ ...prev, [key]: val }));
  };

  return (
    <div className={cn("grid grid-cols-1 md:grid-cols-2 gap-8 w-full", disabled && "opacity-50")}>
      {config.map((setting) => {
        const currentValue = values[setting.key] ?? setting.default;
        return (
          <div key={setting.key} className="flex flex-col gap-2.5 flex-1">
            <div className="flex items-center gap-2 text-sm font-semibold text-foreground">
              {setting.key === "impostorCount" ? <HatGlasses className="w-4 h-4 text-muted-foreground" /> : setting.key.includes("Time") ? <Timer className="w-4 h-4 text-muted-foreground" /> : <RefreshCw className="w-4 h-4 text-muted-foreground" />}
              {setting.label}
            </div>

            {setting.type === "slider" ? (
              <div className={cn("flex items-center gap-4", disabled && "pointer-events-none")}>
                <Slider disabled={disabled} min={setting.min} max={setting.max} step={setting.step} value={[currentValue]} onValueChange={([v]) => updateValue(setting.key, v)} className="flex-1" />
                <span className="text-sm font-bold w-8 text-right tabular-nums">
                  {values[setting.key] ?? setting.default}
                  {setting.key.includes("Time") ? "s" : ""}
                </span>
              </div>
            ) : (
              <ToggleGroup disabled={disabled} type="single" spacing={2} size="sm" value={String(currentValue)} onValueChange={(v) => v && updateValue(setting.key, isNaN(Number(v)) ? v : Number(v))} className={cn(disabled && "pointer-events-none")}>
                {setting.options?.map((opt) => (
                  <ToggleGroupItem
                    key={opt.value}
                    value={String(opt.value)}
                    className={
                      "flex-1 font-display font-bold px-4.5 border-2 transition-all cursor-pointer bg-muted/40 border-border text-muted-foreground hover:border-muted-foreground/40 hover:bg-muted/60 data-[state=on]:bg-primary data-[state=on]:border-primary data-[state=on]:text-primary-foreground data-[state=on]:shadow-md data-[state=on]:opacity-100 disabled:cursor-not-allowed"
                    }>
                    {opt.label}
                  </ToggleGroupItem>
                ))}
              </ToggleGroup>
            )}
          </div>
        );
      })}
    </div>
  );
}

export function SettingsPanel({ className }: SettingsPanelProps) {
  const { lobbyState } = useLobbyContext();
  const { user } = useUserContext();

  const [selectedMode, setSelectedMode] = useState<GameModeId>("impostor");

  const isHost = lobbyState?.host === user?.user_id;
  const settingsConfig = MODE_SETTINGS[selectedMode];

  return (
    <div className={cn("game-card flex-1 p-6 border shadow-sm rounded-2xl bg-card border-border", className)}>
      <div className="flex flex-col gap-4">
        <CodeDisplay />
        <div className="flex items-center justify-between">
          <p className="font-display text-sm font-bold text-muted-foreground uppercase tracking-wider">Spelläge</p>
        </div>
        <GameMode selectedMode={selectedMode} onModeChange={setSelectedMode} disabled={!isHost} />
        <div className="flex items-center justify-between">
          <p className="font-display text-sm font-bold text-muted-foreground uppercase tracking-wider">Spelinställningar</p>
        </div>
        <GameSettings config={settingsConfig} disabled={!isHost} />
      </div>
    </div>
  );
}
