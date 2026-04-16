package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"server/hub"
	"server/words"
	"strings"
)

func main() {
	gameHub, err := hub.NewGameHub(words.TERMINAL_TEST_VECTOR_FILES)
	if err != nil {
		log.Fatalf("Could not initialize game hub: %v", err)
	}

	terminalTesting(gameHub)

	// // Create a Gin router with default middleware (logger and recovery)
	// r := gin.Default()

	// // Define a simple GET endpoint
	// r.GET("/ping", func(c *gin.Context) {
	// 	// Return JSON response
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "pong",
	// 	})
	// })

	// // Start server on port 8080 (default)
	// // Server will listen on 0.0.0.0:8080 (localhost:8080 on Windows)
	// r.Run()
}

func terminalTesting(gameHub *hub.GameHub) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Game terminal started. Type a word and press Enter (type 'exit' to quit):")
	fmt.Printf("Loaded %d words\n", len(gameHub.Dictionary.WordMap))
	fmt.Printf("Active word for this round: %s\n", gameHub.Dictionary.ActiveWord)
	fmt.Println("Type 'new' to start a new round with a random active word")

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()

		input = strings.TrimSpace(input)

		if input == "exit" {
			fmt.Println("Shutting down...")
			break
		}

		if input == "new" {
			if err := gameHub.SetRandomActiveWord(); err != nil {
				log.Printf("Could not start new round: %v\n", err)
				continue
			}

			fmt.Printf("New active word: %s\n", gameHub.Dictionary.ActiveWord)
			continue
		}

		if wordEntry, exists := gameHub.GetWordEntry(input); exists {
			log.Printf("Word exists: %s, Beginning of vector: %.6f \n", wordEntry.Word, wordEntry.WordVector[0])
			distance := gameHub.Dictionary.CalculateDistance(input)
			similarity := 1 - distance
			log.Printf("Distance to active word '%s': %.6f\n", gameHub.Dictionary.ActiveWord, distance)
			log.Printf("Similarity to active word '%s': %.6f\n", gameHub.Dictionary.ActiveWord, similarity)
		} else {
			log.Printf("Word not in dictionary: %s\n", input)
		}

		fmt.Printf("Backend processed word: '%s'\n", input)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}
