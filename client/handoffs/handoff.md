# Impostor UI – Styling Session Handoff

## What was done

### Shared fake-data system (`lib/fakedata.ts`)

A single source of truth for all dev/styling fake data. Exports:

- `fakeUsers` — 12 players with names, IDs, and background colours
- `fakeRoundState` — role, word, active_players (user-4, 8, 12 eliminated)
- `fakePhaseState` — words_cycle, votes_cycle_votes, current_player, game_phase
- `fakeChatMessages` — 14 scripted chat messages
- `MY_ID`, `isImpostor`, `isCurrentPlayer`, `isCurrentUserActive`

To change who "you" are or your role, edit these three constants at the top of the file.

### All impostor phase components have fake data

Every phase component has its real hooks commented out and fake data wired in. The pattern is consistent: real hooks are `// commented` directly above the fake equivalents so restoring is a one-line change per hook.

| Component              | Real hook to restore                                                          |
| ---------------------- | ----------------------------------------------------------------------------- |
| `InputPhase.tsx`       | `useImpostorGame`, `useUserContext`, `useLobbyContext`, `useWebsocketContext` |
| `RevealPhase.tsx`      | `useImpostorGame`                                                             |
| `DiscussionPhase.tsx`  | `useImpostorGame`, `useUserContext`, `useLobbyContext`, `useWebsocketContext` |
| `VotePhase.tsx`        | `useImpostorGame`, `useUserContext`, `useLobbyContext`, `useWebsocketContext` |
| `CountdownBar.tsx`     | `useImpostorGame`                                                             |
| `GameView.tsx`         | `useLobbyContext`, `useRouter`, `useEffect` redirect                          |
| `MainImpostorView.tsx` | `useImpostorGame`                                                             |

### Top-level fake toggles per component

**`MainImpostorView.tsx`**

```ts
const FAKE_SHOW_GET_READY = false; // true → shows 2.5 s get-ready screen before the phase
const FAKE_PHASE = "input" as ImpostorPhase | undefined; // change to any phase
```

**`GameView.tsx`**

```ts
const FAKE_MODE: GameMode = "impostor";
```

**`CountdownBar.tsx`**

```ts
const FAKE_DURATION_S = 15; // countdown length in seconds
```

**`GetReadyScreen.tsx`**

```ts
const FAKE_REMAINING_MS = 2_500; // countdown length
```

**`VotePhase.tsx`** — votes arrive live via `VOTE_SCRIPT`. Hit "↻ Spela upp röstning" to replay, "Nollställ" to clear. `SHOW_VOTE_COUNT` and `SHOW_METER` toggle badge/bar visibility.

### Test page (`app/impostor/test/page.tsx`)

The test page now renders `<GameView />` which renders `<MainImpostorView />` which renders the active phase. Change `FAKE_PHASE` in `MainImpostorView` to switch phases.

### New components

| File                                 | Purpose                                                     |
| ------------------------------------ | ----------------------------------------------------------- |
| `components/game/GetReadyScreen.tsx` | "Gör dig redo..." countdown extracted from MainImpostorView |
| `app/loading.tsx`                    | Global loading page (spinner + animated dots, Swedish)      |
| `app/not-found.tsx`                  | Global 404 page (game-purple 404, Swedish copy, link home)  |

### Layout standardisation

`GameView.tsx` wraps all game modes in `w-full px-8 pt-5`. The per-phase `px-8` instances were removed from `InputPhase` and `DiscussionPhase`.

### Backend fix (Go)

`server/game/impostor.go` — `Run()` now calls `sendGamePhaseUpdate()` immediately after `sendInitialGameState()`. This broadcasts a `new_game_phase` event with `game_phase: "show_word"` so the client's `phaseState` is populated for the reveal phase, consistent with every other phase. Previously the reveal had no `phaseState` entry and clients had to fall back to `roundState.timers`.

---

## Patterns and gotchas discovered

### Module-level `Date.now()` expires before mount

```ts
// BAD — evaluated once at module import time, already in the past by render
const FAKE_END_TIME = Date.now() + 15_000;

// GOOD — evaluated at component mount
const [endTime] = useState(() => Date.now() + 15_000);
```

Affected `CountdownBar` and `GetReadyScreen`; both fixed with `useState` lazy initialiser.

### TypeScript narrows `const` literals, breaking phase comparisons

```ts
// BAD — TypeScript narrows FAKE_PHASE to "input", so phase === "discussion" is always false
const FAKE_PHASE: ImpostorPhase | undefined = "input";

// GOOD — cast prevents literal narrowing, all comparisons stay valid
const FAKE_PHASE = "input" as ImpostorPhase | undefined;
```

### Flex children stretch to cross-axis height by default

The InputPhase input card was growing to match the player list height even with `shrink-0`. The fix is `self-start` (`align-self: flex-start`), which opts the element out of the stretch default.

### `usePhaseReady(undefined)` returns `true` immediately

`GetReadyScreen` was never showing because `MainImpostorView` had `readyTime = undefined`. Fixed with a `FAKE_SHOW_GET_READY` toggle that uses `useState(() => Date.now() + 2_500)` to set a live readyTime at mount.

### `useMemo` with a freshly-created object never memoises

```ts
// BAD — allVotes is a new object every render, so the memo always re-runs
const allVotes = { ...a, ...b };
const tally = useMemo(() => deriveTally(allVotes), [allVotes]);

// GOOD — memoize allVotes itself so the reference is stable
const allVotes = useMemo(() => ({ ...a, ...(b ? b : {}) }), [a, b]);
const tally = useMemo(() => deriveTally(allVotes), [allVotes]);
```

---

## Restoring real data

When ready to reconnect to the backend, for each component:

1. Delete the fake data block (constants + inline objects).
2. Uncomment the real hook calls.
3. Uncomment the early-return null guard.
4. Remove the fake variable assignments (`const users = fakeUsers`, etc.).

For `GameView.tsx` also uncomment `useRouter`, `useEffect`, and the `if (!mode) return null` guard, then delete `FAKE_MODE`.

For `MainImpostorView.tsx` restore `useImpostorGame`, replace `const [fakeReadyTime]` with the real `readyTime` derivation, and remove the `FAKE_SHOW_GET_READY` / `FAKE_PHASE` constants.
