package main

import (
	"encoding/json"
	"net/http"
)

func serveJSON(w http.ResponseWriter, v interface{}) error {
	js, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

type Address struct {
	Street1 string `bson:"street1"`
	Street2 string `bson:"street2"`
	City    string `bson:"city"`
	State   string `bson:"state"`
	ZipCode string `bson:"zip_code"`
}
