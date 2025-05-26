package models

import "time"

type Priority int

const (
	LowPriority Priority = iota
	MediumPriority
	HighPriority
)

type URL struct { //url to be crawled with metadata
	URL      string
	Priority Priority
	Depth    int
	Domain   string
}

type CrawledContent struct {
	URL       string
	Title     string
	Content   string
	Links     []string
	Hash      string
	Timestamp time.Time
}
