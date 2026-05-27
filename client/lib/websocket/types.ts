/**
 * @file types.ts (lib/websocket)
 * Master WebSocket type composition.
 *
 * Imports per-domain event unions and merges them into the top-level
 * WSReceivedEvent and WSSendPayloadMap consumed by WebSocketProvider,
 * subscribe(), and sendEvent().
 *
 * Domain breakdown:
 *   lobby.ts         — connection handshake, lobby management, chat
 *   game/impostor.ts — Impostor mode server → client events
 *   game/shared.ts   — round lifecycle events + all client input actions
 */

export type { LobbyWSReceivedEvent, LobbyWSSendPayloadMap } from "./lobby";
export type { ImpostorWSReceivedEvent } from "./game/impostor";
export type { SharedGameWSReceivedEvent, GameWSSendPayloadMap } from "./game/shared";

import type { LobbyWSReceivedEvent, LobbyWSSendPayloadMap } from "./lobby";
import type { ImpostorWSReceivedEvent } from "./game/impostor";
import type { SharedGameWSReceivedEvent, GameWSSendPayloadMap } from "./game/shared";

// ---------------------------------------------------------------------------
// Server → Client master union
// ---------------------------------------------------------------------------

/**
 * Union of every event the Go server can send to the frontend.
 * Discriminate on `type` to narrow to the correct payload shape.
 */
export type WSReceivedEvent = LobbyWSReceivedEvent | ImpostorWSReceivedEvent | SharedGameWSReceivedEvent;

/** All event type strings the server can send. */
export type WSReceivedEventType = WSReceivedEvent["type"];

/** Maps each received event type string to its payload shape. */
export type WSReceivedPayloadMap = {
  [E in WSReceivedEvent as E["type"]]: E["payload"];
};

// ---------------------------------------------------------------------------
// Client → Server master map
// ---------------------------------------------------------------------------

/**
 * Maps every outbound event type string to its payload shape.
 * Extend LobbyWSSendPayloadMap or GameWSSendPayloadMap to add new C→S events.
 */
export type WSSendPayloadMap = LobbyWSSendPayloadMap & GameWSSendPayloadMap;

/** All event type strings the client can send. */
export type WSSendEventType = keyof WSSendPayloadMap;

/**
 * A fully typed outbound WebSocket event object, ready to be JSON.stringify'd
 * and sent over the socket.
 */
export type WSSendEvent = {
  [K in WSSendEventType]: { type: K; payload: WSSendPayloadMap[K] };
}[WSSendEventType];
