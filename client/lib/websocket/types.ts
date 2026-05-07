/**
 * @file types.ts (lib/websocket)
 * WebSocket event protocol types shared between the Go server and Next.js client.
 *
 * Every message on the wire is a JSON object:
 * ```json
 * { "type": "<EventType>", "payload": { ... } }
 * ```
 *
 * This file defines the discriminated unions for both directions so TypeScript
 * can enforce that the correct payload shape is used for each event type.
 */

import { ChatMessage, GameMode, LobbyState, User } from "@/lib/game/types";

// ---------------------------------------------------------------------------
// Server → Client
// ---------------------------------------------------------------------------

/**
 * Union of all events the Go server can send to the frontend.
 * Use the `type` discriminant to narrow to the correct payload shape.
 *
 * | type               | when it fires                                       |
 * |--------------------|----------------------------------------------------|
 * | connected_to_hub   | Immediately after WebSocket connection is opened   |
 * | joined_lobby       | After the client is registered into a lobby        |
 * | sync_gamestate     | Whenever shared lobby state changes                |
 * | lobby_updated      | (reserved) explicit lobby settings change          |
 * | game_started       | (reserved) game phase transition                   |
 * | error              | Any server-side validation or runtime error        |
 * | success            | Generic acknowledgement                            |
 * | left_room          | Confirmation that the client left a lobby           |
 * | chat_message       | Recieved chat message from the server side         |
 */
export type WSRecievedEvent =
  | {
      type: "connected_to_hub";
      /** The server-generated profile for this connection. */
      payload: { user: User };
    }
  | {
      type: "joined_lobby" | "left_lobby";
      /** No payload — use the preceding sync_gamestate for state. */
      payload: null;
    }
  | {
      type: "sync_gamestate";
      payload: {
        lobbystate: LobbyState;
        /**
         * Optional toast message shown only to the player whose action
         * triggered the sync (e.g. "Du gick med i spelet!").
         */
        message?: string;
      };
    }
  | { type: "chat_message"; payload: ChatMessage }
  | {
      type: "error" | "success";
      payload: { message: string };
    };

/** Union of all event type strings the server can send. */
export type WSEventType = WSRecievedEvent["type"];

// ---------------------------------------------------------------------------
// Client → Server
// ---------------------------------------------------------------------------

/**
 * Maps each outbound event type to its payload shape.
 * Add new client-initiated events here — TypeScript will enforce the
 * correct payload wherever sendMessage is called.
 */
export type WSSendPayloadMap = {
  /** Create a new lobby. The client is automatically made host. */
  create_lobby: null;

  /** Join an existing lobby by its room code. */
  join_lobby: { lobby_code: string };

  /**
   * Update the current player's profile. Only fields that are provided
   * and non-empty are applied on the server.
   */
  update_user: { username?: string; background?: string };

  /** Send a chat message to server */
  send_chatmessage: { message: string };

  /** Switches mode and resets all settings to server defaults */
  change_mode: { mode: GameMode };

  /**
   * Updates a single setting key for the current mode.
   * Value is typed as unknown here — narrow it per-mode in your components.
   */
  update_setting: { key: string; value: number };

  /** Sends a reuqest to the backend to leave the lobby */
  leave_lobby: null;
};

/** Union of all event type strings the client can send. */
export type WSSendEventType = keyof WSSendPayloadMap;

/**
 * A fully typed outbound WebSocket event object, ready to be
 * JSON.stringify'd and sent over the socket.
 */
export type WSSendEvent = {
  [K in WSSendEventType]: { type: K; payload: WSSendPayloadMap[K] };
}[WSSendEventType];
