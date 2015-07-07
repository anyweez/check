package main

import (
	"flag"
	"log"
)

// Stored state requirements
//   - {email address} => {config}

// Flags
var (
	CLIENT_ID   = flag.String("clientid", "", "OAuth 2.0 Client ID.")
	SECRET      = flag.String("secret", "", "OAuth 2.0 Client Secret.")
	CACHE_TOKEN = flag.Bool("cachetoken", true, "cache the OAuth 2.0 token")
)

func main() {
	flag.Parse()

	// Create a new server object that's properly configured.
	server := NewCoreServer()
	// Configure and start the Check web server.
	log.Println("Starting server...")
	server.Start()
	log.Println("Closing...")
}
