"use client";

import { User } from "@/lib/game/types";
import { ToastError } from "@/lib/toast-functions";
import { WSRecievedEvent, WSSendEventType, WSSendPayloadMap } from "@/lib/websocket/types";
import { createContext, ReactNode, useCallback, useContext, useEffect, useState } from "react";

export type SendMessageType = <T extends WSSendEventType>(type: T, payload: WSSendPayloadMap[T]) => void;

export interface GameContextContextProps {
  isConnected: boolean;
  sendMessage: SendMessageType;
  user: User | null;
  connectionError: boolean;
}

export const GameContext = createContext<GameContextContextProps | null>(null);

export function useGameContext() {
  const context = useContext(GameContext);
  if (!context) {
    throw new Error("useGameContext must be used within a GameContextProvider");
  }
  return context;
}

export function GameContextProvider({ children }: { children: ReactNode }) {
  const [websocket, setWebSocket] = useState<WebSocket | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<boolean>(false);
  const [user, setUser] = useState<User | null>(null);

  function reset() {
    setWebSocket(null);
    setIsConnected(false);
    setUser(null);
  }

  useEffect(() => {
    const url = process.env.NEXT_PUBLIC_WS_PATH ? `wss://${process.env.NEXT_PUBLIC_WS_PATH}/ws` : `ws://${process.env.NEXT_PUBLIC_LOCAL_WS_PATH}/ws`;
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

    ws.onclose = () => {
      reset();
    };

    ws.onmessage = (event) => {
      const parsedEvent = JSON.parse(event.data) as WSRecievedEvent;
      const { type, payload } = parsedEvent;

      switch (type) {
        case "lobby_created":
          // Logic for lobby created
          break;

        case "joined_lobby":
          // Logic for joined a lobby
          break;
      }
    };
  }, []);

  const sendMessage: SendMessageType = useCallback(
    (type, payload) => {
      if (!websocket || websocket.readyState !== WebSocket.OPEN) {
        ToastError("Ej ansluten till servern");
        return;
      }
      const event = { type, payload };
      websocket.send(JSON.stringify(event));
    },
    [websocket],
  );

  const value: GameContextContextProps = {
    sendMessage,
    isConnected,
    user,
    connectionError,
  };

  return <GameContext.Provider value={value}>{children}</GameContext.Provider>;
}
