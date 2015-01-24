package main

import (
	"net/http"
	"time"

	"labix.org/v2/mgo/bson"
)

type User struct {
	Id           bson.ObjectId
	CreatedOn    time.Time
	LastActive   time.Time
	Email        string
	PasswordHash string
	Token        string
}

func Login(w http.ResponseWriter, r *http.Request) (*User, error) {
	return nil, nil
}

func Signup(r *http.Request) (*User, error) {
	return nil, nil
}

func LoggedInUser(r *http.Request) (*User, error) {
	return nil, nil
}

func (u User) Logout() error {
	return nil
}
