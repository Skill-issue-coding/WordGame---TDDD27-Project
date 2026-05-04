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
	pongWait     = 40 * time.Second
	pingInterval = 20 * time.Second

	SOCKETREADLIMIT      int64 = 1024
	MAXMESSAGESPERSECOND int   = 30
	MAXMESSAGEWARNINGS   int   = 3
)

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
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

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

		case events.CreateGameEvent:
			code := c.Hub.CreateUniqueRoom()
			lobby := c.Hub.GetRoom(code)

			lobby.Users[c.UserId] = &UserProfile{
				UserId:     c.UserId,
				Username:   c.Username,
				Background: c.Background,
			}
			lobby.Host = c.UserId
			c.Lobby = lobby

			lobby.Register <- c

		case events.JoinGameEvent:
			payload, err := events.DecodePayload[JoinLobbyPayload](event)
			if err != nil {
				log.Printf("Error decoding join_game payload: %v", err)
				continue
			}

			code := strings.TrimSpace(payload.Code)
			if code == "" {
				c.SendError("Spelkod krävs.")
				continue
			}

			lobby := c.Hub.GetRoom(code)
			if lobby == nil {
				c.SendError("Hittade inget rum med den koden.")
				continue
			}

			if lobby.Phase == GameStarted {
				c.SendError("Spelet har redan börjat.")
				continue
			}

			c.Lobby = lobby
			lobby.Users[c.UserId] = &UserProfile{
				UserId:     c.UserId,
				Username:   c.Username,
				Background: c.Background,
			}

			lobby.Register <- c

		default:
			c.SendError("Okänd event-typ")
		}
	}
}

func (c *Client) pongHandler(_ string) error {
	return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
}

func (c *Client) SendEvent(eventType events.EventType, payload any) {
	c.Send <- events.PrepareEvent(eventType, payload)
}

func (c *Client) SendSuccess(message string) {
	c.SendEvent(events.SuccessEvent, map[string]string{"message": message})
}

func (c *Client) SendError(message string) {
	c.SendEvent(events.ErrorEvent, map[string]string{"message": message})
}
