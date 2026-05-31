"use client";

import PhaseTransition from "@/components/game/PhaseTransition";
import { cn, deriveTally } from "@/lib/utils";
// import { useUserContext } from "@/hooks/usercontext";
// import { useWebsocketContext } from "@/hooks/websocketcontext";
// import { useImpostorGame } from "@/hooks/gamecontext";
// import { useLobbyContext } from "@/hooks/lobbycontext";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { fakeUsers, fakeRoundState, MY_ID, isCurrentUserActive } from "@/lib/fakedata";

// --- Fake vote simulation ---
const VOTE_SCRIPT: { t: number; voter: string; target: string | null }[] = [
  { t: 700, voter: "user-2", target: "user-3" },
  { t: 1500, voter: "user-6", target: "user-3" },
  { t: 2300, voter: "user-7", target: null },
  { t: 3100, voter: "user-5", target: "user-3" },
  { t: 3900, voter: "user-11", target: "user-5" },
  { t: 4700, voter: "user-9", target: "user-3" },
  { t: 5500, voter: "user-3", target: "user-5" },
  { t: 6300, voter: "user-10", target: "user-3" },
];

const AVATAR_CAP = 5;

type Votes = Record<string, string | null>;

function VoterStrip({ voters, emptyLabel, center }: { voters: string[]; emptyLabel?: string; center?: boolean }) {
  const shown = voters.slice(0, AVATAR_CAP);
  const extra = voters.length - shown.length;

  return (
    <div className={cn("flex items-center gap-0.5 mt-0.5 flex-wrap min-h-5 w-full", center && "justify-center")}>
      {voters.length === 0 ? (
        emptyLabel ? (
          <span className="text-xs font-display text-muted-foreground leading-5">{emptyLabel}</span>
        ) : null
      ) : (
        <>
          {shown.map((voterId) => {
            const voter = fakeUsers[voterId];
            return (
              <span key={voterId} title={voter?.username} className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-display font-bold text-white border border-card" style={{ backgroundColor: voter?.background }}>
                {voter?.username[0]}
              </span>
            );
          })}
          {extra > 0 && <span className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-display font-bold border border-border bg-muted text-muted-foreground">+{extra}</span>}
        </>
      )}
    </div>
  );
}

export function VotePhase() {
  // const { user } = useUserContext();
  // const { users } = useLobbyContext();
  // const { sendEvent } = useWebsocketContext();
  // const game = useImpostorGame();
  // if (!game || !game.phaseState || !game.roundState || !users || !user) return null;

  const users = fakeUsers;
  const activePlayers = fakeRoundState.active_players;

  const [submittedVotes, setSubmittedVotes] = useState<Votes>({});
  const [myVote, setMyVote] = useState<string | null | undefined>(undefined);
  const timers = useRef<ReturnType<typeof setTimeout>[]>([]);

  const runSimulation = useCallback(() => {
    timers.current.forEach(clearTimeout);
    timers.current = [];
    setSubmittedVotes({});
    VOTE_SCRIPT.forEach(({ t: at, voter, target }) => {
      timers.current.push(setTimeout(() => setSubmittedVotes((prev) => ({ ...prev, [voter]: target })), at));
    });
  }, []);

  useEffect(() => {
    runSimulation();
    return () => timers.current.forEach(clearTimeout);
  }, [runSimulation]);

  const handleVote = (target: string | null) => {
    if (target === myVote || !isCurrentUserActive) return;
    // sendEvent("game_submit_vote", { target: target });
    setMyVote(target);
  };

  const allVotes = useMemo(() => ({ ...submittedVotes, ...(myVote !== undefined ? { [MY_ID]: myVote } : {}) }), [submittedVotes, myVote]);

  const { votersByTarget, skipVoters, counts, skipCount, maxVotes, leader } = useMemo(() => deriveTally(allVotes), [allVotes]);

  const denom = Math.max(maxVotes, skipCount, 1);

  return (
    <PhaseTransition phaseKey="vote">
      <div className="w-full max-w-4xl">
        {/* Header */}
        <div className="mb-6 text-center">
          <h2 className="text-2xl font-bold font-display text-foreground">Rösta</h2>
          <p className="text-sm font-semibold text-muted-foreground font-display">Vem är en imposter?</p>
        </div>

        {/* Player grid */}
        <div className="grid w-full grid-cols-2 gap-3 mb-3">
          {Object.entries(users).map(([playerId, player]) => {
            const isActive = activePlayers[playerId] ?? false;
            const isSelected = myVote === playerId;
            const isCurrentUser = playerId === MY_ID;
            const voters = votersByTarget[playerId] ?? [];
            const isLeading = leader === playerId;
            const share = Math.round(((counts[playerId] ?? 0) / denom) * 100);

            return (
              <button
                key={playerId}
                disabled={!isActive || isCurrentUser}
                onClick={() => handleVote(playerId)}
                className={cn(
                  "game-card relative overflow-hidden flex items-center gap-3 text-left transition-all",
                  isActive && !isCurrentUser && "cursor-pointer hover:border-muted-foreground",
                  isSelected && !isLeading && "border-game-green bg-game-green/40!",
                  isLeading && "border-game-red bg-game-red/40! animate-pulse",
                  (!isActive || isCurrentUser) && "opacity-40 cursor-not-allowed",
                )}>
                <span className="flex items-center justify-center text-sm font-bold text-white rounded-full size-8 shrink-0 font-display" style={{ backgroundColor: player.background }}>
                  {player.username[0]}
                </span>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold truncate font-display text-foreground">{player.username}</p>
                  {!isActive ? <p className="text-xs font-display text-muted-foreground">Eliminerad</p> : <VoterStrip voters={voters} emptyLabel="Inga röster än" />}
                </div>
                {isActive && (
                  <div className="flex flex-col items-center text-center shrink-0 min-w-8">
                    <span className={cn("text-xl font-bold font-display tabular-nums leading-none", voters.length === 0 ? "text-muted-foreground/40" : "text-foreground")}>{voters.length}</span>
                    <span className="text-[10px] font-display text-muted-foreground leading-none mt-0.5">{voters.length === 1 ? "röst" : "röster"}</span>
                  </div>
                )}
                {isActive && <span className="absolute bottom-0 left-0 h-1 transition-all duration-500 bg-primary/40 rounded-b-xl" style={{ width: `${share}%` }} />}
              </button>
            );
          })}
        </div>

        {/* Skip button */}
        <button
          onClick={() => handleVote(null)}
          className={cn(
            "game-card relative overflow-hidden w-full flex items-center gap-3 cursor-pointer hover:border-muted-foreground transition-all mb-6",
            myVote === null && leader !== null && "border-game-green bg-game-green/40!",
            leader === null && "border-game-red bg-game-red/40! animate-pulse",
          )}>
          <div className="flex-1 min-w-0 space-y-1">
            <p className="text-sm font-semibold font-display text-foreground">Skippa röst</p>
            <VoterStrip voters={skipVoters} emptyLabel="Inga röster än" center />
          </div>
          <div className="flex flex-col items-center text-center shrink-0 min-w-8">
            <span className={cn("text-xl font-bold font-display tabular-nums leading-none", skipVoters.length === 0 ? "text-muted-foreground/40" : "text-foreground")}>{skipVoters.length}</span>
            <span className="text-[10px] font-display text-muted-foreground leading-none mt-0.5">{skipVoters.length === 1 ? "röst" : "röster"}</span>
          </div>
          <span className="absolute bottom-0 left-0 h-1 transition-all duration-500 bg-primary/40 rounded-b-xl" style={{ width: `${Math.round((skipCount / denom) * 100)}%` }} />
        </button>
      </div>
    </PhaseTransition>
  );
}
