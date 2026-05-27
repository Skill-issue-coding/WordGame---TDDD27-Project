"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";

import { useLobbyContext } from "@/hooks/lobbycontext";

import { ContextoGameView } from "./gamemodes/ContextoGameView";
import { SynonymDuelView } from "./gamemodes/SynonymDuelView";
import { AntiMatchView } from "./gamemodes/AntiMatchView";
import { MainImpostorView } from "./impostor/MainImposterView";

export function GameView() {
  const { mode } = useLobbyContext();
  const router = useRouter();

  useEffect(() => {
    if (!mode) router.push("/");
  }, [mode, router]);

  if (!mode) return null;

  if (mode === "impostor") return <MainImpostorView />;

  if (mode === "contexto_battle") return <ContextoGameView />;

  if (mode === "synonym_duel") return <SynonymDuelView />;

  if (mode === "anti_match") return <AntiMatchView />;
}
