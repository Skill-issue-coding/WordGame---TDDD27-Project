"use client";

// import { useRouter } from "next/navigation";
// import { useEffect } from "react";
// import { useLobbyContext } from "@/hooks/lobbycontext";
import { GameMode } from "@/lib/game/types";

import { ContextoGameView } from "./gamemodes/ContextoGameView";
import { SynonymDuelView } from "./gamemodes/SynonymDuelView";
import { AntiMatchView } from "./gamemodes/AntiMatchView";
import { MainImpostorView } from "./gamemodes/MainImposterView";

// --- Fake data toggle ---
const FAKE_MODE: GameMode = "impostor";

export function GameView() {
  // const { mode } = useLobbyContext();
  // const router = useRouter();

  // useEffect(() => {
  //   if (!mode) router.push("/");
  // }, [mode, router]);

  // if (!mode) return null;

  const mode = FAKE_MODE;

  return (
    <div className="w-full px-8 pt-5">
      {mode === "impostor" && <MainImpostorView />}
      {mode === "contexto_battle" && <ContextoGameView />}
      {mode === "synonym_duel" && <SynonymDuelView />}
      {mode === "anti_match" && <AntiMatchView />}
    </div>
  );
}
