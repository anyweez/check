package main

import (
	"github.com/braintree/manners"
	"testing"
	"time"
)

// Intended to be run as a goroutine.
func StopSoon(t time.Duration, done chan bool) {
	time.Sleep(t * time.Second)
	manners.Close()

	done <- true
}

// Check to ensure that the server loads at all. Don't worry about whether it
// can respond to any requests for this test.
func TestLaunchServer(t *testing.T) {
	server.Request()
	server.Start(2)
}

////////
var server ServerRequest

type ServerRequest struct {
	Ready  chan bool
	Server CoreServer
}

func init() {
	server.Ready = make(chan bool, 1)
	server.Ready <- true
}

func (ts *ServerRequest) Request() {
	<-ts.Ready
	ts.Server = NewCoreServer()
}

// Run the server for X seconds and release it back.
func (ts *ServerRequest) Start(t time.Duration) {
	go StopSoon(t, ts.Ready)
	manners.ListenAndServe(":8080", ts.Server.Middleware)
}
