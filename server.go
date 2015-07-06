/**
 * Functions related to configuring or running the Check web server.
 */

package main

import (
	"net/http"
	"github.com/gorilla/mux"
//	"github.com/gorilla/sessions"
	oauth2 "github.com/goincremental/negroni-oauth2"
	"github.com/codegangsta/negroni"
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"encoding/gob"
)

/**
 * Some type definitions and object instantiations that are important
 * to make the server work as expected.
 */
type CoreServer struct {
	Router		*mux.Router
	Middleware  *negroni.Negroni  
}

// TODO: encrypt cookie data
//var storage = sessions.NewCookieStore([]byte("cookie"))

var (
	AuthorizationCode = "authentication-code"
	AccessToken = "access-token"
	RefreshToken = "refresh-token"
)

func init() {
	gob.Register(&UserConfig{})
}

/**
 * Constructs a new server object.
 */
func NewCoreServer() CoreServer {
	server := CoreServer{
		Router: mux.NewRouter(),
		Middleware: negroni.New(),
	}

	/**
	 * Add some Negroni middleware.
	 */
	server.Middleware.Use(negroni.NewLogger())
	// TODO: Need to change key.
	storage := cookiestore.New([]byte("temporary"))
	server.Middleware.Use(sessions.Sessions("userprofile", storage))
	config := (oauth2.Config)(GetClientConfig())
	server.Middleware.Use(oauth2.Google(&config ))

	/**
	 * Mux describing routes that require the user to be logged in via
	 * oauth first.
	 */
	secureMux := http.NewServeMux()
	// Core app handlers; these require the user to be logged in.
	secureMux.HandleFunc("/app/fetch", fetch)
	secureMux.HandleFunc("/app", app_handler)
	
	secure := negroni.New()
	secure.Use(oauth2.LoginRequired())
	secure.UseHandler(secureMux)

	/**
	 * Handlers that don't require authentication.
	 */
	server.Router.HandleFunc("/auth", auth_handler)
	// Static content handler
	server.Router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/"))))
	// Simple redirect handler 
	server.Router.HandleFunc("/", main_handler)

	/**
	 * And now connect everything together and call it a day.
	 */
	// Make sure the core router knows to let the secure router handle these routes.
	server.Router.Handle("/app/fetch", secure)
	server.Router.Handle("/app", secure)
	// Set negroni handler
	server.Middleware.UseHandler(server.Router)

	return server
}

func (cs *CoreServer) Start() {
	cs.Middleware.Run(":8080")
}