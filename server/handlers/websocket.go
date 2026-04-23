package handlers

import (
	"fmt"
	"net/http"
	"server/lobby"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// TODO: Implement propper checking
	CheckOrigin: func(r *http.Request) bool {
		// For development, allow all origins.
		return true
	},
}

// HandleWebSocket upgrades the Gin request
func HandleWebSocket(c *gin.Context, hub *lobby.GameHub) {

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}

	client := &lobby.Client{
		Id:   uuid.New(),
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	client.Hub.Register <- client
	go client.WritePump()
	go client.ReadPump()

	// TODO: Implement this
}
