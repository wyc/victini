package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"log"

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

// Drafts returns all drafts the user has been associated with (whether currently active or not)
func (u User) Drafts() (d []*Draft, err error) {
	err = DB.C("Drafts").Find(bson.M{
		"players": bson.M{
			"$elemMatch": bson.M{
				"user_id": []bson.ObjectId{u.Id},
			},
		},
	}).All(&d)
	return
}

// ActiveDraft returns the one active draft for the user
// We assume that there is at most one active draft for each user at all times

func (u User) ActiveDraft() (active *Draft, err error) {
	drafts, err := u.Drafts()
	if err != nil {
		return
	}
	for _, draft := range drafts {
		if !draft.Finished() {
			return draft, nil
		}
	}
	return nil, err
}

func (user User) Save() error { return DB.C("Users").UpdateId(user.Id, user) }

type LoginReq struct {
	Email    string
	Password string
}

func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login request from", r.RemoteAddr)
	req := new(LoginReq)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.ToLower(req.Email)
	log.Println("Logging in as", email)
	user := new(User)
	err = DB.C("Users").Find(bson.M{"email": email}).One(user)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user.Token = uuid.New()
	session, _ := CookieStore.Get(r, "session")
	session.Values["token"] = user.Token
	session.Save(r, w)

	err = user.Save()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("Logged in", email)
	serveJSON(w, nil)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	log.Println("Signup request from", r.RemoteAddr)
	req := new(LoginReq)
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	email := strings.ToLower(req.Email)
	log.Println("Signing up", email)
	count, err := DB.C("Users").Find(bson.M{"email": email}).Count()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else if count > 0 {
		log.Println("Account already exists:", email)
		http.Error(w, "Account already exists", http.StatusBadRequest)
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
		log.Println(err)
		http.Error(w, "Database error", http.StatusBadRequest)
		return
	}

	log.Println("Signed up", email)
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
	log.Println("Logout request from", r.RemoteAddr)
	user, err := LoggedInUser(r)
	if err != nil {
		http.Error(w, "Not logged in", http.StatusBadRequest)
		return
	}

	user.Token = uuid.New()
	err = user.Save()
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	session, _ := CookieStore.Get(r, "session")
	delete(session.Values, "token")
	session.Save(r, w)
	log.Println("Logged out", user.Email)
}
