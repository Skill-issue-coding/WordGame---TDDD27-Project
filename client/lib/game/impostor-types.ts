export type ImpostorGamePhase = "show_word" | "input" | "discussion" | "vote" | "intermediate" | "result";

export type ImpostorRoles = "normal" | "impostor";

export type ImpostorCycle = {
  submissions: Record<string, string>;
  votes: Record<string, string | null>;
};

export type GameTimers = {
  start_time: number;
  end_time: number;
};

export type GamePhaseUpdate = {
  timers: GameTimers;
  words_cycle: Record<string, string>;
  votes_cycle_votes: Record<string, string>;
  current_player: string;
  game_phase: ImpostorGamePhase;
};

export type ImpostorClientGameState = {
  timers: GameTimers;
  role: ImpostorRoles;
  word: string;
  active_players: Record<string, boolean>;
};

export type ImpostorGameCycleUpdate = {
  cycles: ImpostorCycle[];
  active_players: Record<string, boolean>;
};

export type ImpostorVoteResult = {
  timers: GameTimers;
  voted_out: string | null;
  message: string;
};
