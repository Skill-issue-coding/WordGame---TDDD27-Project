"use client";

import { useState } from "react";
import { AnimatePresence } from "framer-motion";
import { useImpostorGame, usePhaseCountdown, usePhaseReady } from "@/hooks/gamecontext";
import { GetReadyScreen } from "@/components/game/GetReadyScreen";
import CountdownBar from "../CountdownBar";
import { fakePhaseState } from "@/lib/fakedata";
import { RoundResultPhase } from "../antimatch/RoundResultPhase";
import { FinalScorePhase } from "../antimatch/FinalScorePhase";
import { InputPhase } from "../antimatch/InputPhase";

type PhaseType = "input" | "round_result" | "final_score" | "show_word";

export function AntiMatchView() {
  const [phase, setPhase] = useState<PhaseType>("input");
  return (
    <div className="w-full space-y-6">
      {phase !== "show_word" && phase !== "final_score" && phase !== "round_result" && <CountdownBar />}
      <AnimatePresence mode="wait">
        {phase === "input" && <InputPhase />}
        {phase === "round_result" && <RoundResultPhase />}
        {phase === "final_score" && <FinalScorePhase />}
      </AnimatePresence>
    </div>
  );
}
