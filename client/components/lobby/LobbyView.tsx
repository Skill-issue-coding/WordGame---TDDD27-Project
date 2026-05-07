"use client";
import { ArrowLeft, Play, BookOpenText } from "lucide-react";
import Link from "next/link";
import { PlayerList } from "@/components/lobby/PlayerList";
import { SettingsPanel } from "@/components/lobby/GameSettings";
import { Button } from "@/components/ui/button";
import { useGameContext } from "@/hooks/gamecontext";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function LobbyView() {
  const router = useRouter();
  const { user, lobbyState, sendEvent } = useGameContext();

  useEffect(() => {
    if (!user || !lobbyState) router.push("/");
  }, [user, lobbyState, router]);

  if (!user || !lobbyState) return null;

  const hostUser = lobbyState?.users && lobbyState.host ? lobbyState.users[lobbyState.host] : null;
  const hostName = hostUser?.username;
  const handleLeave = () => sendEvent("leave_lobby", null);

  return (
    <div className="flex flex-col items-center justify-center min-h-screen p-6">
      <div className="w-full max-w-4xl animate-slide-up">
        <div className="flex items-center flex-1 mb-6 text-center">
          <Link href="/" className="flex items-center w-full h-full">
            <button className="flex items-center gap-2 transition-colors cursor-pointer text-muted-foreground hover:text-foreground">
              <ArrowLeft className="w-4 h-4" />
              <span className="text-sm">Tillbaka</span>
            </button>
          </Link>
          <h1 className="w-full text-4xl font-bold font-display text-glow-green text-game-purple">{hostName?.slice(-1) === "s" ? `${hostName} rum` : `${hostName}s rum`}</h1>
          <div className="w-full" />
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
