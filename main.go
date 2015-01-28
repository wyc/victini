package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"labix.org/v2/mgo"
)

//go:generate gopherjs build -o public/static/gojs/draft.js public/static/gojs/draft.go

var DB_URL = os.Getenv("MONGODB_PORT_27017_TCP_ADDR")

func init() {
	if DB_URL == "" {
		DB_URL = "127.0.0.1"
	}
}

const (
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
	router.HandleFunc("/login", serveLoginPage)
	router.HandleFunc("/draft/{DraftIdHex}", serveDraftPage)
	router.HandleFunc("/api/login", Login)
	router.HandleFunc("/api/logout", Logout)
	router.HandleFunc("/api/signup", Signup)
	router.Handle("/draft/{DraftIdHex}/deck.json", DraftHandler(serveDeck))
	router.Handle("/draft/{DraftIdHex}/gallery.json", DraftHandler(serveGallery))
	router.Handle("/draft/{DraftIdHex}/card_pack_count.json", DraftHandler(serveCardPackCount))
	router.Handle("/draft/{DraftIdHex}/pick", DraftHandler(servePick))
	router.HandleFunc("/draft/{DraftIdHex}/fake_gallery.json", serveFakeGallery)

	http.Handle("/static/", http.FileServer(http.Dir("public")))
	http.Handle("/", router)

	log.Println("Listening on", LISTEN)
	if err := http.ListenAndServe(LISTEN, nil); err != nil {
		log.Fatalf("listening, %v", err)
	}
}
