package handlers

import (
	"net/http"
	"server/session"
	"server/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HandleStatus returns a simple JSON response
func HandleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "online",
		"message": "Game server is running",
	})
}

type NewUsernameRequest struct {
	UserId string `json:"user_id"`
}

func NewUsername(c *gin.Context, hub *session.GameHub) {
	var request NewUsernameRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ogiltig förfrågan"})
		return
	}

	userId, err := uuid.Parse(request.UserId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ogiltigt användar-id"})
		return
	}

	connected := false
	for client := range hub.Clients {
		if client.UserId == userId {
			connected = true
			break
		}
	}

	if !connected {
		c.JSON(http.StatusNotFound, gin.H{"error": "Användaren är inte ansluten"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"username": util.GenerateUsername()})
}
