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
    <button onClick={handleCopy} className="flex items-center gap-3 bg-muted/50 border border-border rounded-xl px-5 py-3 hover:bg-muted/70 transition-colors group cursor-pointer">
      <span className="font-display text-2xl font-bold tracking-[0.3em] text-neon-green text-glow-green">{code}</span>
      {copied ? <Check className="w-5 h-5 text-neon-green" /> : <Copy className="w-5 h-5 text-muted-foreground group-hover:text-foreground transition-colors" />}
    </button>
  );
};

export default RoomCodeDisplay;
