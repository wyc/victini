package main

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
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

type LoginReq struct {
	Email    string
	Password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	return
}

func Signup(w http.ResponseWriter, r *http.Request) {
	passwordHash, err := GenerateFromPassword(req.Password, DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.ToLower(LoginReq.Email)
	count, err := DB.C("Users").Find(bson.M{"email": email}).Count()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if count > 0 {
		http.Error(w, "Account exists", http.StatusBadRequest)
		return
	}

	err = DB.C("Users").Insert(User{
		Id:           bson.NewObjectId(),
		CreatedOn:    time.Now(),
		LastActive:   time.Now(),
		Name:         "",
		Email:        email,
		PasswordHash: passwordHash,
		Token:        "",
	})
	if err != nil {
		http.Error(w, "Account exists", http.StatusBadRequest)
		return
	}

	serveJSON(w, nil)
}

func LoggedInUser(r *http.Request) (user *User, err error) {
	session, _ := CookieStore.Get(r, "session")
	token := session.Values["token"]
	user = new(User)
	err = DB.C("Users").Find(bson.M{"token": token}).One(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u User) Logout() error {
	return nil
}
