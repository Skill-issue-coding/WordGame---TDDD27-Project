"use client";

import PhaseTransition from "@/components/game/PhaseTransition";
import { useImpostorGame } from "@/hooks/gamecontext";
import { useLobbyContext } from "@/hooks/lobbycontext";
import { cn } from "@/lib/utils";
import { useUserContext } from "@/hooks/usercontext";
import { Skull } from "lucide-react";

export function IntermediatePhase() {
  const game = useImpostorGame();
  const { users } = useLobbyContext();
  const { user } = useUserContext();

  if (!game || !game.voteResult || !game.voteResult.voted_out || !user || !users) return null;
  const votedOut = game.voteResult.voted_out;
  const message = game.voteResult.message ? game.voteResult.message : `${users[votedOut].username} röstades ut`;

  return (
    <PhaseTransition phaseKey="Intermediate">
      <div className="w-full max-w-3xl flex flex-col items-center gap-8">
        <div className="text-center">
          <h1 className={cn("text-5xl font-display font-bold mb-2")}>{message}</h1>
        </div>

        {votedOut && (
          <div className="w-full game-card">
            <Skull className="size-12 mb-4" />
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              <div key={votedOut} className="flex flex-col items-center gap-2 p-3 rounded-lg bg-background">
                <span
                  className="shrink-0 w-12 h-12 rounded-full flex items-center justify-center text-xl font-bold text-white"
                  style={{ backgroundColor: users[votedOut].background }}></span>
                <span className="font-display font-bold text-center truncate w-full">{users[votedOut].username}</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </PhaseTransition>
  );
}
