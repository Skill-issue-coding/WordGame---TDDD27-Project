# Game Package — Architecture & Protocol

This document covers the complete message flow from the frontend WebSocket connection through
the server's goroutine topology, down to the active game and back.

---

## Goroutine & Channel Topology

Every layer owns its state exclusively through a single goroutine. Goroutines communicate
only via typed channels — no shared-memory locks except the hub's `LobbiesMutex` (used
only by HTTP upgrade handlers that run outside the hub goroutine).

```mermaid
graph TD
    Browser["Browser\n(WebSocket)"]

    subgraph "per-client goroutines"
        RP["ReadPump\ngoroutine"]
        WP["WritePump\ngoroutine"]
    end

    subgraph "Hub goroutine"
        HUB["GameHub.Run\n(owns Clients map)"]
    end

    subgraph "Lobby goroutine (per room)"
        LOBBY["GameLobby.Run\n(owns all lobby state)"]
    end

    subgraph "Game goroutine (per active game)"
        GAME["Game.Run\n(owns all game state)"]
    end

    Browser -- "raw WS frames" --> RP
    RP -- "Hub.Register / Unregister" --> HUB
    RP -- "Lobby.Register / Unregister\nModeUpdateRequests\nSettingUpdateRequests\nStartGameRequests\nChatMessages\nSyncRequests" --> LOBBY
    RP -- "Lobby.GameInputs\n(GameInput{ClientId, Event})" --> LOBBY
    LOBBY -- "GameInputs → Game.HandleInput" --> GAME
    GAME -- "GameOutputs\n(GameOutput{Target, Type, Payload})" --> LOBBY
    LOBBY -- "client.Send channel" --> WP
    HUB -- "client.Send channel" --> WP
    WP -- "raw WS frames" --> Browser
```

A `nil` `GameOutput.Target` means broadcast to all clients in the lobby.  
A non-nil `Target` means send privately to that one player.

---

## Pre-game: Connection & Lobby Events

### Client → Server

| Event | Payload | Auth | Description |
|---|---|---|---|
| `create_lobby` | — | any | Creates a new room; sender becomes host |
| `join_lobby` | `{lobby_code: string}` | any | Joins existing room by code |
| `leave_lobby` | — | any | Leaves current room |
| `update_user` | `{username?, background?}` | any | Updates display name / color |
| `change_mode` | `{mode: GameMode}` | host only | Switches game mode |
| `update_setting` | `{key: GameSetting, value: float64}` | host only | Updates one setting |
| `send_chatmessage` | `{message: string}` | any | Broadcasts a chat message |
| `start_game` | — | host only | Starts the game |

### Server → Client

| Event | Target | Payload | Trigger |
|---|---|---|---|
| `connected_to_hub` | private | `{user: UserProfile}` | On WS connection |
| `joined_lobby` | private | — | After successful join/create |
| `left_lobby` | private | — | After leaving |
| `sync_gamestate` | broadcast | `{lobbystate: LobbyState, message?}` | Any shared state change |
| `error` | private | `{message: string}` | Validation failure |
| `success` | private | `{message: string}` | Positive acknowledgment |
| `chat_message` | broadcast | `{sender, message, date}` | Chat message received |
| `game_started` | broadcast | — | Host triggers start |

`LobbyState` contains `{code, mode, phase, host, users, settings}`.  
`phase` is either `"lobby"` or `"game_started"`.

---

## Impostor Game Flow

### Phase State Machine

```mermaid
stateDiagram-v2
    direction LR
    [*] --> show_word : game starts\n(8 s + 2 s sync)
    show_word --> input : timer expires\n(input_duration s)
    input --> discussion : timer expires\n(discussion_duration s)
    discussion --> vote : timer expires\n(vote_duration s)
    vote --> result : timer expires
    result --> show_word : game not over\ncycleNumber++
    result --> [*] : game over
```

Each phase transition broadcasts `new_game_phase` with `{start_time, end_time}` (Unix ms)
so clients can render countdown timers.

### Sequence Diagram (single cycle)

```mermaid
sequenceDiagram
    participant C as Client
    participant L as Lobby.Run
    participant G as ImpostorGame.Run

    Note over G: pickImpostorPair()<br/>pickImpostors()

    G->>L: impostor_new_round (private per player)<br/>{role, word, previous_submitted_words}
    G->>L: new_game_phase (broadcast)<br/>{start_time, end_time}
    L->>C: impostor_new_round
    L->>C: new_game_phase

    Note over C,G: ── PhaseShowWord (8 s) ──

    Note over C,G: ── PhaseInput (input_duration s) ──

    C->>L: game_submit_word {word}
    L->>G: GameInput{ClientId, event}
    Note over G: stores submissions[clientId][cycleNumber]

    Note over C,G: (player may resubmit — last write wins)

    G->>L: impostor_discussion_started (broadcast)<br/>{submissions: map[playerId → word]}
    G->>L: new_game_phase (broadcast)
    L->>C: impostor_discussion_started
    L->>C: new_game_phase

    Note over C,G: ── PhaseDiscussion (discussion_duration s) ──

    G->>L: impostor_vote_started (broadcast)<br/>{candidates: [playerId, ...]}
    G->>L: new_game_phase (broadcast)
    L->>C: impostor_vote_started
    L->>C: new_game_phase

    Note over C,G: ── PhaseVote (vote_duration s) ──

    C->>L: game_submit_vote {target: playerId | null}
    L->>G: GameInput{ClientId, event}
    Note over G: stores votes[clientId][cycleNumber]

    Note over C,G: (player may revote — last write wins)

    G->>L: impostor_round_result (broadcast)<br/>{eliminated?, was_impostor, impostors,<br/> vote_results, normal_word, impostor_word,<br/> game_over, impostors_win?}
    L->>C: impostor_round_result

    alt game_over == false
        Note over G: cycleNumber++ → back to PhaseShowWord
    else game_over == true
        Note over G: onDone() → GameDone channel
        L->>C: sync_gamestate (phase: "lobby")
    end
```

### Vote Data Model

Internally the game stores:

```
votes: map[playerId][cycleNumber] → *uuid.UUID
```

A `nil` pointer means the player cast a skip vote for that cycle.

The server tally converts this into the wire format sent in `impostor_round_result`:

```
vote_results: map[candidateId] → []voterId
```

Each key is a player who received votes; the value is the list of players who voted for them.
Empty slices are included for candidates with zero votes.

---

## Settings Reference (Impostor)

| Key | Default | Min | Max |
|---|---|---|---|
| `input_duration` | 30 s | 10 s | 60 s |
| `discussion_duration` | 15 s | 10 s | 60 s |
| `impostor_count` | 1 | 1 | 4 |
| `vote_duration` | 30 s | 10 s | 60 s |

All durations are in seconds. A `SYNC_DELAY` of 2 s is added server-side to every phase
end-time to compensate for network latency before the next phase event arrives.

---

## Known Gaps / TODOs

| # | Location | Description |
|---|---|---|
| 1 | `lobby.go:151` | `ModeContextoBattle` and `ModeSynonymDuel` fall through to "Spelläget stöds inte än" — games not wired up yet. |
| 2 | `impostor.go:404` | `previousRoundSubmissions` only includes players who submitted at least once; players who skipped input are absent. Frontend must handle missing keys. |
| 3 | `impostor.go` | No per-player vote acknowledgment — after casting a vote the voter receives no confirmation event. |
