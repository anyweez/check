package main

import (
	"time"
)

type Item struct {
	Title         string
	CreationDate  time.Time
	Labels        []string
	Snippet       string
	RequestedFrom string
}
