"use client";

/**
 * @file lobbycontext.tsx
 * Global React context that owns lobby state and chat history.
 *
 * Architecture:
 * - Subscribes to the websocket transport for lobby-specific events.
 * - Keeps the latest lobby code in a ref for navigation on join.
 * - Updates lobby state and chat messages in response to server syncs.
 *
 * Usage:
 * ```tsx
 * const { lobbyState, chatMessages } = useLobbyContext();
 * ```
 */

import { ChatMessage, LobbyState } from "@/lib/game/types";
import { ToastSucess } from "@/lib/toast-functions";
import { WSSendEventType, WSSendPayloadMap } from "@/lib/websocket/types";
import { useRouter } from "next/navigation";
import { createContext, ReactNode, useContext, useEffect, useState, useRef } from "react";
import { useWebsocketContext } from "./websocketcontext";

/**
 * Typed sendMessage function. The generic parameter T constrains the payload
 * shape to the one defined in WSSendPayloadMap for the given event type,
 * preventing mismatched event/payload combinations at compile time.
 */
export type SendMessageType = <T extends WSSendEventType>(type: T, payload: WSSendPayloadMap[T]) => void;

/** Shape of the value exposed by GameContext. */
export interface LobbyContextProps {
  /**
   * The full shared lobby state, updated on every sync_gamestate event.
   * Null when the player is not in a lobby.
   */
  lobbyState: LobbyState | null;

  /** Chatmessages sent internaly in the lobby */
  chatMessages: ChatMessage[];
}

export const LobbyContext = createContext<LobbyContextProps | null>(null);

/**
 * useLobbyContext returns the LobbyContext value and throws if called outside
 * of a LobbyContextProvider tree.
 */
/**
 * Access the current lobby state and chat messages.
 *
 * Usage:
 * ```tsx
 * const { lobbyState, chatMessages } = useLobbyContext();
 * ```
 */
export function useLobbyContext() {
  const context = useContext(LobbyContext);
  if (!context) throw new Error("useLobbyContext must be used within a LobbyContextProvider");
  return context;
}

/**
 * GameContextProvider opens a WebSocket connection on mount and provides
 * the GameContext to its children. Should be placed high in the component
 * tree (e.g. in the root layout) so all routes share the same connection.
 */
/**
 * Provides the lobby context to all child components.
 *
 * Usage:
 * ```tsx
 * <LobbyContextProvider>
 *   <App />
 * </LobbyContextProvider>
 * ```
 */
export function LobbyContextProvider({ children }: { children: ReactNode }) {
  const { subscribe } = useWebsocketContext();

  // Use refs for the websocket and lobby code to avoid stale closures and unnecessary re-renders
  const lobbyCodeRef = useRef<string>("");

  const [lobbyState, setLobbyState] = useState<LobbyState | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);

  const router = useRouter();

  useEffect(() => {
    const unsubJoinError = subscribe("join_error", () => router.push("/"));

    const unsubLeftLobby = subscribe("left_lobby", () => {
      setChatMessages([]);
      setLobbyState(null);
      lobbyCodeRef.current = "";
    });

    const unsubChat = subscribe("chat_message", (payload) => setChatMessages((prev) => [...prev, payload]));

    const unsubSync = subscribe("sync_gamestate", (payload) => {
      lobbyCodeRef.current = payload.lobbystate.code;
      setLobbyState(payload.lobbystate);
      if (payload.message) ToastSucess(payload.message);
    });

    const unsubJoined = subscribe("joined_lobby", () => {
      if (lobbyCodeRef.current) router.push(`/lobby/${lobbyCodeRef.current}`);
    });

    return () => {
      unsubJoinError();
      unsubLeftLobby();
      unsubChat();
      unsubSync();
      unsubJoined();
    };
  }, [router, subscribe]);

  const value: LobbyContextProps = { lobbyState, chatMessages };

  return <LobbyContext.Provider value={value}>{children}</LobbyContext.Provider>;
}
