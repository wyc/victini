package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go-uuid/uuid"
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

func (user User) Save() error { return DB.C("Users").UpdateId(user.Id, user) }

type LoginReq struct {
	Email    string
	Password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	req := new(LoginReq)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.ToLower(req.Email)
	user := new(User)
	err = DB.C("Users").Find(bson.M{"email": email}).One(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.Token = uuid.New()
	session, _ := CookieStore.Get(r, "session")
	session.Values["token"] = user.Token
	session.Save(r, w)

	err = user.Save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	serveJSON(w, nil)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	req := new(LoginReq)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.ToLower(req.Email)
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
		PasswordHash: string(passwordHash),
		Token:        uuid.New(),
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

func Logout(w http.ResponseWriter, r *http.Request) {
	user, err := LoggedInUser(r)
	if err != nil {
		http.Error(w, "Not logged in", http.StatusBadRequest)
		return
	}

	user.Token = uuid.New()
	err = user.Save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, _ := CookieStore.Get(r, "session")
	delete(session.Values, "token")
	session.Save(r, w)
}
