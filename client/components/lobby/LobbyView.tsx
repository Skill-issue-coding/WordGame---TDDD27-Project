"use client";

import { ArrowLeft, Play, BookOpenText, Loader2 } from "lucide-react";
import Link from "next/link";
import { PlayerList } from "@/components/lobby/PlayerList";
import { SettingsPanel } from "@/components/lobby/GameSettings";
import { Button } from "@/components/ui/button";
import { useGameContext } from "@/hooks/gamecontext";
import { useState, useEffect } from "react";

export default function LobbyView({ code }: { code: string }) {
  const { user, lobbyState, sendEvent, isConnected } = useGameContext();
  const [hasAttemptedJoin, setHasAttemptedJoin] = useState(false);

  useEffect(() => {
    if (!code || typeof code !== "string") return;

    if (isConnected && user && !lobbyState && !hasAttemptedJoin) {
      setHasAttemptedJoin(true);
      sendEvent("join_lobby", { lobby_code: code });
    }
  }, [isConnected, user, lobbyState, code, hasAttemptedJoin, sendEvent]);

  if (!user || !lobbyState) {
    return (
      <div className="flex flex-col items-center justify-center min-h-screen p-6">
        <Loader2 className="w-10 h-10 animate-spin text-game-purple mb-4" />
        <p className="font-display font-semibold text-muted-foreground mb-6 flex">
          Ansluter till rummet
          <span className="flex w-6">
            <span className="animate-[loading_1.4s_infinite] ml-0.5">.</span>
            <span className="animate-[loading_1.4s_infinite_0.2s] ml-0.5">.</span>
            <span className="animate-[loading_1.4s_infinite_0.4s] ml-0.5">.</span>
          </span>
        </p>
        <Link href="/">
          <Button variant="outline" className="font-body font-bold">
            Avbryt
          </Button>
        </Link>
      </div>
    );
  }

  const hostUser = lobbyState?.users && lobbyState.host ? lobbyState.users[lobbyState.host] : null;
  const hostName = hostUser?.username;
  const handleLeave = () => sendEvent("leave_lobby", null);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-6">
      <div className="w-full max-w-4xl animate-slide-up">
        <div className="relative flex items-center justify-between mb-6">
          <div className="flex-1 flex justify-start">
            <Link href="/" className="flex items-center" onClick={handleLeave}>
              <button className="flex items-center gap-2 transition-colors cursor-pointer text-muted-foreground hover:text-foreground">
                <ArrowLeft className="w-4 h-4" />
                <span className="text-sm">Tillbaka</span>
              </button>
            </Link>
          </div>

          <h1 className="text-4xl font-bold font-display text-game-purple whitespace-nowrap">
            {hostName?.slice(-1) === "s" ? `${hostName} rum` : `${hostName}s rum`}
          </h1>
          <div className="flex-1" />
        </div>
        <div>
          <div className="flex flex-col sm:grid sm:grid-cols-5 gap-6">
            <SettingsPanel className={"sm:col-span-3"} />
            <PlayerList className={"sm:col-span-2"} />
          </div>
          <div className="flex flex-col sm:flex-row mt-6 gap-6">
            <Button variant="glass" size="lg" className="gap-2 flex-1 min-h-12 font-body">
              Snabb Guide
              <BookOpenText />
            </Button>

            <Button size="lg" className="gap-2 flex-1 min-h-12 font-body">
              Starta
              <Play />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
