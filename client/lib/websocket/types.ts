import { User } from "@/lib/game/types";

/* RECIEVED FROM THE BACkEND */
export type WSRecievedEvent =
  | { type: "lobby_created" | "joined_lobby"; payload: { user: User; message: string } }
  | { type: "lobby_updated" | "game_started"; payload: any }
  | { type: "error" | "server_connected" | "left_room"; payload: { message: string } };

export type WSEventType = WSRecievedEvent["type"];

/* SENT TO THE BACKEND */
export type WSSendPayloadMap = {
  create_lobby: { username: string; settings: any };
  join_lobby: { gameCode: string; username: string };
};

export type WSSendEventType = keyof WSSendPayloadMap;

export type WSSendEvent = {
  [K in WSSendEventType]: { type: K; payload: WSSendPayloadMap[K] };
}[WSSendEventType];
