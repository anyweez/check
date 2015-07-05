/**
 * Functions related to configuring or running the Check web server.
 */

package main

import (
	"net/http"
	"golang.org/x/net/context"
	gmail "google.golang.org/api/gmail/v1"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"encoding/json"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"time"
	"log"
	"strings"
)

// TODO: encrypt cookie data
var storage = sessions.NewCookieStore([]byte("cookie"))

var (
	AuthorizationCode = "authentication-code"
	AccessToken = "access-token"
	RefreshToken = "refresh-token"
)

func init() {
	storage.Options = &sessions.Options{
//		Domain: "",
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: false,
	}

	gob.Register(&oauth2.Token{})

	// may want to do something like this: gob.Register(&proto.User{})
}

func ConfigureAndStart() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/auth", auth_handler)

	// Log in / log out handlers
	rtr.HandleFunc("/login", login_handler)
	rtr.HandleFunc("/logout", logout_handler)

	// Core app handler
	rtr.HandleFunc("/app/fetch", fetch)
	rtr.HandleFunc("/app", app_handler)

	// Simple redirect handler 
	rtr.HandleFunc("^/$", main_handler)

	// Static content handler
	rtr.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("web/"))))

	http.Handle("/", rtr)
	http.ListenAndServe(":8080", nil)	
}

func auth_handler(w http.ResponseWriter, r *http.Request) {
		// requests for favicon.ico should be exluded
		if r.URL.Path == "/favicon.ico" {
			log.Println("favicon request doesn't count")
			http.Error(w, "", 404)
			return
		}

		// we want to get the authorization code. check to see if its there.
		if code := r.FormValue("code"); code != "" {

			access, refresh, err := FetchTokens(code)
			if err != nil {
				log.Println("Error retrieving tokens: " + err.Error())
			}
			// Save authorization code.
			session, err := storage.Get(r, "userSession")
			if err != nil {
				log.Println(err.Error())
			}

			if access != nil {
				session.Values[AccessToken] = access				
			}
			if refresh != nil {
				session.Values[RefreshToken] = refresh				
			}

			fmt.Println(session)
			err = session.Save(r, w)
			
			if err != nil {
				log.Println("Error saving session: " + err.Error())
			}
		
			http.Redirect(w, r, "/app", 307)
		} else {
			log.Printf("no code")
			http.Error(w, "", 500)
		}
}

func main_handler(w http.ResponseWriter, r *http.Request) {
	log.Println("main handler invoked")
	// redirect to the app handler until i actually build serverside support
	w.Write([]byte("go to /login"))
}

func login_handler(w http.ResponseWriter, r *http.Request) {
	// Either prompt the user to allow us to read their data, or retrieve a
	// cached refresh token (if available).
	redirectUrl, shouldRedirect := GetUserPermission()

	if shouldRedirect {
		http.Redirect(w, r, redirectUrl, 307)
	} else {
		// Look up user config info and store it in session.
		// TODO: create user sessions

		// Redirect to /app
//		http.Redirect(w, r, "/app", 307)		
		http.Error(w, "", 500)
	}
}

func logout_handler(w http.ResponseWriter, r *http.Request) {

}

func fetch(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")

	if len(query) > 0 {
		log.Println("Fetch request received. Query=" + query)
	} else {
		log.Println("Fetch request received. Query=<none>")		
	}

	tasks := make([]Item, 0)

	// Fetch the authorization code. We'll use it to get an access
	// token and (potentially) a refresh token.
	session, err := storage.Get(r, "userSession")
	if err != nil {
		log.Println("WARN: session decoding error: " + err.Error())
	}

	if token, ok := session.Values[AccessToken].(*oauth2.Token); !ok {
		log.Println("ERR: no retrievable token for this user")
		http.Error(w, "", 500)
	} else {
		config := oauth2.Config{
			ClientID: *CLIENT_ID,
			ClientSecret: *SECRET,
			Endpoint: google.Endpoint,
			Scopes: []string{gmail.MailGoogleComScope},
			RedirectURL: "http://localhost:8080/auth",
		}

		ctx := context.Background()
		client := config.Client(ctx, token)

		service, err := gmail.New(client)
		if err != nil {
			log.Println(err.Error())
		}

		// Get labels and create a mapping between ID's and names. Note that a label
		// on a message will be dropped unless it's mapping is available in this table.
		labelTable := make(map[string]string)
		labelReq := service.Users.Labels.List("me")
		labelResponse, err := labelReq.Do()
		if err != nil {
			log.Println(err.Error())
		}

		for _, label := range labelResponse.Labels {
			// Only keep user-defined labels. Gmail also creates some as well
			// but users don't see them so it's going to look whacky / broken here.
			if label.Type == "user" {
				labelTable[label.Id] = label.Name
			}
		}

		// Get messages
		messageReq := new(gmail.UsersMessagesListCall)
		if (len(query) > 0) {
			messageReq = service.Users.Messages.List("me").Q(query)
		} else {
			messageReq = service.Users.Messages.List("me")
		}
		messageResponse, err := messageReq.Do()

		if err != nil {
			log.Println(err.Error())
		} else {
			// TODO: if the query succeeded, save this as the user's filter.
		}

		log.Println(fmt.Sprintf("Retrieving %d messages...", len(messageResponse.Messages)))
		for _, m := range messageResponse.Messages {
			msg, err := service.Users.Messages.Get("me", m.Id).Do()
			if err != nil {
				log.Println(err.Error())
			}

			subject := ""
			from := ""
			date := time.Unix(0, 0)
			labels := make([]string, 0)

			// Extract important headers that we want to send to the frontend.
			for _, header := range msg.Payload.Headers {
				if header.Name == "Subject" {
					subject = header.Value
				}

				if header.Name == "From" {
					from = header.Value
				}

				if header.Name == "Received" {
					segments := strings.Split(header.Value, ";")
					date, err = time.Parse("Mon, 2 Jan 2006 15:04:05 -0700 (MST)", strings.TrimSpace(segments[1]))

					if err != nil {
						log.Println("Unknown date string: " + segments[1])
					}
				}
			}

			for _, label := range msg.LabelIds {
				if named, ok := labelTable[label]; ok {
					labels = append(labels, named)
				}
			}

			tasks = append(tasks, Item{
				Title: subject,
				CreationDate: date,
				Snippet: msg.Snippet,
				RequestedFrom: from,
				Labels: labels,
			})
		}
	}

	data, err := json.Marshal(tasks)

	if err != nil {
		log.Println("Couldn't serialize item list: " + err.Error())
	}
	
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(data))
}

func app_handler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadFile("web/app.html")

	if err != nil {
		log.Println("Couldn't find web/app.html")
		fmt.Fprintf(w, "Couldn't find web/app.html")
	} else {
		fmt.Fprintf(w, string(body))
	}
}