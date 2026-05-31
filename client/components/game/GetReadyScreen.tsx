"use client";

import { motion } from "framer-motion";
import { useEffect, useState } from "react";

interface GetReadyScreenProps {
  remainingMs: number;
}

// --- Fake data ---
const FAKE_REMAINING_MS = 2_500;

export function GetReadyScreen({ remainingMs: _remainingMs }: GetReadyScreenProps) {
  const [fakeMs, setFakeMs] = useState(FAKE_REMAINING_MS);

  useEffect(() => {
    const interval = setInterval(() => {
      setFakeMs((prev) => {
        const next = Math.max(0, prev - 50);
        if (next === 0) clearInterval(interval);
        return next;
      });
    }, 50);
    return () => clearInterval(interval);
  }, []);

  // const remainingMs = _remainingMs; // restore this and remove fakeMs to use real data
  const remainingMs = fakeMs;

  const seconds = Math.floor(remainingMs / 1000);
  const ms = Math.floor((remainingMs % 1000) / 10);

  return (
    <motion.div key="get-ready" initial={{ opacity: 0, scale: 0.95 }} animate={{ opacity: 1, scale: 1 }} exit={{ opacity: 0, scale: 0.95 }} className="flex flex-col items-center justify-center gap-4 pt-20">
      <p className="text-4xl font-bold font-display text-game-purple animate-pulse">Gör dig redo...</p>
      <p className="text-2xl font-bold font-display text-muted-foreground tabular-nums">
        {seconds}.{String(ms).padStart(2, "0")}
      </p>
    </motion.div>
  );
}
