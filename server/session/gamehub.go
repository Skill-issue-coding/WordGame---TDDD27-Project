package session

import (
	"log"
	"server/util"
	"server/words"
	"time"
)

func NewGameHub() (*GameHub, error) {
	dict, err := words.InitializeDictionary()
	if err != nil {
		return nil, err
	}

	gameHub := &GameHub{
		Dictionary: dict,
		Clients:    make(map[*Client]bool),
		Lobbys:     make(map[string]*GameLobby),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}

	// gameHub.Dictionary.CalculateDistance = func(word string) float64 {
	// 	activeWordEntry, activeWordExists := wordMap[gameHub.Dictionary.ActiveWord]
	// 	guessEntry, guessExists := wordMap[word]

	// 	if !activeWordExists || !guessExists {
	// 		return math.NaN()
	// 	}

	// 	return cosineDistance(activeWordEntry.WordVector, guessEntry.WordVector)
	// }

	// if err := gameHub.SetRandomActiveWord(); err != nil {
	// 	return nil, err
	// }

	return gameHub, nil
}

func (hub *GameHub) Run() {
	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	for {
		select {
		// New Client Joins
		case client := <-hub.Register:
			hub.Clients[client] = true
			log.Printf("[Hub] Client connected (id=%s). Connected: %d | Rooms open: %d | Players in rooms: %d",
				client.UserId,
				len(hub.Clients),
				len(hub.Lobbys),
				hub.totalPlayers())

		// Client Disconnects
		case client := <-hub.Unregister:
			if _, ok := hub.Clients[client]; ok {
				if client.Lobby != nil {
					room := client.Lobby
					client.Lobby = nil
					go func() { room.Unregister <- client }()
				}

				delete(hub.Clients, client)
				close(client.Send)

				log.Printf("[Hub] Client disconnected (id=%s). Connected: %d | Rooms open: %d | Players in rooms: %d",
					client.UserId,
					len(hub.Clients),
					len(hub.Lobbys),
					hub.totalPlayers())
			}

		// BroadCast A Global Message
		case message := <-hub.Broadcast:
			for client := range hub.Clients {
				select {
				case client.Send <- message:
				default:
					// If there was an error, close the connection
					close(client.Send)
					delete(hub.Clients, client)
				}
			}

		// Periodic status log
		case <-statusTicker.C:
			log.Printf("[Hub] Status — Open rooms: %d | Players in rooms: %d | Connected clients: %d",
				len(hub.Lobbys),
				hub.totalPlayers(),
				len(hub.Clients))
		}
	}
}

// totalPlayers calculates and returns the aggregate number of clients across all active rooms in the hub.
func (hub *GameHub) totalPlayers() int {
	total := 0
	for _, room := range hub.Lobbys {
		total += len(room.Clients)
	}
	return total
}

// CreateUniqueRoom generates a unique game code, creates a new GameRoom,
// starts its run loop in a goroutine, and returns the generated code.
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

// GetRoom retrieves and returns a pointer to the GameRoom associated with the given code.
// It utilizes a read-write mutex to ensure thread safety.
func (hub *GameHub) GetRoom(code string) *GameLobby {
	hub.RoomsMutex.RLock()
	defer hub.RoomsMutex.RUnlock()
	return hub.Lobbys[code]
}

// DeleteRoom securely removes the GameRoom associated with the given code from the hub's active rooms.
func (hub *GameHub) DeleteRoom(code string) {
	hub.RoomsMutex.Lock()
	defer hub.RoomsMutex.Unlock()

	delete(hub.Lobbys, code)
	log.Printf("[Hub] Room deleted (code=%s). Open rooms: %d | Players in rooms: %d", code, len(hub.Lobbys), hub.totalPlayers())
}
