"use client";

import PhaseTransition from "@/components/game/PhaseTransition";
// import { useImpostorGame } from "@/hooks/gamecontext";
// import { useLobbyContext } from "@/hooks/lobbycontext";
import { cn } from "@/lib/utils";
// import { useUserContext } from "@/hooks/usercontext";
import { Skull } from "lucide-react";
import { fakeUsers, fakeVoteResult } from "@/lib/fakedata";

export function IntermediatePhase() {
  // const game = useImpostorGame();
  // const { users } = useLobbyContext();
  // const { user } = useUserContext();

  // if (!game || !game.voteResult || !user || !users) return null;
  // const voteResult = game.voteResult;

  const voteResult = fakeVoteResult;
  const users = fakeUsers;
  const votedOut = voteResult.voted_out;
  const message = voteResult.message ? voteResult.message : votedOut ? `${users[votedOut].username} röstades ut` : "Ingen röstades ut";
  const isEliminated = votedOut !== undefined;

  return (
    <PhaseTransition phaseKey="Intermediate">
      <div className="flex flex-col items-center w-full max-w-3xl gap-8">
        <div className="text-center">
          <h1 className={cn("text-5xl font-display font-bold mb-2")}>{message}</h1>
        </div>

        {isEliminated && votedOut && (
          <div className="w-full game-card">
            <Skull className="mb-4 size-12" />
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
              <div key={votedOut} className="flex flex-col items-center p-3 rounded-lg gap-2 bg-background">
                <span className="flex items-center justify-center w-12 h-12 text-xl font-bold text-white rounded-full shrink-0 font-display" style={{ backgroundColor: users[votedOut].background }}>
                  {users[votedOut].username[0]}
                </span>
                <span className="w-full font-bold text-center truncate font-display">{users[votedOut].username}</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </PhaseTransition>
  );
}
