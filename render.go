package main

import (
	"log"
	"net/http"
	"text/template"

	"github.com/ChimeraCoder/godeckbrew"
)

type Page struct {
	DraftIdHex string
	Data       interface{}
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	t, err := template.ParseFiles("templates/"+tmpl+".tmpl", "templates/base.tmpl")
	if err != nil {
		log.Println("render:", err)
		return
	}
	t.ExecuteTemplate(w, "base", p)
}

func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	log.Println("Login page request from", r.RemoteAddr)
	renderTemplate(w, "login", &Page{})
}

func serveDraftPage(w http.ResponseWriter, r *http.Request) {
	log.Println("Draft page request from", r.RemoteAddr)
	set, err := godeckbrew.GetSet("KTK")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	renderTemplate(w, "draft", &Page{
		DraftIdHex: vars["DraftIdHex"],
		Data:       set.NewBoosterPack(),
	})
}

func serveFakeGallery(w http.ResponseWriter, r *http.Request) {
	set, err := godeckbrew.GetSet("KTK")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	serveJSON(w, set.NewBoosterPack())
}
