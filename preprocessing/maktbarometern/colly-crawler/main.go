package main

import (
	//"colly-crawler/internal/scraper"
	"colly-crawler/internal/formatter"
	//"encoding/csv"
	//"fmt"
	"log"
	//"os"
	//"strings"
)

func main() {

	files := []string{
		"maktbarometern-2025-facebook.csv",
		"maktbarometern-2025-instagram.csv",
		"maktbarometern-2025-youtube.csv",
		"maktbarometern-2025-x.csv",
		"maktbarometern-2025-tiktok.csv",
		"arets-makthavare-2025.csv",
	}

	for _, f := range files {
		err := formatter.CleanCSV(f)
		if err != nil {
			log.Printf("Failed to clean %s: %v", f, err)
		} else {
			log.Printf("Successfully cleaned %s", f)
		}
	}

	//https://medieakademin.se/maktbarometern/

	/*
	urls := [6]string{
		"https://medieakademin.se/maktbarometern-2025-facebook/", 
		"https://medieakademin.se/maktbarometern-2025-instagram/", 
		"https://medieakademin.se/maktbarometern-2025-youtube/", 
		"https://medieakademin.se/maktbarometern-2025-x/", 
		"https://medieakademin.se/maktbarometern-2025-tiktok/", 
		"https://medieakademin.se/arets-makthavare-2025/",
	}

	for _, url := range urls {
		err := func() error {
			parts := strings.Split(strings.Trim(url, "/"), "/")

			filepath := "data/" + parts[len(parts)-1] + ".csv"

			items, err := scraper.ScrapeList(url)

			if err != nil {
				return fmt.Errorf("Failed to scrape: %e", err)
			}

			file, err := os.Create(filepath)

			if err != nil {
				return fmt.Errorf("Failed to create file: %e", err)
			}

			defer file.Close()

			writer := csv.NewWriter(file)

			defer writer.Flush()

			writer.Write([]string{"rank", "name", "score"})

			for _, item := range items {
				writer.Write([]string{item.Rank, item.Name, item.Score})
			} 

			fmt.Printf("Saved %d rows to %s\n", len(items), filepath)

			return nil
		}()

		if err != nil {
			log.Println(err)
		}
    }
	*/
}