"use client";

import PhaseTransition from "@/components/game/PhaseTransition";
import { Button } from "@/components/ui/button";
import { useImpostorGame } from "@/hooks/gamecontext";
import { useLobbyContext } from "@/hooks/lobbycontext";
import Link from "next/link";
import { cn } from "@/lib/utils";
import { useUserContext } from "@/hooks/usercontext";

export function ResultPhase() {
  const game = useImpostorGame();
  const { users } = useLobbyContext();
  const { user } = useUserContext();

  if (!game || !game.result || !user || !users) return null;
  const winners = game.result.winners;
  const playerRoles = game.result.roles;
  const words = game.result.words;

  const normalSecretWord = game.result.normal_word;

  const winningRole = winners.length > 0 ? playerRoles[winners[0]] : null;
  const winningTeamText = winningRole === "impostor" ? "Impostors vann!" : "Normala vann!";
  const winningTeamColor = winningRole === "impostor" ? "text-destructive" : "text-game-green";

  return (
    <PhaseTransition phaseKey="result">
      <div className="w-full max-w-3xl flex flex-col items-center gap-8">
        <div className="text-center">
          <h1 className={cn("text-5xl font-display font-bold mb-2", winningTeamColor)}>{winningTeamText}</h1>
          <p className="text-muted-foreground font-display text-lg">
            Det hemliga ordet var: <span className="font-bold text-foreground">{normalSecretWord}</span>
          </p>
        </div>

        <div className="w-full game-card">
          <h2 className="text-xl font-display font-bold text-center mb-4">Vinnare</h2>
          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
            {winners.map((id) => {
              const p = users[id];
              if (!p) return null;
              return (
                <div key={id} className="flex flex-col items-center gap-2 p-3 rounded-lg bg-background">
                  <span
                    className="shrink-0 w-12 h-12 rounded-full flex items-center justify-center text-xl font-bold text-white"
                    style={{ backgroundColor: p.background }}>
                    {p.username[0]}
                  </span>
                  <span className="font-display font-bold text-center truncate w-full">{p.username}</span>
                </div>
              );
            })}
          </div>
        </div>

        <div className="w-full game-card">
          <h2 className="text-xl font-display font-bold text-center mb-4">Resultatöversikt</h2>
          <div className="space-y-3">
            {Object.entries(users).map(([id, player]) => {
              const role = playerRoles[id];
              const playerWord = words[id];

              return (
                <div key={id} className="flex items-center gap-4 p-3 rounded-lg bg-background border">
                  <span
                    className="shrink-0 w-10 h-10 rounded-full flex items-center justify-center text-lg font-bold text-white"
                    style={{ backgroundColor: player.background }}>
                    {player.username[0]}
                  </span>
                  <div className="flex-1 min-w-0">
                    <p className="font-display font-bold truncate">{player.username}</p>
                    <p className={cn("text-xs font-semibold uppercase", role === "impostor" ? "text-destructive" : "text-muted-foreground")}>
                      {role === "impostor" ? "Impostor" : "Normal"}
                    </p>
                  </div>
                  <div className="text-right shrink-0">
                    <p className="font-display font-bold text-foreground">{playerWord || "Inget ord"}</p>
                    <p className="text-xs text-muted-foreground">Deras ord</p>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        <div className="flex justify-center gap-4 w-full max-w-md">
          <Link href="/lobby" className="flex-1">
            <Button size="lg" className="w-full font-display font-bold text-lg h-14">
              Tillbaka till Lobbyn
            </Button>
          </Link>
          <Link href="/" className="flex-1">
            <Button size="lg" variant="outline" className="w-full font-display font-bold text-lg h-14 border-2">
              Statistik
            </Button>
          </Link>
        </div>
      </div>
    </PhaseTransition>
  );
}
