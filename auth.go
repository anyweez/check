package main

import (
	"golang.org/x/oauth2"
	gmail "google.golang.org/api/gmail/v1"
	"net/http"
)

func FetchUser(r *http.Request, w http.ResponseWriter) (UserConfig, error) {
	return UserConfig{}, nil
}

//func FetchEmail() {
//
//}

func GetClientConfig() oauth2.Config {
	return oauth2.Config{
		ClientID:     *CLIENT_ID,
		ClientSecret: *SECRET,
		RedirectURL:  "http://localhost:8080/oauth2callback",
		Scopes:       []string{gmail.MailGoogleComScope},
	}
}
