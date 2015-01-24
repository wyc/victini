package main

import (
	"labix.org/v2/mgo/bson"
)

const DB_URL = "localhost"

func ColUsers() *mgo.Collection { DB.find }
