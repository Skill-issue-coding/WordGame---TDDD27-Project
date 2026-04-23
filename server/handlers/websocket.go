package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
func HandleWebSocket(c *gin.Context) {
	// TODO: Implement this
}
