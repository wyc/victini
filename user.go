package main

import (
	"net/http"
	"time"

	"labix.org/v2/mgo/bson"
)

type User struct {
	Id           bson.ObjectId `bson:"_id"`
	CreatedOn    time.Time     `bson:"created_on"`
	LastActive   time.Time     `bson:"last_active"`
	Name         string        `bson:"name"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"password_hash"`
	Token        string        `bson:"token"`
}

func Login(w http.ResponseWriter, r *http.Request) (*User, error) {
	return nil, nil
}

func Signup(r *http.Request) (*User, error) {
	return nil, nil
}

func LoggedInUser(r *http.Request) (user *User, err error) {
	session, _ := CookieStore.Get(r, "session")
	token = session.Values["token"]
	user := &new(User)
	err = DB.C("Users").Find(bson.M{"token": token}).One(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u User) Logout() error {
	return nil
}
