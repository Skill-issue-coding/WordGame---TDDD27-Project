export type GameModeId = "impostor" | "contexto" | "synonym" | "antimatch";
export type GameModeColor = "green" | "pink" | "blue" | "yellow" | "red";

export interface GameMode {
  id: GameModeId;
  title: string;
  description: string;
  icon: string;
  players: string;
  color: GameModeColor;
  textClass: string;
}

export const GAME_MODES: GameMode[] = [
  {
    id: "impostor",
    title: "Hitta Impostern",
    description: "En spelare får ett unikt ord. Försök att tillsammans hitta impostern innan det är försent!",
    icon: "🕵️",
    players: "4-12 spelare",
    color: "red",
    textClass: "text-game-red",
  },
  {
    id: "contexto",
    title: "Kontext Strid",
    description: "Tävla om att hitta det dålda ordet. Semantiska likheter leder dig närmare och närmare det rätta ordet!",
    icon: "🧠",
    players: "2-12 spelare",
    color: "blue",
    textClass: "text-game-blue",
  },
  {
    id: "synonym",
    title: "Synonym Duell",
    description: "Ange den bästa synonymen varje runda. Den som svarar med den sämsta åker ut!",
    icon: "⚔️",
    players: "3-12 spelare",
    color: "green",
    textClass: "text-game-green",
  },
  {
    id: "antimatch",
    title: "Anti-matchning",
    description: "Tänk anorlunda! Skriv en synonym men var försiktig så det inte matchar någon annans, då får båda noll poäng!",
    icon: "🎯",
    players: "3-12 spelare",
    color: "yellow",
    textClass: "text-game-yellow",
  },
];

export const getMode = (id: string): GameMode => GAME_MODES.find((m) => m.id === id) ?? GAME_MODES[0];
