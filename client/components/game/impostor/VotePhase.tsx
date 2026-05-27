"use client";

import PhaseTransition from "@/components/game/PhaseTransition";
import CountdownBar from "@/components/game/CountdownBar";
import { useState } from "react";
import { cn } from "@/lib/utils";
import type { User } from "@/lib/game/types";

const FAKE_PLAYERS: Record<string, User> = {
  "uuid-1": { user_id: "uuid-1", username: "SnabbÄlg42", background: "#8b5cf6" },
  "uuid-2": { user_id: "uuid-2", username: "GladPingvin7", background: "#ec4899" },
  "uuid-3": { user_id: "uuid-3", username: "KnasigFlamingo19", background: "#3b82f6" },
  "uuid-4": { user_id: "uuid-4", username: "ListigBäver88", background: "#10b981" },
  "uuid-5": { user_id: "uuid-5", username: "MystiskKråka3", background: "#f59e0b" },
  "uuid-6": { user_id: "uuid-6", username: "TrögLama55", background: "#ef4444" },
  "uuid-7": { user_id: "uuid-7", username: "HungrigValross21", background: "#06b6d4" },
  "uuid-8": { user_id: "uuid-8", username: "StoltIgelkott9", background: "#a855f7" },
  "uuid-9": { user_id: "uuid-9", username: "ArgFlamingo67", background: "#8b5cf6" },
  "uuid-10": { user_id: "uuid-10", username: "BlötMård44", background: "#ec4899" },
  "uuid-11": { user_id: "uuid-11", username: "FörvirradAxolotl12", background: "#3b82f6" },
  "uuid-12": { user_id: "uuid-12", username: "ModigTapir88", background: "#10b981" },
};

const FAKE_ACTIVE_PLAYERS: Record<string, boolean> = {
  "uuid-1": true,
  "uuid-2": true,
  "uuid-3": false,
  "uuid-4": true,
  "uuid-5": true,
  "uuid-6": true,
  "uuid-7": false,
  "uuid-8": true,
  "uuid-9": true,
  "uuid-10": false,
  "uuid-11": true,
  "uuid-12": true,
};

// uuid-5 has a clear majority → will pulse red
const FAKE_VOTES: Record<string, string | null> = {
  "uuid-2": "uuid-5",
  "uuid-4": null,
  "uuid-8": null,
  "uuid-6": null,
  "uuid-9": "uuid-1",
  "uuid-11": "uuid-1",
};

// Pretend we are uuid-1 so our own card is disabled
const FAKE_CURRENT_USER_ID = "uuid-1";

export function VotePhase() {
  const players = FAKE_PLAYERS;
  const activePlayers = FAKE_ACTIVE_PLAYERS;
  const currentUserId = FAKE_CURRENT_USER_ID;

  const [liveVotes, setLiveVotes] = useState<Record<string, string | null>>(FAKE_VOTES);
  const [myVote, setMyVote] = useState<string | null | undefined>(liveVotes[currentUserId]);

  const handleVote = (target: string | null) => {
    if (target === myVote) return;
    setMyVote(target);
    setLiveVotes((prev) => ({ ...prev, [currentUserId]: target }));
  };

  // Invert: target UUID → voter UUIDs (excludes skips)
  const votersByTarget = Object.entries(liveVotes).reduce<Record<string, string[]>>((acc, [voterId, target]) => {
    if (target !== null) acc[target] = [...(acc[target] ?? []), voterId];
    return acc;
  }, {});

  // Players who chose to skip
  const skipVoters = Object.entries(liveVotes)
    .filter(([, target]) => target === null)
    .map(([voterId]) => voterId);

  // Mirror the backend's getPlayerWithMostVotes cases.
  // Returns the UUID of the leading player, null if skip is leading, or undefined when nobody pulses (tie / no votes).
  const leader = ((): string | null | undefined => {
    const counts: Record<string, number> = {};
    let skipCount = 0;
    for (const target of Object.values(liveVotes)) {
      if (target === null) skipCount++;
      else counts[target] = (counts[target] ?? 0) + 1;
    }
    let maxVotes = 0;
    let topCandidates: string[] = [];
    for (const [id, n] of Object.entries(counts)) {
      if (n > maxVotes) {
        maxVotes = n;
        topCandidates = [id];
      } else if (n === maxVotes) topCandidates.push(id);
    }
    if (maxVotes === 0 && skipCount === 0) return undefined; // nobody voted
    if (skipCount > maxVotes) return null; // skip wins
    if (skipCount === maxVotes || topCandidates.length > 1) return undefined; // tie
    return topCandidates[0]; // one clear leader
  })();

  return (
    <PhaseTransition phaseKey="vote">
      <div className="w-full max-w-2xl">
        <CountdownBar />
        <div className="mt-6 text-center mb-6">
          <h2 className="font-display text-2xl font-bold text-foreground mb-2">Rösta!</h2>
          <p className="text-muted-foreground text-sm font-display font-semibold">Vem är en imposter?</p>
        </div>

        <div className="grid grid-cols-2 gap-3 mb-3">
          {(Object.entries(players) as [string, User][]).map(([userId, player]) => {
            const isActive = !!activePlayers[userId];
            const isSelected = myVote === userId;
            const isCurrentUser = userId === currentUserId;
            const voters = votersByTarget[userId] ?? [];
            const isLeading = leader === userId;

            return (
              <button
                key={userId}
                disabled={!isActive || isCurrentUser}
                onClick={() => handleVote(userId)}
                className={cn(
                  "game-card flex items-center gap-3 text-left transition-all",
                  isActive && !isCurrentUser && "cursor-pointer hover:border-muted-foreground",
                  isSelected && !isLeading && "border-game-green bg-game-green/40!",
                  isLeading && "border-game-red bg-game-red/40! animate-pulse",
                  (!isActive || isCurrentUser) && "opacity-40 cursor-not-allowed",
                )}>
                <span className="shrink-0 w-9 h-9 rounded-full flex items-center justify-center text-sm font-display font-bold text-white" style={{ backgroundColor: player.background }}>
                  {player.username[0]}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-display font-semibold text-foreground truncate">{player.username}</p>
                  {!isActive ? (
                    <p className="text-xs font-display text-muted-foreground">Eliminerad</p>
                  ) : voters.length > 0 ? (
                    <div className="flex items-center gap-0.5 mt-1 flex-wrap">
                      {voters.map((voterId) => {
                        const voter = players[voterId];
                        return (
                          <span key={voterId} title={voter?.username} className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-display font-bold text-white border border-card" style={{ backgroundColor: voter?.background }}>
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
          className={cn("game-card w-full flex items-center gap-3 cursor-pointer hover:border-muted-foreground transition-all", myVote === null && leader !== null && "border-game-green bg-game-green/40!", leader === null && "border-game-red bg-game-red/40! animate-pulse")}>
          <p className="text-sm font-display font-semibold text-foreground">Skippa röst</p>
          {skipVoters.length > 0 && (
            <div className="flex items-center gap-0.5 flex-wrap ml-2">
              {skipVoters.map((voterId) => {
                const voter = players[voterId];
                return (
                  <span key={voterId} title={voter?.username} className="w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-display font-bold text-white border border-card" style={{ backgroundColor: voter?.background }}>
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
