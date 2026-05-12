package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strings"

	"server/words"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

// rankOf returns how many vocabulary words are MORE similar to the active word
// than the given guess is. Rank 1 = the target itself.
func rankOf(dict *words.Dictionary, guessDistance float64) int {
	activeKey := strings.ToLower(strings.TrimSpace(dict.ActiveWord))
	target, ok := dict.WordMap[activeKey]
	if !ok || len(target.WordVector) == 0 {
		return -1
	}
	guessSim := 1.0 - guessDistance
	rank := 1
	for _, entry := range dict.WordMap {
		if len(entry.WordVector) != len(target.WordVector) {
			continue
		}
		var dot float64
		for i, v := range target.WordVector {
			dot += float64(v) * float64(entry.WordVector[i])
		}
		if dot > guessSim {
			rank++
		}
	}
	return rank
}

func simColor(sim float64) string {
	switch {
	case sim >= 0.80:
		return colorGreen
	case sim >= 0.60:
		return colorYellow
	default:
		return colorRed
	}
}

func simBar(sim float64) string {
	const width = 20
	filled := int(sim * width)
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

func main() {
	fmt.Println("Laddar ordbok…")
	dict, err := words.InitializeDictionary()
	if err != nil {
		log.Fatalf("Kunde inte ladda ordbok: %v", err)
	}

	total := len(dict.WordMap)
	fmt.Printf("Laddade %d ord\n", total)

	if err := dict.SetRandomContextoTarget(); err != nil {
		log.Fatalf("Kunde inte välja målord: %v", err)
	}

	guesses := 0

	fmt.Printf("\n%s╔══════════════════════════╗%s\n", colorBold, colorReset)
	fmt.Printf("%s║        C O N T E X T O   ║%s\n", colorBold, colorReset)
	fmt.Printf("%s╚══════════════════════════╝%s\n\n", colorBold, colorReset)
	fmt.Printf("Gissa det hemliga ordet! (%d ord i ordlistan)\n", total)
	fmt.Printf("%sKommandon:%s new · reveal · exit\n\n", colorDim, colorReset)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Gissning #%d › ", guesses+1)
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "exit", "quit":
			fmt.Printf("\nSvaret var: %s%s%s\n", colorGreen+colorBold, dict.ActiveWord, colorReset)
			return

		case "reveal":
			fmt.Printf("  → Svaret är: %s%s%s\n\n", colorGreen+colorBold, dict.ActiveWord, colorReset)
			continue

		case "new":
			old := dict.ActiveWord
			if err := dict.SetRandomContextoTarget(); err != nil {
				fmt.Println("  Kunde inte välja nytt ord.")
				continue
			}
			guesses = 0
			fmt.Printf("  Förra ordet var: %s%s%s — nytt ord valt!\n\n", colorGreen+colorBold, old, colorReset)
			continue
		}

		// Win check (case-insensitive)
		if strings.EqualFold(input, dict.ActiveWord) {
			guesses++
			fmt.Printf("\n%s🎉  Rätt svar! Du hittade '%s' på %d gissning", colorGreen+colorBold, dict.ActiveWord, guesses)
			if guesses != 1 {
				fmt.Print("ar")
			}
			fmt.Printf("!%s\n\n", colorReset)
			fmt.Print("Nytt ord? (j/n) › ")
			if scanner.Scan() && strings.ToLower(strings.TrimSpace(scanner.Text())) == "j" {
				dict.SetRandomContextoTarget()
				guesses = 0
				fmt.Printf("  Nytt ord valt!\n\n")
			}
			continue
		}

		if !dict.IsValid(input) {
			fmt.Printf("  %s'%s' finns inte i ordlistan.%s\n", colorDim, input, colorReset)
			continue
		}

		distance := dict.CalculateDistance(input)
		if math.IsNaN(distance) {
			fmt.Printf("  Kunde inte beräkna avstånd för '%s'.\n", input)
			continue
		}

		guesses++
		sim := 1.0 - distance
		rank := rankOf(&dict, distance)
		col := simColor(sim)

		fmt.Printf("  %s%-22s%s  %s%s%s  sim %s%.4f%s  rank %s#%d / %d%s\n",
			colorBold, input, colorReset,
			col, simBar(sim), colorReset,
			col, sim, colorReset,
			col, rank, total, colorReset,
		)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Läsfel:", err)
	}
}
