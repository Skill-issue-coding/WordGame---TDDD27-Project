"use client";

import PhaseTransition from "@/components/game/PhaseTransition";
import CountdownBar from "@/components/game/CountdownBar";
import { cn } from "@/lib/utils";
import { useUserContext } from "@/hooks/usercontext";
import { useWebsocketContext } from "@/hooks/websocketcontext";
import { useImpostorGame } from "@/hooks/gamecontext";
import { useLobbyContext } from "@/hooks/lobbycontext";
import { useMemo } from "react";

export function VotePhase() {
  const { user } = useUserContext();
  const { users } = useLobbyContext();
  const { sendEvent } = useWebsocketContext();
  const game = useImpostorGame();

  if (!game || !game.phaseState || !game.roundState || !users || !user) return null;

  const submittedVotes = game.phaseState.votes_cycle_votes;
  const activePlayers = game.roundState.active_players;
  const isCurrentUserActive = user ? !!activePlayers[user.user_id] : false;
  const myVote = submittedVotes[user.user_id];

  const handleVote = (target: string | null) => {
    if (target === myVote || !isCurrentUserActive) return;
    sendEvent("game_submit_vote", { target: target });
  };

  const { votersByTarget, skipVoters, leader } = useMemo(() => {
    const newVotersByTarget: Record<string, string[]> = {};
    const newSkipVoters: string[] = [];
    const counts: Record<string, number> = {};
    let skipCount = 0;

    for (const [voterId, target] of Object.entries(submittedVotes)) {
      if (target === null) {
        newSkipVoters.push(voterId);
        skipCount++;
      } else {
        (newVotersByTarget[target] ??= []).push(voterId);
        counts[target] = (counts[target] ?? 0) + 1;
      }
    }

    // Calculate leader
    let maxVotes = 0;
    let topCandidates: string[] = [];
    for (const [id, n] of Object.entries(counts)) {
      if (n > maxVotes) {
        maxVotes = n;
        topCandidates = [id];
      } else if (n === maxVotes) {
        topCandidates.push(id);
      }
    }

    let finalLeader: string | null | undefined = undefined; // Default: tie or nobody
    if (maxVotes === 0 && skipCount === 0) finalLeader = undefined;
    else if (skipCount > maxVotes) finalLeader = null;
    else if (skipCount < maxVotes && topCandidates.length === 1) finalLeader = topCandidates[0];

    return {
      votersByTarget: newVotersByTarget,
      skipVoters: newSkipVoters,
      leader: finalLeader,
    };
  }, [submittedVotes]);

  return (
    <PhaseTransition phaseKey="vote">
      <div className="w-full max-w-2xl">
        <CountdownBar />
        <div className="mt-6 text-center mb-6">
          <h2 className="font-display text-2xl font-bold text-foreground mb-2">Rösta!</h2>
          <p className="text-muted-foreground text-sm font-display font-semibold">Vem är en imposter?</p>
        </div>

        <div className="grid grid-cols-2 gap-3 mb-3">
          {Object.entries(users).map(([playerId, player]) => {
            const isActive = activePlayers[playerId] ?? false;
            const isSelected = myVote === playerId;
            const isCurrentUser = playerId === user.user_id;
            const voters = votersByTarget[playerId] ?? [];
            const isLeading = leader === playerId;

            return (
              <button
                key={playerId}
                disabled={!isActive || isCurrentUser}
                onClick={() => handleVote(playerId)}
                className={cn(
                  "game-card flex items-center gap-3 text-left transition-all",
                  isActive && !isCurrentUser && "cursor-pointer hover:border-muted-foreground",
                  isSelected && !isLeading && "border-game-green bg-game-green/40!",
                  isLeading && "border-game-red bg-game-red/40! animate-pulse",
                  (!isActive || isCurrentUser) && "opacity-40 cursor-not-allowed",
                )}>
                <span
                  className="shrink-0 w-9 h-9 rounded-full flex items-center justify-center text-sm font-display font-bold text-white"
                  style={{ backgroundColor: player.background }}>
                  {player.username[0]}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-display font-semibold text-foreground truncate">{player.username}</p>
                  {!isActive ? (
                    <p className="text-xs font-display text-muted-foreground">Eliminerad</p>
                  ) : voters.length > 0 ? (
                    <div className="flex items-center gap-0.5 mt-1 flex-wrap">
                      {voters.map((voterId) => {
                        const voter = users[voterId];
                        return (
                          <span
                            key={voterId}
                            title={voter?.username}
                            className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-display font-bold text-white border border-card"
                            style={{ backgroundColor: voter?.background }}>
                            {voter?.username[0]}
                          </span>
                        );
                      })}
                    </div>
                  ) : null}
                </div>
              </button>
            );
          })}
        </div>

        <button
          onClick={() => handleVote(null)}
          className={cn(
            "game-card w-full flex items-center gap-3 cursor-pointer hover:border-muted-foreground transition-all",
            myVote === null && leader !== null && "border-game-green bg-game-green/40!",
            leader === null && "border-game-red bg-game-red/40! animate-pulse",
          )}>
          <p className="text-sm font-display font-semibold text-foreground">Skippa röst</p>
          {skipVoters.length > 0 && (
            <div className="flex items-center gap-0.5 flex-wrap ml-2">
              {skipVoters.map((voterId) => {
                const voter = users[voterId];
                return (
                  <span
                    key={voterId}
                    title={voter?.username}
                    className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-display font-bold text-white border border-card"
                    style={{ backgroundColor: voter?.background }}>
                    {voter?.username[0]}
                  </span>
                );
              })}
            </div>
          )}
        </button>
      </div>
    </PhaseTransition>
  );
}
