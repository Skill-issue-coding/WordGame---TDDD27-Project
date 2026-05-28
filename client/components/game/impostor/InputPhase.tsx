"use client";

import PhaseTransition from "@/components/game/PhaseTransition";
import CountdownBar from "@/components/game/CountdownBar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useImpostorGame } from "@/hooks/gamecontext";
import { cn, isStringEmptyOrOnlySpaces } from "@/lib/utils";
import { useWebsocketContext } from "@/hooks/websocketcontext";
import { useState } from "react";
import { ToastError } from "@/lib/toast-functions";
import { useUserContext } from "@/hooks/usercontext";
import { useLobbyContext } from "@/hooks/lobbycontext";

export function InputPhase() {
  const { sendEvent } = useWebsocketContext();
  const game = useImpostorGame();
  const { user } = useUserContext();
  const { users } = useLobbyContext();
  const [wordSubmission, setWordSubmission] = useState<string>("");

  if (!game || !game.roundState || !game.phaseState || !game.phaseState || !user || !users) return null;

  const isImpostor = game.roundState.role === "impostor";

  const sendWordSubmission = () => {
    if (game.phaseState?.current_player !== user.user_id) {
      ToastError("Det är inte din tur!");
      return;
    }

    if (isStringEmptyOrOnlySpaces(wordSubmission) || wordSubmission.length > 128) {
      ToastError("Skriv in ett ord");
      return;
    }

    sendEvent("game_submit_word", { word: wordSubmission });
  };

  const submittedWords = game.phaseState.words_cycle;
  const activePlayers = game.roundState.active_players;
  const isCurrentPlayer = game.phaseState.current_player === user.user_id;

  return (
    <PhaseTransition phaseKey="input">
      <div className="w-full max-w-md flex flex-col items-center">
        <CountdownBar />
        <div className=" w-full max-w-4xl flex justify-between">
          {isCurrentPlayer && (
            <div className="text-center mb-6">
              <div className="mt-6 text-center mb-6">
                <p
                  className={cn(
                    "text-muted-foreground text-sm mb-2 font-display uppercase tracking-wider font-bold",
                    isImpostor ? "text-destructive" : "text-muted-foreground",
                  )}>
                  {isImpostor ? "Hitta på en bluff" : "Ange en ledtråd"}
                </p>
                <p className="text-xs text-muted-foreground font-display">
                  {isImpostor ? "Välj ett ord som får dig att smälta in i gruppen" : "Skriv ett ord relaterat till ditt hemliga ord"}
                </p>
              </div>
              <div className="flex gap-3 w-full items-center justify-center">
                <Input
                  value={wordSubmission}
                  onChange={(e) => setWordSubmission(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key == "enter") {
                      sendWordSubmission;
                    }
                  }}
                  placeholder={isImpostor ? "Skriv en bluff..." : "Skriv en ledtråd..."}
                  className="bg-card border-2 h-12 text-lg font-display font-bold rounded-2xl"
                  maxLength={128}
                  autoFocus
                />
                <Button size="lg" onClick={sendWordSubmission}>
                  Skicka
                </Button>
              </div>
            </div>
          )}
          <div className="game-card flex-1">
            <h3 className="text-sm font-display font-bold text-muted-foreground uppercase mb-3">Ledtrådar</h3>
            <div className="space-y-3">
              {Object.entries(submittedWords ?? {}).map(([userId, clue]) => {
                const player = users[userId];
                const isActivePlayer = activePlayers[userId];
                return (
                  <div key={userId} className={cn("flex items-center justify-between gap-3", !isActivePlayer && "opacity-40")}>
                    <div className="flex items-center gap-2 min-w-0">
                      <span
                        className="shrink-0 w-7 h-7 rounded-full flex items-center justify-center text-xs font-display font-bold text-white"
                        style={{ backgroundColor: player.background }}>
                        {player.username[0]}
                      </span>
                      <span className="text-sm font-display font-semibold text-muted-foreground truncate">{player.username}</span>
                    </div>
                    {clue ? (
                      <span className="shrink-0 px-3 py-1 rounded-full bg-card border-2 border-border text-sm font-display font-bold text-foreground">
                        {clue}
                      </span>
                    ) : (
                      <span className="shrink-0 px-3 py-1 rounded-full border-2 border-dashed border-border text-sm font-display font-bold text-muted-foreground">
                        —
                      </span>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      </div>
    </PhaseTransition>
  );
}
