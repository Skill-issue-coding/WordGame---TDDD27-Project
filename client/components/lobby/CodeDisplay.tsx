"use client";

import { Copy, Check } from "lucide-react";
import { useState } from "react";

interface RoomCodeDisplayProps {
  code: string;
}

const RoomCodeDisplay = ({ code }: RoomCodeDisplayProps) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <button
      onClick={handleCopy}
      className="flex items-center gap-3 px-5 py-3 transition-colors border cursor-pointer bg-muted/50 border-border rounded-xl hover:bg-muted/70 group">
      <span className="font-display text-2xl font-bold tracking-[0.3em] text-neon-green text-glow-green">{code}</span>
      {copied ? (
        <Check className="w-5 h-5 text-neon-green" />
      ) : (
        <Copy className="w-5 h-5 transition-colors text-muted-foreground group-hover:text-foreground" />
      )}
    </button>
  );
};

export default RoomCodeDisplay;
