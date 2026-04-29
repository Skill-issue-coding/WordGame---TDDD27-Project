"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Gamepad2 } from "lucide-react";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import Link from "next/link";

export default function HomeView() {
  const [roomCode, setRoomCode] = useState("");
  const [selectMode, setSelectMode] = useState<string | null>(null);

  const formatCode = (val: string) => {
    const clean = val
      .replace(/[^a-zA-Z0-9]/g, "")
      .toUpperCase()
      .slice(0, 8);
    if (clean.length > 4) return clean.slice(0, 4) + "-" + clean.slice(4);
    return clean;
  };

  return (
    <div className="min-h-screen overflow-hidden flex flex-col items-center justify-center p-4 sm:p-6">
      <div className="w-full max-w-md flex flex-col animate-slide-up">
        {/* Logga */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center gap-2 mb-2">
            <Gamepad2 className="text-4xl sm:text-5x1" />
            <h1 className="font-display text-4xl sm:text-6x1 font-bold text-game-blue">
              Ordio<span className="text-game-red">Arena.io</span>
            </h1>
          </div>
          <p className="text-muted-foreground text-base font-display font-semibold">Snabbtänkt multiplayer ordspel</p>
        </div>

        <div className="game-card border-game-blue/30 mt-2">
          <h2 className="font-display text-base font-bold mb-3 text-foreground flex items-center gap-2">Gå med i rum</h2>
          <div className="flex gap-2">
            <Input
              placeholder="XXXX-XXXX"
              value={roomCode}
              onChange={(e) => setRoomCode(formatCode(e.target.value))}
              className="font-display text-xl font-bold tracking-widest text-center bg-muted border-2 h-12 rounded-2x1"
              maxLength={9}
            />
            <Button size="lg" /* onClick={handleJoin} */ disabled={roomCode.replace("-", "").length !== 8} className="h-12">
              Gå med
            </Button>
          </div>
        </div>

        <div className="flex items-center gap-3 my-2">
          <div className="flex-1 h-px bg-border" />
          <span className="font-display font-bold text-xs text-muted-foreground uppercase tracking-widest">eller</span>
          <div className="flex-1 h-px bg-border" />
        </div>

        <Link href="/lobby">
          <Button
            className={cn(
              "game-card border-2 text-left transition-all duration-300 hover:scale-[1.02] cursor-pointer group w-full border-neon-green/30 hover:border-neon-green/60 hover:glow-border-green",
            )}>
            <h3 className={cn("font-body text-lg font-bold mb-1 text-blue")}>Skapa rum</h3>
            <ArrowRight className="text-neon-green w-5 h-5" />
          </Button>
        </Link>
      </div>
    </div>
  );
}
