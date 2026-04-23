package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleStatus returns a simple JSON response
func HandleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "online",
		"message": "Game server is running",
	})
}
