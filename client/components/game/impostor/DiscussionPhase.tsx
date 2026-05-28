"use client";

import { ChevronDown, Send } from "lucide-react";
import PhaseTransition from "@/components/game/PhaseTransition";
import { useEffect, useRef, useState } from "react";
import { cn } from "@/lib/utils";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useLobbyContext } from "@/hooks/lobbycontext";
import { useUserContext } from "@/hooks/usercontext";
import { useWebsocketContext } from "@/hooks/websocketcontext";
import { useImpostorGame } from "@/hooks/gamecontext";
import CountdownBar from "../CountdownBar";

export function DiscussionPhase() {
  const { chatMessages, users } = useLobbyContext();
  const game = useImpostorGame();
  const { user } = useUserContext();
  const { sendEvent } = useWebsocketContext();

  const [draft, setDraft] = useState<string>("");

  const scrollRef = useRef<HTMLDivElement>(null);
  const isAtBottomRef = useRef(true);
  const [isAtBottom, setIsAtBottom] = useState(true);
  const [readCount, setReadCount] = useState(0);
  const unreadBelow = Math.max(0, chatMessages.length - readCount);

  // Scroll to bottom on mount.
  useEffect(() => {
    const el = scrollRef.current;
    if (el) el.scrollTop = el.scrollHeight;
    setReadCount(chatMessages.length);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
  // Auto-scroll when a new message arrives and the user is already at the bottom.
  useEffect(() => {
    if (isAtBottomRef.current) {
      const el = scrollRef.current;
      if (el) el.scrollTop = el.scrollHeight;
      setReadCount(chatMessages.length);
    }
  }, [chatMessages.length]);

  const handleScroll = () => {
    const el = scrollRef.current;
    if (!el) return;
    const atBottom = el.scrollTop + el.clientHeight >= el.scrollHeight - 8;
    isAtBottomRef.current = atBottom;
    if (atBottom !== isAtBottom) setIsAtBottom(atBottom);
    if (atBottom) setReadCount(chatMessages.length);
  };

  const scrollToBottom = () => {
    const el = scrollRef.current;
    if (el) el.scrollTo({ top: el.scrollHeight, behavior: "smooth" });
    setReadCount(chatMessages.length);
    isAtBottomRef.current = true;
    setIsAtBottom(true);
  };

  const handleSend = () => {
    if (!draft.trim()) return;
    sendEvent("send_chatmessage", { message: draft });
    setDraft("");
    setIsAtBottom(true);
  };

  if (!game || !game.phaseState || !game.roundState || !users || !user) return null;

  const submittedWords = game.phaseState.words_cycle;
  const activePlayers = game.roundState.active_players;
  const isCurrentUserActive = user ? !!activePlayers[user.user_id] : false;

  return (
    <PhaseTransition phaseKey="discuss">
      <div className="w-full max-w-4xl">
        <CountdownBar />
        <div className="mt-6 text-center mb-6">
          <h2 className="font-display text-2xl font-bold text-foreground mb-2">Diskussions fas</h2>
          <p className="text-muted-foreground text-sm font-display font-semibold">Berätta, vem är misstänksam?</p>
        </div>
        <div className="flex gap-6 w-full justify-between">
          <div className="game-card flex flex-col justify-between flex-3 gap-4">
            <h4 className="text-sm font-display font-bold text-muted-foreground uppercase mb-3">Chatt</h4>
            <div className="flex-1 min-h-0 relative">
              <div
                ref={scrollRef}
                onScroll={handleScroll}
                className="space-y-3 h-full max-h-100 max-w-135 overflow-y-auto px-3 py-2 w-full rounded-lg">
                {chatMessages.length === 0 && (
                  <p className="text-center text-sm text-muted-foreground font-display py-8">
                    Inga medelanden ännu. Säg skriv vem som är misstänsam.
                  </p>
                )}
                {chatMessages.map((msg, i) => (
                  <div key={i} className="flex items-start gap-2 w-full">
                    <span
                      className="shrink-0 w-6 h-6 rounded-full flex items-center justify-center text-xs font-display font-bold text-white mt-0.5"
                      style={{ backgroundColor: msg.sender.background }}>
                      {msg.sender.username[0]}
                    </span>
                    <div className="min-w-0 w-full">
                      <span className="text-xs font-display font-bold mr-1" style={{ color: msg.sender.background }}>
                        {msg.sender.username}
                      </span>
                      <span className="text-sm font-display text-foreground wrap-break-word whitespace-pre-wrap">{msg.message}</span>
                    </div>
                  </div>
                ))}
              </div>
              {!isAtBottom && unreadBelow > 0 && (
                <button
                  onClick={scrollToBottom}
                  className="absolute bottom-2 left-1/2 -translate-x-1/2 whitespace-nowrap flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-primary text-primary-foreground text-xs font-display font-bold shadow-lg border-2 border-primary/50 transition-opacity">
                  <ChevronDown className="w-3 h-3" />
                  {unreadBelow} nya meddelanden
                </button>
              )}
            </div>
            <div className="flex gap-4">
              <Input
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter" && !e.shiftKey) {
                    e.preventDefault();
                    handleSend();
                  }
                }}
                placeholder={isCurrentUserActive ? "Skriv ett meddelande..." : "Du är inte aktiv..."}
                maxLength={200}
                disabled={!isCurrentUserActive}
                className="font-body font-semibold h-10 border-2 rounded-2xl"
              />
              <Button
                onClick={handleSend}
                disabled={!draft.trim() || !isCurrentUserActive}
                size="icon"
                className="h-10 w-10 shrink-0"
                aria-label="Skicka meddelande">
                <Send className="w-4 h-4" />
              </Button>
            </div>
          </div>
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
