"use client";

import { AnimatePresence } from "framer-motion";
import { useImpostorGame, usePhaseCountdown, usePhaseReady } from "@/hooks/gamecontext";
import { GetReadyScreen } from "@/components/game/GetReadyScreen";
import { RevealPhase } from "../impostor/RevealPhase";
import { DiscussionPhase } from "../impostor/DiscussionPhase";
import { InputPhase } from "../impostor/InputPhase";
import { VotePhase } from "../impostor/VotePhase";
import { ResultPhase } from "../impostor/ResultPhase";
import { IntermediatePhase } from "../impostor/IntermediatePhase";
import CountdownBar from "../CountdownBar";

export const MainImpostorView = () => {
  const game = useImpostorGame();

  // All hooks must be called before any conditional return (Rules of Hooks).
  // Both hooks accept undefined and return safe defaults when data isn't available.
  const readyTime = game?.phaseState?.timers?.ready_time ?? game?.roundState?.timers?.ready_time;
  const isReady = usePhaseReady(readyTime);
  const remainingMs = usePhaseCountdown(readyTime);

  if (!game || !game.phaseState) return null;

  if (!isReady) return <GetReadyScreen remainingMs={remainingMs} />;

  const phase = game.phaseState.game_phase;

  return (
    <div className="w-full space-y-6">
      {phase !== "show_word" && <CountdownBar />}
      <AnimatePresence mode="wait">
        {phase === "show_word" && <RevealPhase key="reveal" />}
        {phase === "input" && <InputPhase key="input" />}
        {phase === "discussion" && <DiscussionPhase key="discussion" />}
        {phase === "vote" && <VotePhase key="vote" />}
        {phase === "intermediate" && <IntermediatePhase key="intermediate" />}
        {phase === "result" && <ResultPhase key="result" />}
      </AnimatePresence>
    </div>
  );
};
