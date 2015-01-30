package main

import (
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/ChimeraCoder/godeckbrew"
	"github.com/gorilla/mux"
)

type Page struct {
	DraftIdHex string
	DraftIdInt int
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

	idInt, err := strconv.Atoi(vars["DraftIdHex"])
	if err != nil {
		idInt = 1
	}
	renderTemplate(w, "draft", &Page{
		DraftIdHex: vars["DraftIdHex"],
		DraftIdInt: idInt,
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
