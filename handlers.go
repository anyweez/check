package main

import (
	"encoding/json"
	"fmt"
	oauth2 "github.com/goincremental/negroni-oauth2"
	"golang.org/x/net/context"
	goauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

/**
 * Presents the index page for users, which links to /login.
 */
func main_handler(w http.ResponseWriter, r *http.Request) {
	// If logged in, forward along to the main app directly.
	if UserLoggedIn(r) {
		http.Redirect(w, r, "/app", 307)
		// If not logged in, display the index page.
	} else {
		body, err := ioutil.ReadFile("web/index.html")

		if err != nil {
			log.Println("Couldn't find web/index.html")
			fmt.Fprintf(w, "Couldn't find index.html")
		} else {
			fmt.Fprintf(w, string(body))
		}
	}
}

/**
 * Receives the OAuth callback from Google, retrieves the one-time use authentication
 * code that can be used to fetch tokens.
 *
 * TODO: Also needs to call Google data API to retrieve the user's email address, which
 * we'll use to identify them (both internally and in the UI).
 */
func auth_handler(w http.ResponseWriter, r *http.Request) {
	// requests for favicon.ico should be excluded
	if r.URL.Path == "/favicon.ico" {
		log.Println("favicon request doesn't count")
		http.Error(w, "", 404)
		return
	}

	// We want to get the authorization code. Check to see if its there.
	user, err := FetchUser(r, w)
	if err != nil {
		log.Println("Couldn't retrieve user: " + err.Error())
		// On success, redirect to the main app
	} else {
		StoreUserInSession(r, user)
		http.Redirect(w, r, "/app", 307)
	}
}

func fetch(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("q")
	tasks := make([]Item, 0)

	if len(query) > 0 {
		log.Println("Fetch request received. Query=" + query)
	} else {
		log.Println("Fetch request received. Query=<none>")
	}

	tokens := oauth2.GetToken(r)
	// Retrieve the access token if it's available. If not then
	// its an error.
	if !UserLoggedIn(r) {
		log.Println("ERR: no retrievable token for this user")
		http.Error(w, "", 500)
	} else {
		service, err := getGmailService(tokens)
		if err != nil {
			http.Error(w, "", 500)
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
		if len(query) > 0 {
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

			tasks = append(tasks, parseMessage(msg, labelTable))
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

func getGmailService(tokens oauth2.Tokens) (*gmail.Service, error) {
	config := goauth2.Config{
		ClientID:     *CLIENT_ID,
		ClientSecret: *SECRET,
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.MailGoogleComScope},
		RedirectURL:  "http://localhost:8080/auth",
	}

	ctx := context.Background()
	tok := (goauth2.Token)(tokens.Get())
	client := config.Client(ctx, &tok)

	return gmail.New(client)
}

func parseMessage(msg *gmail.Message, labelTable map[string]string) Item {
	item := Item{
		Snippet: msg.Snippet,
	}

	// Extract important headers that we want to send to the frontend.
	for _, header := range msg.Payload.Headers {
		if header.Name == "Subject" {
			item.Title = header.Value
		}

		if header.Name == "From" {
			item.RequestedFrom = header.Value
		}

		if header.Name == "Received" {
			segments := strings.Split(header.Value, ";")
			date, err := time.Parse("Mon, 2 Jan 2006 15:04:05 -0700 (MST)", strings.TrimSpace(segments[1]))

			if err != nil {
				log.Println("Unknown date string: " + segments[1])
			} else {
				item.CreationDate = date
			}
		}
	}

	labels := make([]string, 0)
	for _, label := range msg.LabelIds {
		if named, ok := labelTable[label]; ok {
			labels = append(labels, named)
		}
	}

	item.Labels = labels

	return item
}
