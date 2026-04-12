package formatter

import (
	"bufio"
	"encoding/csv"
	"os"
	"strings"
)

func CleanCSV(fileName string) error{
	filePath := "data/" + fileName

	file, err := os.Open(filePath)

	if err != nil {
		return err
	}

	defer file.Close()

	var dataLines []string

	scanner := bufio.NewScanner(file)

	isFirstLine := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		if isFirstLine {
			isFirstLine = false
			continue
		}

		if line != "" {
			dataLines = append(dataLines, line)
		}
	}

	outFile, err := os.Create("data/formatted_" + fileName)

	if err != nil {
		return err
	}

	defer outFile.Close()

	writer := csv.NewWriter(outFile)

	defer writer.Flush()

	writer.Write([]string{"rank", "name", "score"})

	for i := 0; i < len(dataLines); i += 3 {
		if i+2 < len(dataLines) {
			row := []string{
				dataLines[i],
				dataLines[i+1],
				dataLines[i+2],
			}

			writer.Write(row)
		}
	}

	return nil
}