package lobby

import (
	"server/words"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Dictionary struct {
	ActiveWord        string
	WordMap           map[string]words.WordEntry
	RandomWord        func() (words.WordEntry, error)
	CalculateDistance func(word string) float64
	IsValid           func(word string) bool
}

type Lobby struct {
}

type Client struct {
	Id       uuid.UUID       // UUID Refrence
	Username string          // Username
	Hub      *GameHub        // Gamehub Refrence
	Conn     *websocket.Conn // Websocket Connection Refrence
	Send     chan []byte     // Send chan to send messages to client
	Lobby    *Lobby          // Lobby refrence
}

type GameHub struct {
	Dictionary Dictionary        // Word struct refrence
	Clients    map[*Client]bool  // Connected clients map
	Lobbys     map[string]*Lobby // Lobbys map
	RoomsMutex sync.RWMutex      // Read/Write mutex
	Broadcast  chan []byte       // Broadcast chan
	Register   chan *Client      // Register user chan
	Unregister chan *Client      // Unregister user chan
}
