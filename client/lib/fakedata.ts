import { ChatMessage, User } from "./game/types";
import { ImpostorClientGameState, ImpostorPhaseUpdate } from "@/lib/game/impostor-types";

export const fakeUsers: Record<string, User> = {
  "user-1": { user_id: "user-1", username: "Emil", background: "#4f46e5" },
  "user-2": { user_id: "user-2", username: "Alice", background: "#16a34a" },
  "user-3": { user_id: "user-3", username: "Bob", background: "#dc2626" },
  "user-4": { user_id: "user-4", username: "Sara", background: "#d97706" },
  "user-5": { user_id: "user-5", username: "Erik", background: "#4f46e5" },
  "user-6": { user_id: "user-6", username: "Anna", background: "#16a34a" },
  "user-7": { user_id: "user-7", username: "Carl", background: "#dc2626" },
  "user-8": { user_id: "user-8", username: "Ida", background: "#d97706" },
  "user-9": { user_id: "user-9", username: "Johan", background: "#4f46e5" },
  "user-10": { user_id: "user-10", username: "Klara", background: "#16a34a" },
  "user-11": { user_id: "user-11", username: "Lars", background: "#dc2626" },
  "user-12": { user_id: "user-12", username: "Maja", background: "#d97706" },
};

export const isImpostor = false;
export const isCurrentPlayer = true;
export const MY_ID = "user-1";

export const fakeRoundState: ImpostorClientGameState = {
  timers: { start_time: Date.now(), ready_time: Date.now(), end_time: Date.now() + 60_000 },
  role: isImpostor ? "impostor" : "normal",
  word: "Sommarsemester",
  active_players: {
    "user-1": true,
    "user-2": true,
    "user-3": true,
    "user-4": false,
    "user-5": true,
    "user-6": true,
    "user-7": true,
    "user-8": false,
    "user-9": true,
    "user-10": true,
    "user-11": true,
    "user-12": false,
  },
};

export const fakePhaseState: ImpostorPhaseUpdate = {
  timers: { start_time: Date.now(), ready_time: Date.now(), end_time: Date.now() + 60_000 },
  words_cycle: {
    "user-1": "",
    "user-2": "Strand",
    "user-3": "Solkräm",
    "user-4": "",
    "user-5": "Grill",
    "user-6": "Semester",
    "user-7": "",
    "user-8": "Björn",
    "user-9": "",
    "user-10": "",
    "user-11": "",
    "user-12": "Kladdigt",
  },
  votes_cycle_votes: {
    "user-2": "user-3",
    "user-5": "user-3",
    "user-6": "user-3",
    "user-9": "user-3",
    "user-10": "user-3",
    "user-7": null,
    "user-11": "user-5",
  },
  current_player: isCurrentPlayer ? MY_ID : "user-2",
  game_phase: "discussion", //! Change this to the state
};

export const isCurrentUserActive = !!fakeRoundState.active_players[MY_ID];

export const fakeChatMessages: ChatMessage[] = [
  { sender: fakeUsers["user-2"], message: "Har någon ett konstigt ord?", date: Date.now() - 120_000 },
  { sender: fakeUsers["user-4"], message: "Jag vet inte riktigt, solkräm känns väl okej?", date: Date.now() - 110_000 },
  { sender: fakeUsers["user-5"], message: "Jag tycker mer att 'Björn' stack ut.", date: Date.now() - 100_000 },
  { sender: fakeUsers["user-6"], message: "Håller med Erik, björn är skumt.", date: Date.now() - 90_000 },
  { sender: fakeUsers["user-7"], message: "Vem var det som skrev björn?", date: Date.now() - 85_000 },
  { sender: fakeUsers["user-8"], message: "Det var ju jag, vad är problemet?", date: Date.now() - 80_000 },
  { sender: fakeUsers["user-9"], message: "Men björn och sommarsemester?", date: Date.now() - 75_000 },
  { sender: fakeUsers["user-10"], message: "Kanske tänkte på en björn som vaknar ur idet?", date: Date.now() - 65_000 },
  { sender: fakeUsers["user-11"], message: "Nja, det är på våren...", date: Date.now() - 55_000 },
  { sender: fakeUsers["user-12"], message: "Ja, det känns absolut som impostorn.", date: Date.now() - 45_000 },
  { sender: fakeUsers["user-3"], message: 'Jag håller med, "Björn" är jättesus.', date: Date.now() - 30_000 },
  { sender: fakeUsers["user-1"], message: "Hmm, då röstar jag på Ida.", date: Date.now() - 20_000 },
  { sender: fakeUsers["user-2"], message: "Jag med.", date: Date.now() - 10_000 },
  { sender: fakeUsers["user-8"], message: "Ni gör ett misstag!!nfewuifnewiufweiufhewiuhfeiuwhfeuihfiuewhfiuewhfwehfuewhfibvjhfd vd vd shjv dfhvdfhjvdfshjvdfhjvdfhjvfhjdvhjfd bdf jbfd. fdh dfh", date: Date.now() - 5_000 },
];
