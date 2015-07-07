package main

import (
	"net/http"
	"time"
	"github.com/goincremental/negroni-sessions"
	oauth2 "github.com/goincremental/negroni-oauth2"
)

type UserConfig struct {
	// Hashed email address
	UserId				[]byte
	EmailAddress		string
	CreatedOn			time.Time

	// Filter strings are passed to Google to define what messages should
	// come back as tasks. This parameter stores the last user-provided
	// FilterString that returned a successful response code from Google.
	FilterString		string

	// User-provided function input that last executed successfully. Note
	// that this is user-provided and should never be passed to any other
	// users under any circumstances.
	RankingFunc			string
}

func GetUserFromSession(r *http.Request) (UserConfig, error) {
	session := sessions.GetSession(r)
	user := session.Get("config").(UserConfig)

	return user, nil
}

func StoreUserInSession(r *http.Request, uc UserConfig) error {
	session := sessions.GetSession(r)
	session.Set("config", uc)

	return nil
}

// TODO: implement
func GetUserFromStorage(email string) (UserConfig, error) {
	return UserConfig{}, nil
}

// TODO: implement
func StoreUserInStorage(uc UserConfig) error {
	return nil
}

func RemoveUserFromSession(r *http.Request) error {
	session := sessions.GetSession(r)
	session.Delete("config")

	return nil
}

func UserLoggedIn(r *http.Request) bool {
	token := oauth2.GetToken(r)
	return token != nil && token.Valid()
}