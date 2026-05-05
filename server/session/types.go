package session

import (
	"server/game"
	"server/words"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GamePhase string
type GameMode string

const (
	LobbyPhase  GamePhase = "lobby"
	GameStarted GamePhase = "game_started"

	ModeImpostor       GameMode = "impostor"
	ModeContextoBattle GameMode = "contexto_battle"
	ModeSynonymDuel    GameMode = "synonym_duel"
	ModeAntiMatch      GameMode = "anti_match"
)

type Client struct {
	UserId  uuid.UUID
	Profile *UserProfile
	Hub     *GameHub
	Conn    *websocket.Conn
	Send    chan []byte
	Lobby   *GameLobby
}

type UserProfile struct {
	UserId     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	Background string    `json:"background"`
	Score      int       `json:"score"`
}

// LobbyState is the shared state broadcast to all clients.
// Settings is always the struct for the currently selected mode.
type LobbyState struct {
	Code     string                     `json:"code"`
	Mode     GameMode                   `json:"mode"`
	Phase    GamePhase                  `json:"phase"`
	Host     uuid.UUID                  `json:"host"`
	Users    map[uuid.UUID]*UserProfile `json:"users"`
	Settings any                        `json:"settings"` // typed per mode, see ModeSettings()
}

type GameLobby struct {
	ID           string
	Clients      map[*Client]bool
	Broadcast    chan []byte
	Register     chan *Client
	Unregister   chan *Client
	SyncRequests chan struct{}
	Host         uuid.UUID
	Phase        GamePhase

	// Shared player roster
	Users map[uuid.UUID]*UserProfile

	// Active mode + its settings (only one is non-nil at a time)
	Mode                   GameMode
	ImpostorSettings       game.ImpostorSettings
	ContextoBattleSettings game.ContextoBattleSettings
	SynonymDuelSettings    game.SynonymDuelSettings
	AntiMatchSettings      game.AntiHiveSettings
}

type GameHub struct {
	Dictionary words.Dictionary
	Clients    map[*Client]bool
	Lobbys     map[string]*GameLobby
	RoomsMutex sync.RWMutex
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}
