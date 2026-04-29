"use client";

import { createContext, useContext, useEffect, useState, ReactNode } from "react";

const PALETTE = ["#8b5cf6", "#ec4899", "#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#06b6d4", "#a855f7"];

interface UserProfile {
  username: string;
  color: string;
}

interface UserContextValue extends UserProfile {
  setUsername: (n: string) => void;
  setColor: (c: string) => void;
  palette: string[];
}

const UserContext = createContext<UserContextValue | undefined>(undefined);

const randomName = () => {
  const adjectives = ["Swift", "Clever", "Brave", "Witty", "Sneaky", "Lucky", "Bold", "Cosmic"];
  const nouns = ["Fox", "Owl", "Tiger", "Panda", "Wolf", "Hawk", "Otter", "Lynx"];
  return `${adjectives[Math.floor(Math.random() * adjectives.length)]}${nouns[Math.floor(Math.random() * nouns.length)]}`;
};

export const UserProvider = ({ children }: { children: ReactNode }) => {
  const [profile, setProfile] = useState<UserProfile>({
    username: "Player",
    color: PALETTE[0],
  });

  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem("profile");
    if (stored) {
      try {
        setProfile(JSON.parse(stored));
      } catch (e) {
        setProfile({ username: randomName(), color: PALETTE[0] });
      }
    } else {
      setProfile({
        username: randomName(),
        color: PALETTE[Math.floor(Math.random() * PALETTE.length)],
      });
    }
    setMounted(true);
  }, []);

  return (
    <UserContext.Provider
      value={{
        ...profile,
        setUsername: (username) => setProfile((p) => ({ ...p, username })),
        setColor: (color) => setProfile((p) => ({ ...p, color })),
        palette: PALETTE,
      }}>
      {children}
    </UserContext.Provider>
  );
};

export const useUser = () => {
  const ctx = useContext(UserContext);
  if (!ctx) throw new Error("useUser must be used within UserProvider");
  return ctx;
};
