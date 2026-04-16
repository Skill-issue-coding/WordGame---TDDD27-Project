# Project Context

**Act as an expert Full-Stack Developer, Game Designer, and NLP Engineer.**

Below is the core context and architecture for a multiplayer word-based party game I am developing. Please keep this context in mind for all future queries in this conversation.

## 1. Core Architecture & Lobby System

The game features a standard multiplayer lobby system where players can create rooms, invite friends, select a game mode, and start matches.

- **State Management:** Real-time synchronization is required for lobbies, timers, player states, and voting mechanisms.
- **Language:** The game and all NLP calculations are strictly in Swedish.

## 2. Tech Stack

- **Frontend:** React (Next.js)
- **Backend:** Go (Handles game logic, real-time WebSocket connections, lobby state, and distance calculations).
- **Data Preprocessing & NLP:** Python. Used for preparing data pipelines, utilizing Korp (Språkbanken), Maktbarometern, custom CSV lists, and fastText models for Swedish.

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
