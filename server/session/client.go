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
	// to keep the connection alive. Must be less than pongWait.
	pingInterval = 20 * time.Second

	// SOCKETREADLIMIT is the maximum size in bytes of a single incoming
	// WebSocket message. Messages exceeding this are rejected.
	SOCKETREADLIMIT int64 = 1024

	// MAXMESSAGESPERSECOND is the rate limit threshold. If a client exceeds
	// this many messages within a one-second window, a warning is issued.
	MAXMESSAGESPERSECOND int = 30

	// MAXMESSAGEWARNINGS is the number of rate limit violations allowed before
	// the client is forcibly disconnected.
	MAXMESSAGEWARNINGS int = 3
)

// WritePump runs in its own goroutine and is the only writer to the WebSocket
// connection. It drains the client's Send channel and forwards each message to
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
// MAXMESSAGESPERSECOND messages in a rolling one-second window more than
// MAXMESSAGEWARNINGS times, the connection is closed.
//
// When ReadPump exits for any reason, it sends the client to hub.Unregister,
// which closes the Send channel and causes WritePump to exit as well.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	if err := c.Conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		fmt.Println(err)
		return
	}

	c.Conn.SetReadLimit(SOCKETREADLIMIT)
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
		if messageCount > MAXMESSAGESPERSECOND {
			messageWarnings++
			messageCount = 0
			log.Printf("Warning: Client %s sent packages too quickly!", c.UserId)
			if messageWarnings >= MAXMESSAGEWARNINGS {
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
		// and registers them into it. The lobby's Register handler sends
		// the initial sync_gamestate back.
		case events.CreateLobbyEvent:
			code := c.Hub.CreateUniqueRoom()
			lobby := c.Hub.GetRoom(code)

			// Point the lobby's user roster at this client's profile pointer.
			// No copy is made — mutations to c.Profile are immediately visible
			// in lobby.Users and vice versa.
			lobby.Users[c.UserId] = c.Profile
			lobby.Host = c.UserId
			c.Lobby = lobby

			lobby.Register <- c

		// join_lobby — validates the room code and registers the client into
		// an existing lobby. The lobby's Register handler sends the sync.
		case events.JoinLobbyEvent:
			payload, err := events.DecodePayload[JoinLobbyPayload](event)
			if err != nil {
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

		// update_user — updates the client's username and/or background color.
		// Because lobby.Users holds a pointer to the same UserProfile as
		// c.Profile, no re-insertion is needed. A SyncRequests signal is sent
		// to the lobby so all other players receive the updated roster.
		case events.UpdateUserEvent:
			payload, err := events.DecodePayload[UpdateUserPayload](event)
			if err != nil {
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

		default:
			c.SendError("Okänd event-typ")
		}
	}
}

// pongHandler is called automatically by the gorilla/websocket library whenever
// a pong frame is received. It extends the read deadline to keep the connection
// alive for another pongWait duration.
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
// Convenience wrapper around SendEvent.
func (c *Client) SendSuccess(message string) {
	c.SendEvent(events.SuccessEvent, map[string]string{"message": message})
}

// SendError sends an error event with a human-readable message string.
// Convenience wrapper around SendEvent.
func (c *Client) SendError(message string) {
	c.SendEvent(events.ErrorEvent, map[string]string{"message": message})
}
