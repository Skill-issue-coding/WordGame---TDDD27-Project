package session

import (
	"server/words"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GamePhase string

const (
	LobbyState  GamePhase = "lobby"
	GameStarted GamePhase = "game_started"
)

// Used for websocket connection purpose
type Client struct {
	UserId   uuid.UUID       // UUID Refrence
	Username string          // Username
	Hub      *GameHub        // Gamehub Refrence
	Conn     *websocket.Conn // Websocket Connection Refrence
	Send     chan []byte     // Send chan to send messages to client
	Lobby    *GameLobby      // Lobby refrence
}

// User represents a client-facing user payload.
type User struct {
	UserId   string `json:"id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
	Team     int    `json:"team"`
	Avatar   string `json:"avatar"`
}

// UserProfile stores identity/cosmetic data and is not gameplay state.
type UserProfile struct {
	UserId   uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   string    `json:"avatar"`
}

// ParticipantState stores gameplay-only values for active players.
type ParticipantState struct {
	Score      int  `json:"score"`
	Team       int  `json:"team"`
	Eliminated bool `json:"eliminated"`
}

type GameHub struct {
	Dictionary words.Dictionary      // Word struct refrence
	Clients    map[*Client]bool      // Connected clients map
	Lobbys     map[string]*GameLobby // Lobbys map
	RoomsMutex sync.RWMutex          // Read/Write mutex
	Broadcast  chan []byte           // Broadcast chan
	Register   chan *Client          // Register user chan
	Unregister chan *Client          // Unregister user chan
}
type GameLobby struct {
	ID         string                     // Gamecode XXXX-XXXX
	Clients    map[*Client]bool           // All Connected Clients
	Profiles   map[uuid.UUID]*UserProfile // All user profiles in the lobby
	Broadcast  chan []byte                // Broadcast messages to the clients
	Register   chan *Client               // Client Joined The Lobby
	Unregister chan *Client               // Client Disconnects From The Lobby
	Host       uuid.UUID                  // Client Connection That Is Host
	BaseState  BaseState                  // Base Game State
}

// BaseState holds the data EVERY game mode shares
type BaseState struct {
	Mode         string
	Phase        GamePhase
	StartTime    float64
	EndTime      float64
	Host         uuid.UUID
	Participants map[uuid.UUID]*ParticipantState
}

// BaseStateClient is a sanitized state payload sent to frontend clients.
type BaseStateClient struct {
	Mode      string             `json:"mode"`
	Phase     GamePhase          `json:"phase"`
	StartTime float64            `json:"startTime"`
	EndTime   float64            `json:"endTime"`
	Host      string             `json:"host"`
	Users     map[uuid.UUID]User `json:"users"`
}
