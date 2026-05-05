import { LobbyState, User } from "@/lib/game/types";

/* GO (Backend) -> NEXT.JS (FRONTEND) */
export type WSRecievedEvent =
  | { type: "connected_to_hub"; payload: { user: User } }
  | { type: "joined_lobby"; payload: null }
  | { type: "sync_gamestate"; payload: { lobbystate: LobbyState; message?: string } }
  | { type: "lobby_updated" | "game_started"; payload: { lobbystate: LobbyState } }
  | { type: "error" | "left_room"; payload: { message: string } };

export type WSEventType = WSRecievedEvent["type"];

/* NEXT.JS (FRONTEND) -> GO (Backend) */
export type WSSendPayloadMap = {
  create_lobby: null;
  join_lobby: { gameCode: string };
  update_user: { updates: Partial<User> };
};

export type WSSendEventType = keyof WSSendPayloadMap;

export type WSSendEvent = {
  [K in WSSendEventType]: { type: K; payload: WSSendPayloadMap[K] };
}[WSSendEventType];
