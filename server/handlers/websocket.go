package handlers

import (
	"fmt"
	"net/http"
	"server/session"
	"server/util"

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
func HandleWebSocket(c *gin.Context, hub *session.GameHub) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}

	profile := &session.UserProfile{
		UserId:     uuid.New(),
		Username:   util.GenerateUsername(),
		Background: util.GenerateBackgroundColor(),
	}

	client := &session.Client{
		UserId:  profile.UserId,
		Profile: profile,
		Hub:     hub,
		Conn:    conn,
		Send:    make(chan []byte, 256),
	}

	client.Hub.Register <- client
	go client.WritePump()
	go client.ReadPump()
}
