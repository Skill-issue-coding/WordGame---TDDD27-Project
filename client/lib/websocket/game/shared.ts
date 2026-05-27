/**
 * @file shared.ts (lib/websocket/game)
 * WebSocket event types shared across all game modes.
 *
 * Server → Client: round lifecycle events (game_round_started, game_round_result,
 *   game_result, new_game_phase).
 * Client → Server: all player input actions (submit word, guess, or vote).
 */

import { GameTimers } from "@/lib/game/impostor-types";

// ---------------------------------------------------------------------------
// Server → Client
// ---------------------------------------------------------------------------

export type SharedGameWSReceivedEvent =
  | {
      /** Broadcast at the start of each round for Anti-Match, Synonym Duel, and Contexto Battle. */
      type: "game_round_started";
      payload: { timers: GameTimers };
    }
  | {
      /** Broadcast when a round ends with per-player scores. Shape varies by mode. */
      type: "game_round_result";
      payload: unknown; // TODO: define per-mode result payload shapes
    }
  | {
      /** Broadcast when the whole game is over (all rounds complete). Shape varies by mode. */
      type: "game_result";
      payload: unknown; // TODO: define per-mode game-over payload shapes
    }
  | {
      /** Phase timer update for modes with timed game phases. */
      type: "new_game_phase";
      payload: GameTimers;
    };

// ---------------------------------------------------------------------------
// Client → Server
// ---------------------------------------------------------------------------

export type GameWSSendPayloadMap = {
  /** Submit a word during the input phase (Impostor, Anti-Match, Synonym Duel). */
  game_submit_word: { word: string };

  /** Submit a guess in Contexto Battle — multiple guesses are allowed per round. */
  game_submit_guess: { word: string };

  /**
   * Cast a vote during the Impostor vote phase.
   * A null target means the player chose to skip (no elimination).
   */
  game_submit_vote: { target: string | null };
};
