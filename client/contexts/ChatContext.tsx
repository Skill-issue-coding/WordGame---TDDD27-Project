"use client";

import { createContext, useContext, useState, ReactNode, useCallback } from "react";

export interface ChatMessage {
  id: string;
  author: string;
  color: string;
  text: string;
  timestamp: number;
  isSelf?: boolean;
}

interface ChatContextValue {
  messages: ChatMessage[];
  unread: number;
  sendMessage: (text: string, author: string, color: string) => void;
  markRead: () => void;
}

const ChatContext = createContext<ChatContextValue | undefined>(undefined);

const seed: ChatMessage[] = [
  { id: "1", author: "Luna", color: "#ec4899", text: "gl hf everyone! 🎉", timestamp: Date.now() - 60000 },
  { id: "2", author: "Rex", color: "#3b82f6", text: "ready when you are", timestamp: Date.now() - 30000 },
];

export const ChatProvider = ({ children }: { children: ReactNode }) => {
  const [messages, setMessages] = useState<ChatMessage[]>(seed);
  const [unread, setUnread] = useState(0);

  const sendMessage = useCallback((text: string, author: string, color: string) => {
    const trimmed = text.trim();
    if (!trimmed) return;
    setMessages((m) => [
      ...m,
      {
        id: `${Date.now()}-${Math.random()}`,
        author,
        color,
        text: trimmed.slice(0, 200),
        timestamp: Date.now(),
        isSelf: true,
      },
    ]);
  }, []);

  const markRead = useCallback(() => setUnread(0), []);

  return <ChatContext.Provider value={{ messages, unread, sendMessage, markRead }}>{children}</ChatContext.Provider>;
};

export const useChat = () => {
  const ctx = useContext(ChatContext);
  if (!ctx) throw new Error("useChat must be used within ChatProvider");
  return ctx;
};
