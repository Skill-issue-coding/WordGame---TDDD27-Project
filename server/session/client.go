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
	pongWait      = 40 * time.Second
	pingIntervall = 20 * time.Second

	SOCKETREADLIMIT      int64 = 1024
	MAXMESSAGESPERSECOND int   = 30
	MAXMESSAGEWARNINGS   int   = 3
)

// writePump pumps messages from the hub to the websocket connection.
// It listens on the client's send channel and writes the messages to the socket,
// while also handling periodic ping messages to keep the connection alive.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingIntervall)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The gamehub closed the channel
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

		// If the websocket crashes
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// SOME IMPLEMENTAION OF RATE LIMITING ?
		now := time.Now()
		if now.Sub(windowStart) >= time.Second {
			// Reset time window
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

		// Read event envelope once, then decode payload by type.
		event, err := events.ParseEvent(message)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			continue
		}

		// Switch case for each event
		switch event.Type {
		case events.CreateGameEvent:
			createData, err := events.DecodePayload[CreateLobbyPayload](event)
			if err != nil {
				log.Printf("Error when reading create_game payload: %v", err)
				continue
			}

			username := strings.TrimSpace(createData.Username)
			if username == "" {
				c.sendError("Användarnamnet får inte vara tomt.")
				continue
			}

			c.Username = username

			// Create the room and register the client to the room
			code := c.Hub.CreateUniqueRoom()
			lobby := c.Hub.GetRoom(code)

			// Set host
			lobby.Host = c.UserId
			lobby.BaseState.Host = c.UserId

			c.Lobby = lobby
			lobby.Register <- c

			c.sendEvent(events.GameCreatedEvent, CreatedLobbyResponsePayload{
				User:      User{UserId: c.UserId.String(), Username: c.Username},
				GameState: lobby.BuildBaseStateForClient(),
				Message:   "Du skapade ett nytt spel!"})

		case events.JoinGameEvent:
		default:
			c.sendError("Okänd event-typ")
		}
	}
}

// pongHandler handles websocket pong messages by extending the read deadline,
// ensuring the connection is kept alive.
func (c *Client) pongHandler(pongMessage string) error {
	return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
}

// Helper function to send events to the client
func (c *Client) sendEvent(eventType events.EventType, payload any) {
	c.Send <- events.PrepareEvent(eventType, payload)
}

// Helper function to just send a "Success Event"
func (c *Client) sendSuccess(message string) {
	c.sendEvent(events.SuccessEvent, map[string]string{"message": message})
}

// Helper function to just send a "Error Event"
func (c *Client) sendError(message string) {
	c.sendEvent(events.ErrorEvent, map[string]string{"message": message})
}
