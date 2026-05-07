package session

import (
	"fmt"
	"log"
	"server/events"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// pongWait is the maximum time to wait for a pong response before
	// treating the connection as dead and closing it.
	pongWait = 40 * time.Second

	// pingInterval is how often the server sends a ping frame to the client
	// to keep the connection alive. It must be less than pongWait.
	pingInterval = 20 * time.Second

	// socketReadLimit is the maximum size in bytes of a single incoming
	// WebSocket message. Messages exceeding this limit are rejected.
	socketReadLimit int64 = 1024

	// maxMessagesPerSecond is the rate limit threshold. If a client exceeds
	// this many messages within a one-second window, a warning is issued.
	maxMessagesPerSecond int = 30

	// maxMessageWarnings is the number of rate limit violations allowed before
	// the client is forcibly disconnected.
	maxMessageWarnings int = 3
)

// WritePump runs in its own goroutine and is the only writer to the WebSocket
// connection. It drains the client's Send channel, forwards each message to
// the socket, and sends periodic ping frames to keep the connection alive.
//
// When the Send channel is closed (by the hub on disconnect), WritePump sends
// a WebSocket close frame and exits, which in turn causes ReadPump to exit.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Hub closed the channel — send a graceful close frame.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump runs in its own goroutine and is the only reader from the WebSocket
// connection. It parses incoming event envelopes and dispatches them to the
// appropriate handler logic via a switch on event.Type.
//
// ReadPump also enforces per-client rate limiting: if a client sends more than
// maxMessagesPerSecond messages in a rolling one-second window more than
// maxMessageWarnings times, the connection is closed.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		fmt.Println(err)
		return
	}

	c.Conn.SetReadLimit(socketReadLimit)
	c.Conn.SetPongHandler(c.pongHandler)

	messageCount := 0
	messageWarnings := 0
	windowStart := time.Now()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Rate limiting: reset the counter every second.
		now := time.Now()
		if now.Sub(windowStart) >= time.Second {
			messageCount = 0
			windowStart = now
		}

		messageCount++
		if messageCount > maxMessagesPerSecond {
			messageWarnings++
			messageCount = 0
			log.Printf("Warning: Client %s sent packages too quickly!", c.UserId)
			if messageWarnings >= maxMessageWarnings {
				break
			}
			continue
		}

		event, err := events.ParseEvent(message)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			continue
		}

		switch event.Type {

		// create_lobby — creates a new room, assigns this client as host,
		// and registers them into it.
		case events.CreateLobbyEvent:
			code := c.Hub.CreateUniqueRoom()
			lobby := c.Hub.GetRoom(code)

			lobby.Users[c.UserId] = c.Profile
			lobby.Host = c.UserId
			c.Lobby = lobby

			lobby.Register <- c

		// join_lobby — validates the room code and registers the client into
		// an existing lobby.
		case events.JoinLobbyEvent:
			payload, err := events.DecodePayload[JoinLobbyPayload](event)
			if err != nil {
				c.SendError("Serverfel vid inläsningen av lobby koden")
				log.Printf("Error decoding join_game payload: %v", err)
				continue
			}

			lobbyCode := strings.TrimSpace(payload.LobbyCode)
			if lobbyCode == "" {
				c.SendError("Spelkod krävs.")
				continue
			}

			lobby := c.Hub.GetRoom(lobbyCode)
			if lobby == nil {
				c.SendError("Hittade inget rum med den koden.")
				continue
			}

			if lobby.Phase == GameStarted {
				c.SendError("Spelet har redan börjat.")
				continue
			}

			lobby.Users[c.UserId] = c.Profile
			c.Lobby = lobby
			lobby.Register <- c

		case events.LeaveLobbyRequestEvent:
			if c.Lobby == nil {
				c.SendError("Du är inte i ett rum")
				continue
			}
			c.Lobby.Unregister <- c

		// update_user — updates the client's username and/or background color.
		case events.UpdateUserEvent:
			payload, err := events.DecodePayload[UpdateUserPayload](event)
			if err != nil {
				c.SendError("Serverfel vid inläsningen av uppdateringarna")
				log.Printf("Error decoding update_user payload: %v", err)
				continue
			}

			if username := strings.TrimSpace(payload.Username); username != "" {
				c.Profile.Username = username
			}
			if payload.Background != "" {
				c.Profile.Background = payload.Background
			}

			if c.Lobby != nil {
				c.Lobby.SyncRequests <- struct{}{}
			}

		case events.ChatMessageRequestEvent:
			payload, err := events.DecodePayload[ChatMessageRequestPayload](event)
			if err != nil {
				c.SendError("Serverfel vid skickandet av meddelandet")
				log.Printf("Error decoding send_chatmessage payload: %v", err)
				continue
			}
			if c.Lobby == nil {
				c.SendError("Du är inte i ett rum")
				continue
			}

			serverTimestamp := float64(time.Now().UnixMilli())
			chatMessage := ChatMessage{
				Sender:  *c.Profile,
				Message: payload.Message,
				Date:    serverTimestamp,
			}

			c.Lobby.ChatMessages <- chatMessage

		case events.ChangeModeEvent:
			if c.Lobby == nil || c.Lobby.Host != c.UserId {
				c.SendError("Endast hosten kan ändra spelläge.")
				continue
			}

			payload, err := events.DecodePayload[ChangeModePayload](event)
			if err != nil {
				c.SendError("Serverfel vid inläsningen av spelläge")
				log.Printf("Error decoding change_mode payload: %v", err)
				continue
			}

			c.Lobby.ModeUpdateRequests <- payload.Mode

		case events.UpdateSettingEvent:
			if c.Lobby == nil || c.Lobby.Host != c.UserId {
				c.SendError("Endast hosten kan ändra inställningar.")
				continue
			}
			payload, err := events.DecodePayload[UpdateSettingPayload](event)
			if err != nil {
				c.SendError("Serverfel vid uppdatering av inställningarna")
				log.Printf("Error decoding update_setting payload: %v", err)
				continue
			}
			c.Lobby.SettingUpdateRequests <- payload

		default:
			log.Printf("Unknown event type %s", event.Type)
			c.SendError("Okänd event-typ")
		}
	}
}

// pongHandler is called automatically by the gorilla/websocket library whenever
// a pong frame is received. It extends the read deadline to keep the connection alive.
func (c *Client) pongHandler(_ string) error {
	return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
}

// SendEvent serialises the given payload into an event envelope with the
// specified type and queues it on the client's Send channel for WritePump
// to deliver. It is safe to call from any goroutine.
func (c *Client) SendEvent(eventType events.EventType, payload any) {
	b, err := events.PrepareEvent(eventType, payload)
	if err != nil {
		log.Printf("error preparing event: %v", err)
		return
	}
	c.Send <- b
}

// SendSuccess sends a success event with a human-readable message string.
// It is a convenience wrapper around SendEvent.
func (c *Client) SendSuccess(message string) {
	c.SendEvent(events.SuccessEvent, map[string]string{"message": message})
}

// SendError sends an error event with a human-readable message string.
// It is a convenience wrapper around SendEvent.
func (c *Client) SendError(message string) {
	c.SendEvent(events.ErrorEvent, map[string]string{"message": message})
}
