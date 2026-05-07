import { useEffect, useState } from "react";
import { cn } from "@/lib/utils";

interface CountdownBarProps {
  duration: number;
  isRunning: boolean;
  onComplete?: () => void;
  color?: "green" | "orange" | "red";
}

const barColorMap = {
  green: "--heat-hot",
  orange: "--heat-warm",
  red: "--heat-cold",
};

const CountdownBar = ({ duration, isRunning, onComplete, color = "green" }: CountdownBarProps) => {
  const [timeLeft, setTimeLeft] = useState(duration);

  useEffect(() => {
    setTimeLeft(duration);
  }, [duration, isRunning]);

  useEffect(() => {
    if (!isRunning) return;
    if (timeLeft <= 0) {
      onComplete?.();
      return;
    }
    const timer = setInterval(() => setTimeLeft((t) => t - 0.05), 50);
    return () => clearInterval(timer);
  }, [isRunning, timeLeft, onComplete]);

  const percent = (timeLeft / duration) * 100;
  const isUrgent = percent < 25;

  return (
    <div className="w-full">
      <div className="flex justify-center items-center mb-2">
        <span className={cn("font-display text-3xl font-bold tabular-nums px-4 py-1 rounded-xl bg-card border-2 border-border", isUrgent && "text-destructive border-destructive animate-pulse")}>{Math.ceil(timeLeft)}</span>
      </div>
      <div className="h-3 bg-muted rounded-full overflow-hidden border border-border">
        <div className={cn("h-full rounded-full transition-all duration-100", barColorMap[color], isUrgent && "bg-destructive")} style={{ width: `${Math.max(0, percent)}%` }} />
      </div>
    </div>
  );
};

export default CountdownBar;
