package main

import (
	"labix.org/v2/mgo/bson"
)

const (
	DB_URL  = "localhost"
	DB_NAME = "db"
)

var DB *mgo.Database

func main() {
	session, err := mgo.Dial(DB_URL)
	if err != nil {
		log.Fatalf(`Could not connect to database (url="%s"): %s`, DB_URL, err)
	}

	DB = session.DB(DB_NAME)
}
