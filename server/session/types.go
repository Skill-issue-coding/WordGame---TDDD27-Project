// Package session manages all WebSocket client connections, game lobbies,
// and the central hub that coordinates them. It is the core runtime layer
// of the game server.
package session

import (
	"server/game"
	"server/words"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// GamePhase represents the current lifecycle stage of a lobby.
type GamePhase string

// GameMode identifies which of the four game modes is active in a lobby.
type GameMode string

const (
	// LobbyPhase indicates the lobby is in the waiting room — players can join
	// and settings can be changed, but the game has not started.
	LobbyPhase GamePhase = "lobby"

	// GameStarted indicates the game is in progress. New players cannot join.
	GameStarted GamePhase = "game_started"

	// ModeImpostor is the "Who is the Impostor?" word-deduction mode.
	ModeImpostor GameMode = "impostor"

	// ModeContextoBattle is the competitive Contexto-style guessing mode.
	ModeContextoBattle GameMode = "contexto_battle"

	// ModeSynonymDuel is the King-of-the-Hill synonym elimination mode.
	ModeSynonymDuel GameMode = "synonym_duel"

	// ModeAntiMatch is the Anti-Hivemind unique-word mode.
	ModeAntiMatch GameMode = "anti_match"
)

// Client represents a single connected WebSocket player.
// It is created when a player upgrades their HTTP connection and lives until
// the connection is closed.
//
// Profile is the single source of truth for player identity. Both the Client
// and any lobby the player joins point to the same *UserProfile, so updates
// (e.g. username changes) are reflected everywhere without an explicit sync.
type Client struct {
	// UserId is the player's unique identifier, assigned at connection time.
	UserId uuid.UUID

	// Profile holds the player's display identity. Use the Username() and
	// Background() helpers rather than accessing fields directly.
	Profile *UserProfile

	// Hub is a reference to the central GameHub this client is registered with.
	Hub *GameHub

	// Conn is the underlying gorilla/websocket connection.
	Conn *websocket.Conn

	// Send is a buffered channel of outgoing JSON-encoded event messages.
	// Write to it via SendEvent — never directly.
	Send chan []byte

	// Lobby is the room the client is currently in, or nil if they are in
	// the hub but have not yet joined a lobby.
	Lobby *GameLobby
}

// Username returns the player's current display name from their profile.
func (c *Client) Username() string { return c.Profile.Username }

// Background returns the player's current background color from their profile.
func (c *Client) Background() string { return c.Profile.Background }

// UserProfile stores a player's identity and cosmetic data.
// It is shared by pointer between Client and GameLobby.Users, meaning a
// single mutation (e.g. Profile.Username = "x") is immediately visible to
// anyone holding that pointer — no copy needed.
type UserProfile struct {
	UserId     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	Background string    `json:"background"`
	Score      int       `json:"score"`
}

// LobbyState is the complete shared game state that is broadcast to every
// client in a lobby on each sync. It contains no private per-player data.
//
// Settings holds the mode-specific configuration struct (e.g. ImpostorSettings)
// for the currently active GameMode. It is serialised as a concrete JSON object,
// so the frontend receives fully typed settings without any extra mapping.
type LobbyState struct {
	Code     string                     `json:"code"`
	Mode     GameMode                   `json:"mode"`
	Phase    GamePhase                  `json:"phase"`
	Host     uuid.UUID                  `json:"host"`
	Users    map[uuid.UUID]*UserProfile `json:"users"`
	Settings any                        `json:"settings"`
}

// GameLobby is an active game room. It has its own Run() goroutine that
// processes all state changes sequentially via channels, making it safe
// to mutate lobby fields only from within that goroutine.
//
// Each lobby holds one set of mode-specific settings. Only the settings
// struct matching the active Mode should be considered valid; the others
// hold their zero/default values.
type GameLobby struct {
	// ID is the human-readable room code (e.g. "AbCd-1234").
	ID string

	// Clients is the set of currently connected players in this lobby.
	Clients map[*Client]bool

	// Broadcast sends a raw message to all clients in the lobby.
	Broadcast chan []byte

	// Register adds a new client to the lobby. Send the *Client to this
	// channel from ReadPump; the lobby's Run() loop handles the rest.
	Register chan *Client

	// Unregister removes a client from the lobby (on disconnect or leave).
	Unregister chan *Client

	// ChatmMessages is a chan to send chatmessages to all clients in a lobby
	ChatMessages chan ChatMessage

	// SyncRequests is a signal channel. Send an empty struct to trigger a
	// SyncStateToClients broadcast without going through Register/Unregister.
	// Used by UpdateUserEvent to propagate profile changes.
	SyncRequests chan struct{}

	// Host is the UserId of the player with host privileges.
	Host uuid.UUID

	// Phase is the current lifecycle stage of the lobby.
	Phase GamePhase

	// Users is the player roster keyed by UserId. Values are shared pointers
	// with Client.Profile — mutating a UserProfile here mutates it everywhere.
	Users map[uuid.UUID]*UserProfile

	Mode                   GameMode
	ImpostorSettings       game.ImpostorSettings
	ContextoBattleSettings game.ContextoBattleSettings
	SynonymDuelSettings    game.SynonymDuelSettings
	AntiMatchSettings      game.AntiHiveSettings
}

// GameHub is the top-level coordinator for all connected clients and lobbies.
// It runs a single goroutine (Run) that owns the Clients and Lobbys maps,
// ensuring they are only mutated by one goroutine at a time.
//
// Lobbys access is additionally protected by RoomsMutex for the helper methods
// (CreateUniqueRoom, GetRoom, DeleteRoom) which may be called from outside Run.
type GameHub struct {
	// Dictionary is the in-memory Swedish fastText word vector store used by
	// all game modes for semantic distance calculations.
	Dictionary words.Dictionary

	// Clients is the set of all currently connected WebSocket clients,
	// including those not yet in a lobby.
	Clients map[*Client]bool

	// Lobbys maps room codes to active GameLobby instances.
	Lobbys map[string]*GameLobby

	// RoomsMutex guards Lobbys for concurrent access from handler goroutines.
	RoomsMutex sync.RWMutex

	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

type ChatMessage struct {
	Sender  UserProfile `json:"sender"`
	Message string      `json:"message"`
	Date    float64     `json:"date"`
}
