/**
 * @file impostor.ts (lib/websocket/game)
 * WebSocket event types for the Impostor game mode (server → client).
 *
 * Canonical lifecycle events (game_round_started, new_game_phase, game_result)
 * are defined in game_events.go and reused here with Impostor-specific payloads.
 * Use lobbyState.mode === "impostor" to know these payload shapes are active.
 *
 * Client → server inputs are in shared.ts (game_submit_word, game_submit_vote).
 */

import {
  ImpostorClientGameState,
  ImpostorCycleUpdate,
  ImpostorGameResult,
  ImpostorPhaseUpdate,
  ImpostorVoteResult,
  ImpostorVoteUpdate,
} from "@/lib/game/impostor-types";

// ---------------------------------------------------------------------------
// Server → Client
// ---------------------------------------------------------------------------

export type ImpostorWSReceivedEvent =
  | {
      /** Sent privately to each player at game start with their secret word and role. */
      type: "game_round_started";
      payload: ImpostorClientGameState;
    }
  | {
      /** Broadcast when the phase transitions or the current input turn advances. */
      type: "new_game_phase";
      payload: ImpostorPhaseUpdate;
    }
  | {
      /** Broadcast once when the game ends with full results for all players. */
      type: "game_result";
      payload: ImpostorGameResult;
    }
  | {
      /** Broadcast when a vote phase ends, revealing who (if anyone) was eliminated. */
      type: "impostor_vote_result";
      payload: ImpostorVoteResult;
    }
  | {
      /** Broadcast at the start of a new cycle with updated player roster and history. */
      type: "impostor_new_cycle";
      payload: ImpostorCycleUpdate;
    }
  | {
      /** Broadcast immediately each time any player casts a vote (live vote updates). */
      type: "impostor_vote_update";
      payload: ImpostorVoteUpdate;
    };
