"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Sun, Check, Moon } from "lucide-react";
import { cn } from "@/lib/utils";
import { useUser } from "@/contexts/UserContext";
import { useEffect } from "react";
import { useTheme } from "next-themes";

export default function UserProfileButton() {
  const { username, color, setUsername, setColor, palette } = useUser();
  const [mounted, setMounted] = useState(false);
  const [open, setOpen] = useState(false);
  const [draftName, setDraftName] = useState(username);
  const [draftColor, setDraftColor] = useState(color);
  const { theme, setTheme } = useTheme();

  const handleOpen = (v: boolean) => {
    if (v) {
      setDraftName(username);
      setDraftColor(color);
    }
    setOpen(v);
  };

  const handleSave = () => {
    const trimmed = draftName.trim().slice(0, 16);
    if (trimmed) setUsername(trimmed);
    setColor(draftColor);
    setOpen(false);
  };

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) return null;

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark");
  };

  return (
    <>
      <button
        onClick={() => handleOpen(true)}
        aria-label="Edit profile"
        className="fixed bottom-6 left-6 z-50 w-14 h-14 rounded-full flex items-center justify-center font-display font-bold text-xl text-white border-4 border-card transition-transform hover:scale-110 active:scale-95"
        style={{
          backgroundColor: color,
          boxShadow: `0 4px 0 0 ${color}88, 0 8px 20px hsl(var(--game-shadow) / 0.2)`,
        }}>
        {username.charAt(0).toUpperCase()}
      </button>

      <Dialog open={open} onOpenChange={handleOpen}>
        <DialogContent className="rounded-3xl border-2">
          <DialogHeader>
            <DialogTitle className="font-display text-2xl flex items-center gap-2">
              <span>👤</span> Din Profil
            </DialogTitle>
            <DialogDescription className="font-display font-semibold">
              Välj ett namn och en färg andra spelare får se.
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col items-center gap-4 py-2">
            <div
              className="w-24 h-24 rounded-full flex items-center justify-center font-display font-bold text-4xl text-white border-4 border-card"
              style={{
                backgroundColor: draftColor,
                boxShadow: `0 4px 0 0 ${draftColor}88`,
              }}>
              {(draftName || "?").charAt(0).toUpperCase()}
            </div>

            <div className="w-full">
              <label className="text-xs font-display font-bold text-muted-foreground uppercase tracking-wider mb-2 block">
                Användarnamn
              </label>
              <Input
                value={draftName}
                onChange={(e) => setDraftName(e.target.value)}
                placeholder="Enter a username"
                maxLength={16}
                className="font-display text-lg font-bold h-12 border-2 rounded-2xl text-center"
                autoFocus
              />
              <p className="text-xs text-muted-foreground font-display mt-1 text-right">{draftName.length}/16</p>
            </div>

            <div className="w-full">
              <label className="text-xs font-display font-bold text-muted-foreground uppercase tracking-wider mb-2 block">
                Avatar Färg
              </label>
              <div className="grid grid-cols-8 gap-2">
                {palette.map((c) => (
                  <button
                    key={c}
                    onClick={() => setDraftColor(c)}
                    className={cn(
                      "aspect-square rounded-xl flex items-center justify-center transition-transform hover:scale-110",
                      draftColor === c && "ring-2 ring-offset-2 ring-offset-background ring-foreground scale-110",
                    )}
                    style={{ backgroundColor: c }}
                    aria-label={`Color ${c}`}>
                    {draftColor === c && <Check className="w-4 h-4 text-white" />}
                  </button>
                ))}
              </div>
            </div>

            <div className="w-full">
              <label className="text-xs font-display font-bold text-muted-foreground uppercase tracking-wider mb-2 block">
                Tema
              </label>

              <Button variant="glass" onClick={toggleTheme} className="w-full justify-start gap-3 h-12">
                {theme === "light" ? (
                  <>
                    <Sun className="w-5 h-5 text-game-yellow" />
                    Ljust Läge
                    <span className="ml-auto text-xs text-muted-foreground">Tryck för att byta</span>
                  </>
                ) : (
                  <>
                    <Moon className="w-5 h-5 text-game-blue" />
                    Mörkt Läge
                    <span className="ml-auto text-xs text-muted-foreground">Tryck för att byta</span>
                  </>
                )}
              </Button>
            </div>
          </div>

          <DialogFooter className="gap-2 sm:gap-2">
            <Button variant="glass" onClick={() => setOpen(false)} className="flex-1">
              Avbryt
            </Button>
            <Button onClick={handleSave} disabled={!draftName.trim()} className="flex-1">
              Spara
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
