"use client";

import { useLobbyContext } from "@/hooks/lobbycontext";
import { Copy, Check } from "lucide-react";
import { useState } from "react";

const RoomCodeDisplay = () => {
  const [copied, setCopied] = useState(false);
  const { lobbyState } = useLobbyContext();

  const handleCopy = () => {
    navigator.clipboard.writeText(lobbyState?.code || "xxxx-xxxx");
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const chars = (lobbyState?.code || "xxxx-xxxx").split("");

  return (
    <button onClick={handleCopy} className="flex items-center justify-center gap-3 px-5 py-2.5 transition-colors border cursor-pointer bg-muted/40 border-border rounded-lg hover:bg-border/50 group">
      {/*<span className="font-display text-2xl font-bold tracking-[0.3em] text-neon-green text-glow-green">{code}</span>*/}
      <div className="flex items-center gap-1.5">
        {chars.map((c, i) =>
          c === "-" ? (
            <span key={i} className="text-2xl font-display font-bold text-muted-foreground px-1">
              -
            </span>
          ) : (
            <span key={i} className="game-tile w-10 h-12 flex items-center justify-center font-display text-2xl font-bold text-game-purple">
              {c}
            </span>
          ),
        )}
      </div>
      {copied ? <Check className="w-5 h-5 text-game-green" /> : <Copy className="w-5 h-5 transition-colors text-muted-foreground group-hover:text-foreground" />}
    </button>
  );
};

export default RoomCodeDisplay;
