package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

/**
 * For first-time users, asks for permission to access their data and
 * performs an oauth exchange with Google if accepted. For returning
 * users, looks up their refresh token.
 */
func GetUserPermission() (string, bool) {
	// TODO: check to see if we have a token cached already
	// Set up the application configuration.
	config := oauth2.Config{
		ClientID: *CLIENT_ID,
		ClientSecret: *SECRET,
		Endpoint: google.Endpoint,
		Scopes: []string{gmail.MailGoogleComScope},
		RedirectURL: "http://localhost:8080/auth",
	}

	randState := "188"

	// TODO: remove ApprovalForce
	return config.AuthCodeURL(randState, oauth2.ApprovalForce), true
}

func FetchTokens(authorization_code string) (*oauth2.Token, *oauth2.Token, error) {
	// TODO: dedup with definition above
	config := oauth2.Config{
		ClientID: *CLIENT_ID,
		ClientSecret: *SECRET,
		Endpoint: google.Endpoint,
		Scopes: []string{gmail.MailGoogleComScope},
		RedirectURL: "http://localhost:8080/auth",
	}

	ctx := context.Background()

	token, err := config.Exchange(ctx, authorization_code)
	
	if err != nil {
		return nil, nil, err
	}

	return token, nil, nil
}