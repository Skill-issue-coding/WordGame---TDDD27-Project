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
    <div className="min-h-screen flex flex-col items-center justify-center p-6">
      <div className="w-full max-w-3xl animate-slide-up">
        {/* Logga */}
        <div className="text-center mb-12">
          <div className="flex items-center justify-center gap-3 mb-3">
            <Gamepad2 className="w-10 h-10 text-neon-green" />
            <h1 className="font-display text-5xl font-bold text-glow-green text-neon-green">OrdioArena.io </h1>
          </div>
          <p className="text-muted-foreground text-lg">Snabbtänkt multiplayer ordspel</p>
        </div>

        <div className="game-card mb-8">
          <h2 className="font-body text-lg font-semibold mb-4 text-foreground">Gå med i rum</h2>
          <div className="flex gap-3">
            <Input
              placeholder="XXXX-XXXX"
              value={roomCode}
              onChange={(e) => setRoomCode(formatCode(e.target.value))}
              className="font-display text-xl tracking-widest text-center bg-muted/50 border-border h-12"
              maxLength={9}
            />
            <Button variant="neonGreen" size="xl" /* onClick={handleJoin} */ disabled={roomCode.replace("-", "").length !== 8}>
              Gå med
            </Button>
          </div>
        </div>

        <Link href="/lobby">
          <Button
            className={cn(
              "game-card border-2 text-left transition-all duration-300 hover:scale-[1.02] cursor-pointer group w-full border-neon-green/30 hover:border-neon-green/60 hover:glow-border-green",
            )}>
            <h3 className={cn("font-body text-lg font-bold mb-1 text-neon-green")}>Skapa rum</h3>
            <ArrowRight className="text-neon-green w-5 h-5" />
          </Button>
        </Link>
      </div>
    </div>
  );
}
