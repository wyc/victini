package main

import (
	"log"
	"net/http"
	"text/template"
)

type Page struct{}

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
