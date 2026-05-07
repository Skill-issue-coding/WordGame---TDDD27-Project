"use client";

/**
 * @file GameContext.tsx
 * Global React context that owns the WebSocket connection and exposes game
 * state to the entire component tree.
 *
 * Architecture:
 * - A single WebSocket connection is opened when the provider mounts and
 *   kept alive for the lifetime of the session.
 * - Incoming events are dispatched in `ws.onmessage` and update the
 *   corresponding state slices.
 * - Child components interact with the server exclusively via `sendMessage`.
 *
 * Usage:
 * ```tsx
 * const { user, lobbyState, sendMessage } = useGameContext();
 * sendMessage("join_lobby", { lobby_code: "AbCd-1234" });
 * ```
 */

import { ChatMessage, LobbyState, LocalStorageProfile, User } from "@/lib/game/types";
import { tryCatch } from "@/lib/try-catch";
import { ToastError, ToastSucess } from "@/lib/toast-functions";
import { WSRecievedEvent, WSSendEventType, WSSendPayloadMap } from "@/lib/websocket/types";
import { useRouter } from "next/navigation";
import { createContext, ReactNode, useCallback, useContext, useEffect, useState } from "react";
import axios from "axios";

/**
 * Typed sendMessage function. The generic parameter T constrains the payload
 * shape to the one defined in WSSendPayloadMap for the given event type,
 * preventing mismatched event/payload combinations at compile time.
 */
export type SendMessageType = <T extends WSSendEventType>(type: T, payload: WSSendPayloadMap[T]) => void;

/**
 * The set of background colors available for player avatars.
 * Must be kept in sync with the `palette` slice in the Go util package.
 */
const palette = ["#8b5cf6", "#ec4899", "#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#06b6d4", "#a855f7"];

/** Shape of the value exposed by GameContext. */
export interface GameContextContextProps {
  /** Whether the WebSocket connection is currently open. */
  isConnected: boolean;

  /**
   * Send a typed event to the Go server. Silently no-ops if the socket is
   * not open and shows an error toast instead.
   */
  sendEvent: SendMessageType;

  /**
   * The current player's profile as assigned/confirmed by the server.
   * Null until the connected_to_hub event is received.
   */
  user: User | null;

  /**
   * The full shared lobby state, updated on every sync_gamestate event.
   * Null when the player is not in a lobby.
   */
  lobbyState: LobbyState | null;

  /** True if the WebSocket failed to connect or encountered an error. */
  connectionError: boolean;

  /**
   * Optimistically update the local user profile and emit an update_user
   * event to the server. The server will propagate the change to all lobby
   * members via sync_gamestate.
   */
  updateUser: (updates: Partial<User>) => Promise<void>;

  /** The palette of selectable avatar background colors. */
  palette: string[];

  /** Chatmessages sent internaly in the lobby */
  chatMessages: ChatMessage[];
}

export const GameContext = createContext<GameContextContextProps | null>(null);

/**
 * useGameContext returns the GameContext value and throws if called outside
 * of a GameContextProvider tree.
 */
export function useGameContext() {
  const context = useContext(GameContext);
  if (!context) throw new Error("useGameContext must be used within a GameContextProvider");
  return context;
}

export const GetLocalStorageProfile = (): LocalStorageProfile | undefined => {
  const stored = typeof window !== "undefined" ? localStorage.getItem("profile") : null;
  if (!stored) return undefined;
  try {
    return JSON.parse(stored) as LocalStorageProfile;
  } catch (e) {
    console.error("Could not load profile: ", e);
    return undefined;
  }
};

const SaveLocalStorageProfile = (profile: LocalStorageProfile) => {
  if (typeof window === "undefined") return;
  localStorage.setItem("profile", JSON.stringify(profile));
};

export async function GetNewUsername(userId: string): Promise<string> {
  const { data, error } = await tryCatch(
    axios.post<{ username: string; error?: string }>("http://localhost:8080/game/username", {
      user_id: userId,
    }),
  );

  if (error) throw error;
  if (!data.data.username) throw new Error(data.data.error ?? "Could not fetch username");

  return data.data.username;
}

/**
 * GameContextProvider opens a WebSocket connection on mount and provides
 * the GameContext to its children. Should be placed high in the component
 * tree (e.g. in the root layout) so all routes share the same connection.
 */
