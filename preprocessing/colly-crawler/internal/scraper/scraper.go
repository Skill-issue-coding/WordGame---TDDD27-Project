package scraper

import (
	"log"

	"github.com/gocolly/colly/v2"
)

type Item struct {
	Rank string
	Name string
	Score string
}

func ScrapeList(url string) ([]Item, error) {
	var items []Item
	c := newCollector()

	c.OnHTML("h2", func(e *colly.HTMLElement) {
		log.Print("found h2")
	})

	c.OnHTML("table tbody.ant-table-tbody tr.ant-table-row", func(e *colly.HTMLElement) {
		item := Item{
			Rank: e.ChildText("td:nth-child(1)"),
			Name: e.ChildText("td:nth-child(2) a p"),
			Score: e.ChildText("td:nth-child(3) span"),
		}

		items = append(items, item)
	})

	err := c.Visit(url)

	log.Print("in scrape file")

	return items, err
}

func newCollector() *colly.Collector {
	return colly.NewCollector(colly.AllowedDomains("medieakademin.se"))
}