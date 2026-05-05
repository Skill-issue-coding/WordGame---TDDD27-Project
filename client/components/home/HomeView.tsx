"use client";

import { useState } from "react";
import { Input } from "@/components/ui/input";
import { Gamepad2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { useGameContext } from "@/hooks/gamecontext";

export default function HomeView() {
  const { sendMessage } = useGameContext();
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

  const handleCreateLobby = () => sendMessage("create_lobby", null);
  return (
    <div className="h-screen w-screen overflow-hidden flex flex-col items-center justify-center p-4 sm:p-6">
      <div className="w-full max-w-md flex flex-col animate-slide-up gap-4">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center gap-2 mb-2">
            <Gamepad2 className="h-14 w-14 text-game-purple" />
            <h1 className="font-display text-4xl sm:text-6xl font-bold text-game-purple">
              Ordio<span className="text-game-pink">Arena</span>
            </h1>
          </div>
          <p className="text-muted-foreground text-base font-display font-semibold">Snabbtänkt multiplayer ordspel</p>
        </div>

        <div className="game-card border-game-blue/30 mt-2">
          <h2 className="font-display text-base font-bold mb-3 text-foreground flex items-center gap-2">Gå med i rum</h2>
          <div className="flex gap-2">
            <Input placeholder="XXXX-XXXX" value={roomCode} onChange={(e) => setRoomCode(formatCode(e.target.value))} className="font-body text-base font-bold tracking-widest text-center bg-muted border-2 h-12 rounded-lg" maxLength={9} />
            <Button size="lg" disabled={roomCode.replace("-", "").length !== 8} className="h-12 px-8 font-bold text-base">
              Gå med
            </Button>
          </div>
        </div>

        <div className="flex items-center gap-3 mt-1.5">
          <div className="flex-1 h-px bg-border" />
          <span className="font-display font-bold text-xs text-muted-foreground uppercase tracking-widest">eller</span>
          <div className="flex-1 h-px bg-border" />
        </div>

        <Button size="lg" className="w-full h-16 text-lg font-bold mb-4" onClick={handleCreateLobby}>
          Skapa rum +
        </Button>
      </div>
    </div>
  );
}
