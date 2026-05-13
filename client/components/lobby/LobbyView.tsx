"use client";

import { ArrowLeft, Play, BookOpenText, Loader2 } from "lucide-react";
import Link from "next/link";
import { PlayerList } from "@/components/lobby/PlayerList";
import { SettingsPanel } from "@/components/lobby/GameSettings";
import { Button } from "@/components/ui/button";
import { motion } from "framer-motion";
import { snapIn } from "@/lib/animation-util";
import { useState, useEffect } from "react";
import { useUserContext } from "@/hooks/usercontext";
import { useLobbyContext } from "@/hooks/lobbycontext";
import { useWebsocketContext } from "@/hooks/websocketcontext";

export default function LobbyView({ code }: { code: string }) {
  const { user } = useUserContext();
  const { lobbyState } = useLobbyContext();
  const { sendEvent, connectionStatus } = useWebsocketContext();
  const [hasAttemptedJoin, setHasAttemptedJoin] = useState(false);

  useEffect(() => {
    if (!code || typeof code !== "string") return;

    if (connectionStatus === "connected" && user && !lobbyState && !hasAttemptedJoin) {
      setHasAttemptedJoin(true);
      sendEvent("join_lobby", { lobby_code: code });
    }
  }, [connectionStatus, user, lobbyState, code, hasAttemptedJoin, sendEvent]);

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
        <motion.div className="relative flex items-center justify-between mb-6" {...snapIn({ delay: 0.08, strength: 1.2, y: 10 })}>
          <div className="flex-1 flex justify-start">
            <Link href="/" className="flex items-center" onClick={handleLeave}>
              <button className="flex items-center gap-2 transition-colors cursor-pointer text-muted-foreground hover:text-foreground">
                <ArrowLeft className="w-4 h-4" />
                <span className="text-sm">Tillbaka</span>
              </button>
            </Link>
          </div>

          <h1 className="text-4xl font-bold font-display text-game-purple whitespace-nowrap">{hostName?.slice(-1) === "s" ? `${hostName} rum` : `${hostName}s rum`}</h1>
          <div className="flex-1" />
        </motion.div>
        <div>
          <div className="flex flex-col sm:grid sm:grid-cols-5 gap-6">
            <motion.div className="sm:col-span-3" {...snapIn({ delay: 0.16, x: -12, y: 12 })}>
              <SettingsPanel className={"sm:col-span-3"} />
            </motion.div>
            <motion.div className="sm:col-span-2" {...snapIn({ delay: 0.2, x: 12, y: 12 })}>
              <PlayerList className={"sm:col-span-2"} />
            </motion.div>
          </div>
          <motion.div className="flex flex-col sm:flex-row mt-6 gap-6" {...snapIn({ delay: 0.24, y: 14, rotate: 1.5 })}>
            <Button variant="glass" size="lg" className="gap-2 flex-1 min-h-12 font-body">
              Snabb Guide
              <BookOpenText />
            </Button>

            <Button size="lg" className="gap-2 flex-1 min-h-12 font-body">
              Starta
              <Play />
            </Button>
          </motion.div>
        </div>
      </div>
    </div>
  );
}
