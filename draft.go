package main

import (
	"time"

	"gopkg.in/mgo.v2"

	"github.com/ChimeraCoder/godeckbrew"
)

type Player struct {
	Id     bson.ObjectId
	UserId bson.UserId // A User has many Players
	Cards  []godeckbrew.Card
}

type CardPack struct {
	Name  string
	Cards []godeckbrew.Card
}

type Draft struct {
	Id        bson.ObjectId
	CreatedOn time.Time
	Name      string
	CardPacks []CardPack
	Players   []Player
}
