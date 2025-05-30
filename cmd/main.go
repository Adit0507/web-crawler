package main

import (
	"log"
	"webcrawler/internal/crawler"
	"webcrawler/internal/models"
)

func main() {
	allowedDomains := []string{"example.com", "test.com", "github.com"}

	useBloomFilter := true

	webcrawler := crawler.NewWebCrawler(allowedDomains, 3, useBloomFilter)

	err := webcrawler.AddSeedURL("https://github.com", models.HighPriority)
	if err != nil {
		log.Fatalf("Error adding seed URL: %v", err)
	}


	log.Println("Starting web crawler with robots.txt compliance and bloom filter...")
    webcrawler.Crawl()
    log.Println("Crawling finished!")
}
