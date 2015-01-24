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
	DB_NAME = "draft"
	LISTEN  = ":8000"
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
	router.HandleFunc("/api/login", Login)
	router.HandleFunc("/api/logout", Logout)
	router.HandleFunc("/api/signup", Signup)
	router.Handle("/draft/{DraftIdHex}/deck.json", DraftHandler(serveDeck))
	router.Handle("/draft/{DraftIdHex}/gallery.json", DraftHandler(serveGallery))
	router.Handle("/draft/{DraftIdHex}/card_pack_count.json", DraftHandler(serveCardPackCount))

	log.Println("Listening on", LISTEN)
	if err := http.ListenAndServe(LISTEN, router); err != nil {
		log.Fatalf("listening, %v", err)
	}
}
