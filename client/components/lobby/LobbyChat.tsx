"use client";

import { useEffect, useRef, useState } from "react";
import { usePathname } from "next/navigation";
import { MessageCircle, Send } from "lucide-react";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useGameContext } from "@/hooks/gamecontext";

const ALLOWED = ["/lobby", "/game"];

const formatTime = (ts: number) => new Date(ts).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });

const ChatButton = () => {
  const location = usePathname();
  const { chatMessages, sendEvent, user, lobbyState } = useGameContext();
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState("");
  const [lastReadIndex, setLastReadIndex] = useState(0);
  const scrollRef = useRef<HTMLDivElement>(null);

  const visible = ALLOWED.includes(location) && Boolean(lobbyState);

  const unread = Math.max(0, chatMessages.length - lastReadIndex);

  useEffect(() => {
    if (open) {
      setLastReadIndex(chatMessages.length);
      return;
    }
    if (lastReadIndex > chatMessages.length) {
      setLastReadIndex(chatMessages.length);
    }
  }, [open, chatMessages, lastReadIndex]);

  if (!visible) return null;

  const handleSend = () => {
    if (!lobbyState) return;
    if (!draft.trim()) return;
    // sendMessage(draft, user?.username || "", user?.background || "");
    sendEvent("send_chatmessage", { message: draft });
    setDraft("");
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <button aria-label="Open chat" className="fixed bottom-6 right-6 z-50 w-14 h-14 rounded-full group outline-none overflow-visible cursor-pointer">
          <div className="w-full h-full rounded-full flex items-center justify-center text-white border-4 border-card bg-primary transition-all duration-200 ease-out group-hover:scale-110 group-active:scale-95 shadow-[0_4px_0_0_oklch(from_var(--color-primary)_l_c_h/0.5),0_8px_20px_oklch(from_var(--color-game-shadow)_l_c_h/0.2)]">
            <MessageCircle className="w-6 h-6 -scale-x-100" />
            {unread > 0 && <span className="absolute -top-1 -right-1 min-w-5 h-5 px-1 rounded-full bg-game-red text-white text-xs font-display font-bold flex items-center justify-center border-2 border-card">{unread > 9 ? "9+" : unread}</span>}
          </div>
        </button>
      </PopoverTrigger>

      <PopoverContent side="top" align="end" sideOffset={12} className="z-50 w-70 max-w-[calc(100vw-2rem)] p-0 rounded-lg border-2 overflow-hidden flex flex-col h-112 max-h-[calc(100vh-7rem)] shadow-xl" onOpenAutoFocus={(e) => e.preventDefault()}>
        <div className="px-4 py-3 border-b-2 border-border shrink-0 flex items-center gap-2">
          <MessageCircle className="w-5 h-5 text-game-blue" />
          <div className="font-display font-bold text-base">Chattrum</div>
        </div>

        <div ref={scrollRef} className="flex-1 min-h-0 overflow-y-auto px-3 py-2 space-y-3">
          {chatMessages.length === 0 && <p className="text-center text-sm text-muted-foreground font-display py-8">Inga medelanden ännu. Säg hej! 👋</p>}
          {chatMessages.map((m) => {
            const self = m.sender.user_id === user?.user_id;
            return (
              <div key={`${m.sender.user_id}-${m.date}`} className={cn("flex gap-3 items-end", self && "flex-row-reverse")}>
                {/* User Avatar */}
                <div className="w-7 h-7 rounded-full flex items-center justify-center font-display font-bold text-white text-xs border-2 border-card shrink-0" style={{ backgroundColor: m.sender.background }}>
                  {m.sender.username.charAt(0).toUpperCase()}
                </div>
                <div className={cn("max-w-[75%] flex flex-col min-w-0", self ? "items-end" : "items-start")}>
                  <div
                    className={cn(
                      "px-3 py-2 rounded-2xl border-2 font-display font-semibold text-sm w-fit max-w-full wrap-break-word whitespace-pre-wrap",
                      self ? "bg-primary text-primary-foreground border-primary rounded-br-md" : "bg-muted border-border text-foreground rounded-bl-md",
                    )}>
                    {m.message}
                  </div>
                  <div className="text-[10px] font-display font-bold text-muted-foreground px-1 mt-1">
                    {self ? "You" : m.sender.username} · {formatTime(m.date)}
                  </div>
                </div>
              </div>
            );
          })}
        </div>

        <div className="p-3 border-t-2 border-border shrink-0 flex gap-2">
          <Input
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && !e.shiftKey) {
                e.preventDefault();
                handleSend();
              }
            }}
            placeholder="Skriv ett meddelande..."
            maxLength={200}
            className="font-body font-semibold h-10 border-2 rounded-2xl"
          />
          <Button onClick={handleSend} disabled={!draft.trim()} size="icon" className="h-10 w-10 shrink-0" aria-label="Send message">
            <Send className="w-4 h-4" />
          </Button>
        </div>
      </PopoverContent>
    </Popover>
  );
};

export default ChatButton;
