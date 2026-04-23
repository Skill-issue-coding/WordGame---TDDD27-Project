package session

import (
	"server/words"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GameState string

const (
	LobbyState  = "lobby"
	GameStarted = "game_started"
)

type Client struct {
	UserId   uuid.UUID       // UUID Refrence
	Username string          // Username
	Hub      *GameHub        // Gamehub Refrence
	Conn     *websocket.Conn // Websocket Connection Refrence
	Send     chan []byte     // Send chan to send messages to client
	Lobby    *GameLobby      // Lobby refrence
	User     *User           // User Refrence
}

type User struct {
	UserId   string `json:"id"`
	Username string `json:"username"`
	Score    int    `json:"score"`
	Team     int    `json:"team"`
	Avatar   string `json:"avatar"`
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
	ID         string           // Gamecode XXXX-XXXX
	Clients    map[*Client]bool // All Connected Clients
	Broadcast  chan []byte      // Broadcast messages to the clients
	Register   chan *Client     // Client Joined The Lobby
	Unregister chan *Client     // Client Disconnects From The Lobby
	Host       uuid.UUID        // Client Connection That Is Host
	BaseState  BaseState        // Base Game State
}

// BaseState holds the data EVERY game mode shares
type BaseState struct {
	Mode          string              `json:"mode"`
	Phase         string              `json:"phase"` // e.g., "starting", "playing", "voting", "ended"
	TimeRemaining float64             `json:"timeRemaining"`
	Host          uuid.UUID           `json:"host"`
	Players       map[uuid.UUID]*User `json:"users"`
}
