package lobby

import (
	"fmt"
	"time"
)

const (
	pongWait      = 60 * time.Second
	pingIntervall = 50 * time.Second

	SOCKETREADLIMIT      int64 = 1024
	MAXMESSAGESPERSECOND int   = 30
	MAXMESSAGEWARNINGS   int   = 3
)

func (c *Client) WritePump() {

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
}

// pongHandler handles websocket pong messages by extending the read deadline,
// ensuring the connection is kept alive.
func (c *Client) pongHandler(pongMessage string) error {
	return c.Conn.SetReadDeadline(time.Now().Add(pongWait))
}
