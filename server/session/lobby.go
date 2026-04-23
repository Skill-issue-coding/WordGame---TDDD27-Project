package session

import (
	"fmt"
	"log"
	"server/events"
	"time"

	"github.com/google/uuid"
)

func NewLobby(id string) *GameLobby {
	return &GameLobby{
		ID:         id,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (lobby *GameLobby) Run() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	tick := 0

	for {
		select {
		// A client joins the room
		case client := <-lobby.Register:
			if lobby.BaseState.Phase == GameStarted {
				client.Send <- events.PrepareEvent(events.ErrorEvent, map[string]string{"message": "Spelet har redan börjat, kan inte ansluta"})
				continue
			}

			// Add the client to the map
			lobby.Clients[client] = true

			// Count number of players in each team
			countA := 0
			countB := 0
			for client := range lobby.Clients {
				// if client.Player.Team == {
				// 	countA++
				// } else {
				// 	countB++
				// }
			}

			// assignedTeam := "a"
			// if countB < countA {
			// 	assignedTeam = "b"
			// }
			// client.Team = assignedTeam

			newUser := User{
				UserId:   client.Player.UserId,
				Username: client.Username,
				Team:     assignedTeam,
			}

			r.State.Players[client.Id] = &newUser

			joinResponse := CreatedJoinGameResponse{
				GameState: r.State.ToClientState(client.Team),
				User:      newUser,
				Message:   "Du gick med i spelet!",
			}

			if client.Id != r.State.Host {
				client.send <- PrepareEvent(JoinedGameEvent, joinResponse)
			}

			log.Printf("[Room %s] Player '%s' joined (team=%s). Players in room: %d", r.ID, client.Username, assignedTeam, len(r.Clients))

			// Send out that the a new client has joined
			for c := range r.Clients {
				if c.Id != client.Id {
					c.send <- PrepareEvent(LobbyUpdateEvent, r.State.ToClientState(c.Team))
				}
			}

			// A client leaves the room
		case client := <-r.Unregister:
			delete(r.Clients, client)

			// Remove ghost locks
			if client.Team != "" {
				teamLetters := r.State.Teams[client.Team].Letters
				lettersChanged := false
				for id, letter := range teamLetters {
					if letter.IsLocked && letter.LockedBy == client.Id {
						letter.IsLocked = false
						letter.LockedBy = uuid.Nil
						teamLetters[id] = letter
						lettersChanged = true
					}
				}
				if lettersChanged {
					r.State.Teams[client.Team].Letters = teamLetters
					finalMessage := PrepareEvent(UpdatedTeamLetterEvent, UpdatedTeamLettersResponse{TeamLetters: teamLetters})
					for c := range r.Clients {
						if c.Team == client.Team {
							c.send <- finalMessage
						}
					}
				}

				delete(r.State.Players, client.Id)
				client.Team = ""
			}

			log.Printf("[Room %s] Player '%s' left. Players remaining: %d", r.ID, client.Username, len(r.Clients))

			// Delete the room if there are no other clients
			if len(r.Clients) == 0 {
				log.Printf("[Room %s] Room is empty, closing.", r.ID)
				client.hub.DeleteRoom(r.ID)
				return
			}

			if r.State.Host == client.Id {
				for remainingClient := range r.Clients {
					r.State.Host = remainingClient.Id
					break
				}
			}

			if r.State.GameStarted {
				message := PrepareEvent(ErrorEvent, map[string]string{"message": fmt.Sprintf("%s lämnade spelet", client.Username)})
				for c := range r.Clients {
					c.send <- message
				}
			}

			for c := range r.Clients {
				c.send <- PrepareEvent(LobbyUpdateEvent, r.State.ToClientState(c.Team))
			}
		}
	}
}
