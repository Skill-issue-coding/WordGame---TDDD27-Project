package session

import (
	"log"
	"server/events"
	"server/util"
	"server/words"
	"time"
)

// NewGameHub initialises and returns a GameHub with a loaded word dictionary.
// Returns an error if the dictionary files cannot be read or parsed.
// The hub's Run goroutine must be started separately by the caller.
func NewGameHub() (*GameHub, error) {
	dict, err := words.InitializeDictionary()
	if err != nil {
		return nil, err
	}

	return &GameHub{
		Dictionary: dict,
		Clients:    make(map[*Client]bool),
		Lobbys:     make(map[string]*GameLobby),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}, nil
}

// Run is the hub's main event loop and must be started in its own goroutine.
// It is the single owner of the Clients map, so all mutations to it happen
// here without additional locking.
//
// The loop handles four cases:
//   - Register: a newly connected client is added to the hub and sent a
//     connected_to_hub event containing their generated username and color.
//   - Unregister: a disconnected client is removed; if they were in a lobby,
//     they are forwarded to that lobby's Unregister channel.
//   - Broadcast: a raw message is forwarded to every connected client.
//   - statusTicker: logs a periodic health summary every 30 seconds.
func (hub *GameHub) Run() {
	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	for {
		select {

		case client := <-hub.Register:
			hub.Clients[client] = true

			// Immediately inform the client of their server-assigned identity.
			client.SendEvent(events.ConnectedEvent, ConnectedToHubPayload{
				User: UserProfile{
					UserId:     client.UserId,
					Username:   client.Username(),
					Background: client.Background(),
				},
			})

			log.Printf("[Hub] Client connected (id=%s). Connected: %d | Rooms open: %d | Players in rooms: %d",
				client.UserId, len(hub.Clients), len(hub.Lobbys), hub.totalPlayers())

		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client]; ok {
				if client.Lobby != nil {
					room := client.Lobby
					client.Lobby = nil
					// Forward to the lobby's unregister channel in a goroutine
					// to avoid a deadlock between the hub and lobby event loops.
					go func() { room.Unregister <- client }()
				}

				delete(hub.Clients, client)
				close(client.Send) // signals WritePump to exit

				log.Printf("[Hub] Client disconnected (id=%s). Connected: %d | Rooms open: %d | Players in rooms: %d",
					client.UserId, len(hub.Clients), len(hub.Lobbys), hub.totalPlayers())
			}

		case message := <-hub.Broadcast:
			for client := range hub.Clients {
				select {
				case client.Send <- message:
				default:
					// Send buffer is full — treat as a dead connection.
					close(client.Send)
					delete(hub.Clients, client)
				}
			}

		case <-statusTicker.C:
			log.Printf("[Hub] Status — Open rooms: %d | Players in rooms: %d | Connected clients: %d",
				len(hub.Lobbys), hub.totalPlayers(), len(hub.Clients))
		}
	}
}

// totalPlayers returns the total number of players across all active lobbies.
// Must only be called from within the Run goroutine (or while holding
// RoomsMutex) as it reads the Lobbys map.
func (hub *GameHub) totalPlayers() int {
	total := 0
	for _, room := range hub.Lobbys {
		total += len(room.Clients)
	}
	return total
}

// CreateUniqueRoom generates a collision-free room code, creates a new lobby,
// starts its Run goroutine, registers it in the hub, and returns the code.
// It is safe to call from any goroutine.
func (hub *GameHub) CreateUniqueRoom() string {
	hub.RoomsMutex.Lock()
	defer hub.RoomsMutex.Unlock()

	var code string
	for {
		code = util.GenerateGameCode()
		if _, exists := hub.Lobbys[code]; !exists {
			newRoom := NewLobby(code)
			hub.Lobbys[code] = newRoom
			go newRoom.Run()
			log.Printf("[Hub] Room created (code=%s). Open rooms: %d", code, len(hub.Lobbys))
			break
		}
	}
	return code
}

// GetRoom returns the GameLobby for the given room code, or nil if no such
// lobby exists. It is safe to call from any goroutine.
func (hub *GameHub) GetRoom(code string) *GameLobby {
	hub.RoomsMutex.RLock()
	defer hub.RoomsMutex.RUnlock()
	return hub.Lobbys[code]
}

// DeleteRoom removes the lobby with the given code from the hub. It is
// typically called by the lobby's own Run goroutine when the last player
// leaves. It is safe to call from any goroutine.
func (hub *GameHub) DeleteRoom(code string) {
	hub.RoomsMutex.Lock()
	defer hub.RoomsMutex.Unlock()
	delete(hub.Lobbys, code)
	log.Printf("[Hub] Room deleted (code=%s). Open rooms: %d | Players in rooms: %d",
		code, len(hub.Lobbys), hub.totalPlayers())
}
