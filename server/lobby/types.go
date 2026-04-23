package lobby

import (
	"server/words"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type GameLobby struct {
	ID         string           // Gamecode XXXX-XXXX
	Clients    map[*Client]bool // All Connected Clients
	Broadcast  chan []byte      // Broadcast messages to the clients
	Register   chan *Client     // Client Joined The Lobby
	Unregister chan *Client     // Client Disconnects From The Lobby
}

type Client struct {
	Id       uuid.UUID       // UUID Refrence
	Username string          // Username
	Hub      *GameHub        // Gamehub Refrence
	Conn     *websocket.Conn // Websocket Connection Refrence
	Send     chan []byte     // Send chan to send messages to client
	Lobby    *GameLobby      // Lobby refrence
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
