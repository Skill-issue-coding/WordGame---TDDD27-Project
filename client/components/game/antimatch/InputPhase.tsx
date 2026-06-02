"use client";

import { cn } from "@/lib/utils";
import { useState } from "react";
import { Check, Send } from "lucide-react";
import { PlayerAvatar } from "@/components/lobby/PlayerList";
import { Button } from "@/components/ui/button";

// Mock Data
const MOCK_TARGET = "kärlek";
const MOCK_PLAYERS = [
  { id: "1", name: "Du", color: "#8b5cf6", submitted: false, isSelf: true },
  { id: "2", name: "Saga", color: "#10b981", submitted: true, isSelf: false },
  { id: "3", name: "Astrid", color: "#fbd38d", submitted: true, isSelf: false },
  { id: "4", name: "Nils", color: "#f6ad55", submitted: true, isSelf: false },
  { id: "5", name: "Oskar", color: "#4fd1c5", submitted: true, isSelf: false },
];

export function InputPhase() {
  const [inputValue, setInputValue] = useState("");

  return (
    <div className="flex flex-col items-center w-full max-w-2xl mx-auto mt-8 relative z-10">
      <div className="text-center mb-8">
        <p className="font-display text-xs font-bold text-muted-foreground uppercase tracking-widest mb-2">Målord</p>
        <h1 className="font-display text-6xl font-extrabold text-game-green tracking-tight mb-4">{MOCK_TARGET}</h1>
        <p className="font-body font-semibold text-muted-foreground">
          Var närmast — men <span className="text-foreground font-bold">unik</span>. Dubbletter ger 0 poäng.
        </p>
      </div>

      <div className="w-full mb-10 flex gap-2">
        <input
          type="text"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          placeholder="Skriv ett relaterat ord..."
          autoFocus
          className="w-full text-center font-display font-bold text-xl p-4 border-2 border-primary rounded-lg bg-card shadow-sm focus:outline-none focus:ring-4 focus:ring-primary/20 transition-all placeholder:text-muted-foreground/50"
        />
        <Button
          //onClick={sendWordSubmission}
          //disabled={!wordSubmission.trim() || !isCurrentPlayer}
          size="icon"
          className="size-15 shrink-0"
          aria-label="Skicka meddelande">
          <Send className="size-6" />
        </Button>
      </div>

      <div className="flex items-center justify-center gap-4">
        {MOCK_PLAYERS.map((p) => (
          <div key={p.id} className="relative flex flex-col items-center">
            <PlayerAvatar name={p.name} color={p.color} className="w-10 h-10 border-3 font-display font-bold" />
            {p.submitted && (
              <div className="absolute -bottom-6 w-5 h-5 rounded-full flex items-center justify-center">
                <Check className="w-4 h-4 text-game-green" strokeWidth={4} />
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
