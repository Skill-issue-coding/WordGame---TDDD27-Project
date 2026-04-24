# Project Context

**Act as an expert Full-Stack Developer, Game Designer, and NLP Engineer.**

Below is the core context and architecture for a multiplayer word-based party game I am developing. Please keep this context in mind for all future queries in this conversation.

## 1. Core Architecture & Lobby System

The game features a standard multiplayer lobby system where players can create rooms, invite friends, select a game mode, and start matches.

- **State Management:** Real-time synchronization is required for lobbies, timers, player states, and voting mechanisms.
- **Language:** The game and all NLP calculations are strictly in Swedish.

## 2. Tech Stack

Below is the current tech stack

### 1) Preprocessing (Python)

Main script: `preprocessing/main.py`

Pipeline behavior:

- Loads `cc.sv.300.bin` fastText model.
- Reduces vectors to 100 dimensions.
- Uses spaCy `sv_core_news_sm` for POS filtering.
- Builds output CSVs into `server/wordfiles/`.

Generated files expected by backend loader:

- `celebrities_vectors.csv`
- `companies_vectors.csv`
- `kelly_vectors.csv`
- `korp_vectors.csv`
- `maktbarometern_vectors.csv`

Data sources currently used by script logic:

- Kelly XML word list
- Company datasets
- Celebrity dataset
- Maktbarometern datasets
- Korp statistics + stopword filtering

Important note:

Related utility:

- `preprocessing/colly-crawler/` contains a Go-based scraper/formatter workflow for Maktbarometern data cleanup.

### 2. Server (Go)

Entry point: `server/main.go`

Active routes:

- `GET /api/status` -> basic health response.
- `GET /ws/game` -> upgrades to WebSocket and registers client in hub.

Current architecture:

- `session.GameHub`: tracks connected clients and active lobbies.
- `session.GameLobby`: room structure with register/unregister/broadcast channels.
- `words.Dictionary`: in-memory map of word -> vector entry loaded from CSV files.
- `util.CosineDistance`: similarity primitive used by gameplay.

WebSocket events (currently defined):

- Client -> server: `create_lobby`, `join_lobby`
- Server -> client: `lobby_created`, `joined_lobby`, `error`, `success`

Game modes:

- Folder exists (`server/game/`) with interface + mode files.
- Concrete mode logic files are currently placeholders.

### 3. Client (Next.js)

App routes:

- `/` -> home view (`client/components/home/HomeView.tsx`)
- `/lobby` -> lobby view (`client/components/lobby/LobbyView.tsx`)

Current frontend state:

- UI for room code input, create/join entry points, game mode cards, and lobby mock layout.
- Global `GameContextProvider` opens WebSocket connection and exposes `sendMessage`.
- Tailwind v4 + custom theme tokens + shadcn/ui primitives in use.

## NLP and Language Constraints

- Game semantics are Swedish-first.
- Vector distance calculations depend on Swedish fastText vectors.
- POS filtering and vocabulary curation are part of the preprocessing pipeline.

## 3. Game Modes

The game consists of four distinct word-based game modes. All distance/similarity calculations rely on the Swedish fastText model.

### Mode 1: The Impostor

- **Players:** Atleast 4 players (3 "Normal", 1 "Impostor").
- **Mechanic:** Normal players receive a secret word (e.g., "Äpple"). The Impostor receives a very similar word (e.g., "Päron").
- **Action:** All players have 2 seconds to type a related word.
- **Resolution:** After the input phase, players enter a 10-20 second discussion/voting phase to eliminate the suspected Impostor (or vote to skip).

### Mode 2: Contexto Versus

- **Players:** 2 or more players.
- **Mechanic:** A competitive "Contexto" style mode under time pressure.
- **Action:** Players continuously guess words to find a hidden target word.
- **Resolution:** When the timer runs out, the player whose guessed word has the closest semantic distance to the target word wins.

### Mode 3: Contexto Battle Royale (King of the Hill)

- **Players:** Multiplayer.
- **Mechanic:** The backend provides a target word (e.g., "Bil").
- **Action:** All players have 3 seconds to type a synonym or highly related word.
- **Resolution:** The backend calculates the semantic distance of all submitted words to the target word. The player with the furthest distance is eliminated. This repeats until only one player remains.

### Mode 4: Anti-Hivemind (Unique Closest)

- **Players:** Multiplayer.
- **Mechanic:** The backend provides a target word.
- **Action:** Players have 2 seconds to type a relevant/similar word. The backend enforces a maximum semantic distance threshold (using fastText) to prevent players from entering completely random words.
- **Resolution:** If two or more players type the exact same word, they all receive 0 points. Among the remaining unique words, the player with the closest distance to the target word wins.

Then plan is to export multiple csv files from the preprocessing with a words and their corresponding vectors

### How to assist me

When I ask questions, I will specify which part of the stack (Next.js, Go, or Python) or which game mode I am working on. Please provide code, architecture advice, or solutions optimized for real-time performance and scalable multiplayer architecture.