export function GameContextProvider({ children }: { children: ReactNode }) {
  const [websocket, setWebSocket] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<boolean>(false);
  const [user, setUser] = useState<User | null>(null);
  const [lobbyState, setLobbyState] = useState<LobbyState | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);

  const router = useRouter();

  /** Resets all session state on connection close or error. */
  function reset() {
    setWebSocket(null);
    setIsConnected(false);
    setUser(null);
    setLobbyState(null);
  }

  console.log(lobbyState);
  console.log(chatMessages);

  useEffect(() => {
    const profile = GetLocalStorageProfile();

    const url = process.env.NEXT_PUBLIC_WS_PATH ? `wss://${process.env.NEXT_PUBLIC_WS_PATH}/ws/game` : `ws://localhost:8080/ws/game`;
    const ws = new WebSocket(url);

    ws.onopen = () => {
      setWebSocket(ws);
      setIsConnected(true);
    };

    ws.onerror = () => {
      setIsConnected(false);
      ToastError("Fel med anslutning till servern");
      setConnectionError(true);
    };

    ws.onclose = () => reset();

    ws.onmessage = (event) => {
      const parsedEvent = JSON.parse(event.data) as WSRecievedEvent;
      const { type, payload } = parsedEvent;

      switch (type) {
        case "error":
          ToastError(payload.message);
          break;

        case "success":
          ToastSucess(payload.message);
          break;

        case "connected_to_hub":
          const serverUser = payload.user;

          if (profile) {
            const mergeUser = {
              ...serverUser,
              username: profile.username ?? payload.user.username,
              background: profile.background ?? payload.user.background,
            };

            setUser(mergeUser);

            ws.send(
              JSON.stringify({
                type: "update_user",
                payload: { username: mergeUser.username, background: mergeUser.background },
              }),
            );
          } else {
            setUser(payload.user);
          }

          ToastSucess("Välkommen till OrdioArena!");
          break;

        case "left_lobby":
          setChatMessages([]);
          setLobbyState(null);
          break;

        case "chat_message":
          setChatMessages((prev) => [...prev, payload]);
          break;

        case "sync_gamestate":
          // Primary state update — replaces lobbyState wholesale.
          setLobbyState(payload.lobbystate);
          if (payload.message) ToastSucess(payload.message);
          break;

        case "joined_lobby":
          // Server confirmation that the lobby registration succeeded.
          // Navigate to the lobby view; state is already set via sync_gamestate.
          router.push("/lobby");
          break;
      }
    };

    // No cleanup / reconnect logic yet — add exponential back-off here if needed.
  }, []);

  /**
   * sendMessage serialises and sends a typed event to the server.
   * Memoised on the websocket instance to avoid stale closures in child components.
   */
  const sendEvent: SendMessageType = useCallback(
    (type, payload) => {
      if (!websocket || websocket.readyState !== WebSocket.OPEN) {
        ToastError("Ej ansluten till servern");
        return;
      }
      websocket.send(JSON.stringify({ type, payload }));
    },
    [websocket],
  );

  /**
   * updateUser applies a partial update to the local user state optimistically
   * and emits an update_user event to the server. The server will re-broadcast
   * the change to all lobby members via sync_gamestate.
   */
  const updateUser = async (updates: Partial<Pick<User, "username" | "background">>) => {
    const currentUser = user;
    if (!currentUser) return;

    let nextUpdates = { ...updates };
    const trimmedUsername = updates.username?.trim();
    if (trimmedUsername === "") {
      try {
        const newUsername = await GetNewUsername(currentUser.user_id);
        nextUpdates = { ...nextUpdates, username: newUsername };
      } catch (error) {
        ToastError("Kunde inte skapa nytt anvandarnamn");
        return;
      }
    }

    sendEvent("update_user", nextUpdates);

    setUser((prev) => {
      if (!prev) return null;

      const existingProfile = GetLocalStorageProfile();

      SaveLocalStorageProfile({
        ...existingProfile,
        username: nextUpdates.username ?? prev.username,
        background: nextUpdates.background ?? prev.background,
      });

      if (lobbyState) {
        setLobbyState((current) => {
          if (!current || !current.users[currentUser.user_id]) return current;
          return {
            ...current,
            users: {
              ...current.users,
              [currentUser.user_id]: {
                ...current.users[currentUser.user_id],
                ...nextUpdates,
              },
            },
          };
        });
      }
      return { ...prev, ...nextUpdates };
    });
  };

  const value: GameContextContextProps = {
    sendEvent,
    isConnected,
    user,
    connectionError,
    updateUser,
    palette,
    lobbyState,
    chatMessages,
  };

  return <GameContext.Provider value={value}>{children}</GameContext.Provider>;
}
