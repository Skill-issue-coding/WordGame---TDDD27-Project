"use client";

import { useLobbyContext } from "@/hooks/lobbycontext";
import { ImpostorView } from "./gamemodes/ImposterGameView";
import { ContextoGameView } from "./gamemodes/ContextoGameView";
import { SynonymDuelView } from "./gamemodes/SynonymDuelView";
import { AntiMatchView } from "./gamemodes/AntiMatchView";

export function GameView() {
  const { lobbyState } = useLobbyContext();
  if (!lobbyState) return <ImpostorView />; // TODO: Change to redirecting to "/"

  if (lobbyState.mode === "impostor") return <ImpostorView />;

  if (lobbyState.mode === "contexto_battle") return <ContextoGameView />;

  if (lobbyState.mode === "synonym_duel") return <SynonymDuelView />;

  if (lobbyState.mode === "anti_match") return <AntiMatchView />;
}
