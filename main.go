package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"labix.org/v2/mgo"
)

const (
	DB_URL  = "localhost"
	DB_NAME = "db"
)

var CookieStore = sessions.NewCookieStore([]byte(os.Getenv("COOKIE_SECRET")))

var DB *mgo.Database

func main() {
	session, err := mgo.Dial(DB_URL)
	if err != nil {
		log.Fatalf(`Could not connect to database (url="%s"): %s`, DB_URL, err)
	}

	DB = session.DB(DB_NAME)

	router := mux.NewRouter()
	router.StrictSlash(true)
	router.HandleFunc("/login", Login)
	router.HandleFunc("/logout", Logout)
	router.HandleFunc("/signup", Signup)
	router.Handle("/deck.json", DraftHandler(serveDeck))
	router.Handle("/gallery.json", DraftHandler(serveGallery))
	router.Handle("/card_pack_count.json", DraftHandler(serveCardPackCount))
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatalf("listening, %v", err)
	}
}
