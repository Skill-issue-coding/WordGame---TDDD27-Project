"use client";

/**
 * @file websocketcontext.tsx
 * Global React context that manages the WebSocket transport layer.
 *
 * Architecture:
 * - This provider acts strictly as the network manager. It handles connecting,
 * reconnecting, and holding the actual `WebSocket` instance.
 * - It implements a Publish/Subscribe (PubSub) pattern. Instead of storing
 * incoming game state directly in React state (which causes massive re-renders),
 * it routes incoming JSON payloads to whatever child contexts are listening.
 * - Child contexts (like UserContext or LobbyContext) use the `subscribe`
 * function to listen for specific events.
 *
 * Usage:
 * ```tsx
 * const { sendEvent, subscribe, connectionStatus } = useWebsocketContext();
 *
 * useEffect(() => {
 * const unsubscribe = subscribe("chat_message", (payload) => {
 * console.log("New message:", payload);
 * });
 * return unsubscribe; // Cleans up when the component unmounts
 * }, [subscribe]);
 * ```
 */

import { ToastError, ToastSucess } from "@/lib/toast-functions";
import { WSRecievedEvent, WSRecievedEventType, WSRecievedPayloadMap, WSSendEventType, WSSendPayloadMap } from "@/lib/websocket/types";
import { createContext, ReactNode, useCallback, useContext, useEffect, useRef, useState } from "react";

/**
 * Typed sendMessage function. The generic parameter T constrains the payload
 * shape to the one defined in WSSendPayloadMap for the given event type,
 * preventing mismatched event/payload combinations at compile time.
 */
export type SendMessageType = <T extends WSSendEventType>(type: T, payload: WSSendPayloadMap[T]) => void;

/**
 * Represents the current state of the WebSocket connection.
 * Useful for displaying loading spinners or offline indicators in the UI.
 */
type ConnectionStatus = "connected" | "disconnected" | "error";

/**
 * Define the shape of our subscriber callbacks.
 * Currently typed as 'any' for the payload, but can be refined later to
 * map specific event strings to specific payload shapes.
 */
type EventCallback<T extends WSRecievedEventType> = (payload: WSRecievedPayloadMap[T]) => void;

/** Shape of the value exposed by WebSocketContext. */
export interface WebsocketContextProps {
  /**
   * Send a typed event to the Go server. Silently no-ops if the socket is
   * not open and shows an error toast instead.
   */
  sendEvent: SendMessageType;

  /**
   * Returns information about the current connection status.
   * - "connected": Socket is open and ready.
   * - "disconnected": Socket is closed (intentionally or dropped).
   * - "error": An error occurred during connection.
   */
  connectionStatus: ConnectionStatus;

  /**
   * Registers a callback function to run whenever a specific WebSocket event
   * is received from the server.
   * @param eventType The string identifier of the event (e.g., "sync_gamestate").
   * @param callback The function to execute with the event payload.
   * @returns A cleanup function to immediately unsubscribe (ideal for useEffect returns).
   */
  subscribe: <T extends WSRecievedEventType>(eventType: T, callback: EventCallback<T>) => () => void;
}

export const WebSocketContext = createContext<WebsocketContextProps | null>(null);

/**
 * useWebsocketContext returns the WebSocketContext value and throws if called outside
 * of a WebSocketProvider tree.
 */
export function useWebsocketContext() {
  const context = useContext(WebSocketContext);
  if (!context) throw new Error("useWebsocketContext must be used within a WebSocketProvider");
  return context;
}

/**
 * WebSocketProvider mounts high in the component tree, opens the connection,
 * and provides the network transport and routing layer to all child contexts.
 */
export function WebSocketProvider({ children }: { children: ReactNode }) {
  /**
   * Holds the actual WebSocket instance. Using a ref ensures that we don't
   * trigger React re-renders just because the socket instance changes.
   */
  const wsRef = useRef<WebSocket | null>(null);

  /** Track connection status primarily for UI feedback. */
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>("disconnected");

  /**
   * PUBSUB STATE: A map where the key is the event type (e.g., "sync_gamestate")
   * and the value is a Set of callback functions.
   * Kept in a ref so adding/removing listeners is silent and fast.
   */
  const subscribersRef = useRef<Map<WSRecievedEventType, Set<EventCallback<WSRecievedEventType>>>>(new Map());

  /**
   * PUBSUB METHOD: Expose this to child contexts so they can listen to events.
   * Wrapped in useCallback so it has a stable signature for child useEffects.
   */
  const subscribe = useCallback(<T extends WSRecievedEventType>(eventType: T, callback: EventCallback<T>) => {
    if (!subscribersRef.current.has(eventType)) subscribersRef.current.set(eventType, new Set());

    subscribersRef.current.get(eventType)!.add(callback as EventCallback<WSRecievedEventType>);

    // Return a cleanup function so useEffects in child contexts can easily unsubscribe
    return () => subscribersRef.current.get(eventType)?.delete(callback as EventCallback<WSRecievedEventType>);
  }, []);

  useEffect(() => {
    /**
     * Initializes the connection and attaches all base event listeners.
     * Wrapped in a function to make future exponential backoff implementations easier.
     */
    const connect = () => {
      const url = process.env.NEXT_PUBLIC_WS_PATH ? `wss://${process.env.NEXT_PUBLIC_WS_PATH}/ws/game` : `ws://localhost:8080/ws/game`;

      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => setConnectionStatus("connected");

      ws.onerror = () => {
        setConnectionStatus("error");
        ToastError("Fel med anslutning till servern");
      };

      ws.onclose = () => {
        setConnectionStatus("disconnected");
        // TODO: Exponential backoff connect() call goes here later
      };

      /**
       * The master message router. Every incoming websocket message hits this first.
       */
      ws.onmessage = (event) => {
        const parsedEvent = JSON.parse(event.data) as WSRecievedEvent;
        const { type, payload } = parsedEvent;

        // 1. Handle global socket-level UI
        if (type === "error") ToastError(payload.message);
        if (type === "success") ToastSucess(payload.message);

        // 2. Route the payload to anyone listening to this specific event type
        const listeners = subscribersRef.current.get(type as WSRecievedEventType);
        if (listeners) listeners.forEach((callback) => callback(payload));
      };
    };

    connect();

    // Cleanup function for React 18 Strict Mode and component unmounts
    return () => {
      if (wsRef.current) wsRef.current.close();
    };
  }, []);

  /**
   * sendEvent serialises and sends a typed event to the server.
   * Memoised on the websocket instance to avoid stale closures in child components.
   */
  const sendEvent: SendMessageType = useCallback((type, payload) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      ToastError("Ej ansluten till servern");
      return;
    }
    wsRef.current.send(JSON.stringify({ type, payload }));
  }, []);

  const value: WebsocketContextProps = {
    connectionStatus,
    sendEvent,
    subscribe,
  };

  return <WebSocketContext.Provider value={value}>{children}</WebSocketContext.Provider>;
}
