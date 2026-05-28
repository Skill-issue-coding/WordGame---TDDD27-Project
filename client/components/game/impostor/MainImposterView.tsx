"use client";

import { AnimatePresence, motion } from "framer-motion";
import { useImpostorGame, usePhaseCountdown, usePhaseReady } from "@/hooks/gamecontext";
import { RevealPhase } from "./RevealPhase";
import { DiscussionPhase } from "./DiscussionPhase";
import { InputPhase } from "./InputPhase";
import { VotePhase } from "./VotePhase";

export const MainImpostorView = () => {
  const game = useImpostorGame();

  // All hooks must be called before any conditional return (Rules of Hooks).
  // Both hooks accept undefined and return safe defaults when data isn't available.
  const readyTime = game?.phaseState?.timers?.ready_time ?? game?.roundState?.timers?.ready_time;
  const isReady = usePhaseReady(readyTime);
  const remainingMs = usePhaseCountdown(readyTime);

  if (!game) return null;

  const phase = game.phaseState?.game_phase;

  if (!isReady) {
    const seconds = Math.floor(remainingMs / 1000);
    const ms = Math.floor((remainingMs % 1000) / 10);

    return (
      <motion.div
        key="get-ready"
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.95 }}
        className="flex flex-col items-center justify-center pt-20 gap-4">
        <p className="font-display text-4xl font-bold text-game-purple animate-pulse">Gör dig redo...</p>
        <p className="font-display text-2xl font-bold text-muted-foreground tabular-nums">
          {seconds}.{String(ms).padStart(2, "0")}
        </p>
      </motion.div>
    );
  }

  return (
    <div className="pt-5">
      <AnimatePresence mode="wait">
        {!phase && <RevealPhase key="show_word" />}
        {phase === "input" && <InputPhase key="input" />}
        {phase === "discussion" && <DiscussionPhase key="discussion" />}
        {phase === "vote" && <VotePhase key="vote" />}
        {phase === "intermediate" && <span key="intermediate" />}
        {phase === "result" && <span key="result" />}
      </AnimatePresence>
    </div>
  );
};
