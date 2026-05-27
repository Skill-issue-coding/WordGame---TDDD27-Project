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

export function InputPhase() {
  const { sendEvent } = useWebsocketContext();
  const game = useImpostorGame();
  const { user } = useUserContext();
  const [wordSubmission, setWordSubmission] = useState<string>("");

  if (!game || !game.roundState || !game.phaseState || !game.phaseState.current_player || !user) return null;

  const isImpostor = game.roundState.role === "impostor";

  const sendWordSubmission = () => {
    if (isStringEmptyOrOnlySpaces(wordSubmission) || wordSubmission.length > 128) {
      ToastError("Skriv in ett ord");
      return;
    }

    if (game.phaseState?.current_player !== user.user_id) {
      ToastError("Det är inte din tur!");
      return;
    }

    sendEvent("game_submit_word", { word: wordSubmission });
  };

  return (
    <PhaseTransition phaseKey="input">
      <div className="w-full max-w-md flex flex-col items-center">
        <CountdownBar />
        <div className="mt-6 text-center mb-6">
          <p className={cn("text-muted-foreground text-sm mb-2 font-display uppercase tracking-wider font-bold", isImpostor ? "text-destructive" : "text-muted-foreground")}>{isImpostor ? "Hitta på en bluff" : "Ange en ledtråd"}</p>
          <p className="text-xs text-muted-foreground font-display">{isImpostor ? "Välj ett ord som får dig att smälta in i gruppen" : "Skriv ett ord relaterat till ditt hemliga ord"}</p>
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
    </PhaseTransition>
  );
}
