export type GamePhase = "lobby" | "game_started";
export type GameMode = "impostor" | "contexto_battle" | "synonym_duel" | "anti_match";

export type User = {
  user_id: string;
  username: string;
  background: string;
  score: number;
};

export type LobbyState = {
  code: string;
  mode: GameMode;
  phase: GamePhase;
  host: string;
  users: Record<string, User>;
  // Settings any                        `json:"settings"`
};
