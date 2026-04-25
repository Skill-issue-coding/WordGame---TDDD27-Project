package session

import (
	"log"
	"server/events"
	"time"

	"github.com/google/uuid"
)

func NewLobby(id string) *GameLobby {
	return &GameLobby{
		ID:         id,
		Clients:    make(map[*Client]bool),
		Profiles:   make(map[uuid.UUID]*UserProfile),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		BaseState: BaseState{
			Phase:        LobbyState,
			Participants: make(map[uuid.UUID]*ParticipantState),
		},
	}
}

func (lobby *GameLobby) Run() {
	gameticker := time.NewTicker(1 * time.Second)
	defer gameticker.Stop()

	for {
		select {
		case client := <-lobby.Register:
			if lobby.BaseState.Phase == GameStarted {
				client.Send <- events.PrepareEvent(events.ErrorEvent, map[string]string{"message": "Spelet har redan börjat, kan inte ansluta"})
				continue
			}

			lobby.Clients[client] = true
			client.Lobby = lobby
			lobby.EnsureParticipant(client)

			log.Printf("[Room %s] Player '%s' joined. Players in room: %d", lobby.ID, client.Username, len(lobby.Clients))

		case client := <-lobby.Unregister:
			if _, exists := lobby.Clients[client]; !exists {
				continue
			}

			delete(lobby.Clients, client)
			lobby.RemoveParticipant(client.UserId)

			log.Printf("[Room %s] Player '%s' left. Players remaining: %d", lobby.ID, client.Username, len(lobby.Clients))

			if len(lobby.Clients) == 0 {
				log.Printf("[Room %s] Room is empty, closing.", lobby.ID)
				if client.Hub != nil {
					client.Hub.DeleteRoom(lobby.ID)
				}
				return
			}

			if lobby.Host == client.UserId {
				for remainingClient := range lobby.Clients {
					lobby.Host = remainingClient.UserId
					lobby.BaseState.Host = remainingClient.UserId
					break
				}
			}

		case <-gameticker.C:
			// Placeholder tick for game mode timers.
		}
	}
}

// EnsureParticipant syncs lobby profile data with gameplay participant data.
func (lobby *GameLobby) EnsureParticipant(client *Client) {
	if lobby.Profiles == nil {
		lobby.Profiles = make(map[uuid.UUID]*UserProfile)
	}

	if lobby.BaseState.Participants == nil {
		lobby.BaseState.Participants = make(map[uuid.UUID]*ParticipantState)
	}

	lobby.Profiles[client.UserId] = &UserProfile{
		UserId:   client.UserId,
		Username: client.Username,
	}

	if _, exists := lobby.BaseState.Participants[client.UserId]; !exists {
		lobby.BaseState.Participants[client.UserId] = &ParticipantState{}
	}
}

func (lobby *GameLobby) RemoveParticipant(userID uuid.UUID) {
	delete(lobby.Profiles, userID)
	delete(lobby.BaseState.Participants, userID)
}

// BuildUserList creates a deterministic user list for client payloads.
func (lobby *GameLobby) BuildUserList() map[uuid.UUID]User {
	var users map[uuid.UUID]User = make(map[uuid.UUID]User)
	for id := range lobby.BaseState.Participants {
		users[id] = lobby.BuildUser(id)
	}

	return users
}

func (lobby *GameLobby) BuildUser(userID uuid.UUID) User {
	profile, hasProfile := lobby.Profiles[userID]
	participant, hasParticipant := lobby.BaseState.Participants[userID]

	if !hasProfile {
		profile = &UserProfile{UserId: userID}
	}

	if !hasParticipant {
		participant = &ParticipantState{}
	}

	return User{
		UserId:   userID.String(),
		Username: profile.Username,
		Score:    participant.Score,
		Team:     participant.Team,
		Avatar:   profile.Avatar,
	}
}

func (lobby *GameLobby) BuildBaseStateForClient() BaseStateClient {
	return BaseStateClient{
		Mode:      lobby.BaseState.Mode,
		Phase:     lobby.BaseState.Phase,
		StartTime: lobby.BaseState.StartTime,
		EndTime:   lobby.BaseState.EndTime,
		Host:      lobby.BaseState.Host.String(),
		Users:     lobby.BuildUserList(),
	}
}
