package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/ChimeraCoder/godeckbrew"
	"github.com/gorilla/mux"
)

type Player struct {
	Id        bson.ObjectId     `bson:"_id"`
	UserId    bson.ObjectId     `bson:"user_id"`
	Cards     []godeckbrew.Card `bson:"cards"`
	CardPacks []CardPack        `bson:"card_packs"`
	Position  int               `bson:"position"`
}

// Gallery returns the current collection (multiset) of cards that the player
// can currently choose from
func (player Player) Gallery() []godeckbrew.Card {
	if len(player.CardPacks) == 0 {
		return make([]godeckbrew.Card, 0)
	}
	return player.CardPacks[0].Cards
}

type CardPack godeckbrew.Set

type Draft struct {
	Id        bson.ObjectId `bson:"_id"`
	CreatedOn time.Time     `bson:"created_on"`
	Name      string        `bson:"name"`
	CardPacks []CardPack    `bson:"card_packs"`
	Players   []Player      `bson:"players"`
}

func (draft Draft) PlayerAfter(player Player) Player {
	nextPosition := (player.Position + 1) % len(draft.Players)
	for _, np := range draft.Players {
		if np.Position == nextPosition {
			return np
		}
	}
	log.Println("No next player found, returning current player")
	return player
}

func (draft Draft) Save() error { return DB.C("Drafts").UpdateId(draft.Id, draft) }

func (draft Draft) Deal() error {
	if len(draft.CardPacks) < len(draft.Players) {
		return fmt.Errorf("%d card packs is not enough for %d players",
			len(draft.CardPacks), len(draft.Players))
	}
	var i int
	for i = 0; i < len(draft.Players); i++ {
		draft.Players[i].CardPacks = append(draft.Players[i].CardPacks, draft.CardPacks[i])
	}
	draft.CardPacks = draft.CardPacks[i:]
	return draft.Save()
}

// DraftHandler is a wrapper that injects a special Draft handler func with
// a draft and the player index making the action
type DraftHandler func(w http.ResponseWriter, r *http.Request, draft Draft, playerIdx int) error

func (dh DraftHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := LoggedInUser(r)
	if err != nil {
		log.Println("User not logged in:", r.RemoteAddr)
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	draftIdHex := vars["DraftIdHex"]
	if !bson.IsObjectIdHex(draftIdHex) {
		log.Printf(`Invalid Draft ID: "%s"`, draftIdHex)
		http.Error(w, "Invalid Draft ID", http.StatusBadRequest)
		return
	}

	draftId := bson.ObjectIdHex(draftIdHex)
	var draft Draft
	if err := DB.C("Drafts").FindId(draftId).One(&draft); err != nil {
		log.Println(`Draft not found: "%s"`, draftId.Hex())
		http.Error(w, "Draft not found", http.StatusBadRequest)
		return
	}

	playerIdx := -1
	for idx, player := range draft.Players {
		if player.UserId == user.Id {
			playerIdx = idx
			break
		}
	}
	if playerIdx == -1 {
		log.Println(`Player "%s" not part of draft "%s"`, user.Name, draft.Id.Hex())
		http.Error(w, "You are not part of this draft", http.StatusUnauthorized)
		return
	}

	err = dh(w, r, draft, playerIdx)
	if err != nil {
		log.Println("DraftHandler call:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func serveDeck(w http.ResponseWriter, r *http.Request, draft Draft, pIdx int) error {
	serveJSON(w, draft.Players[pIdx].Cards)
	return nil
}

func serveGallery(w http.ResponseWriter, r *http.Request, draft Draft, pIdx int) error {
	serveJSON(w, draft.Players[pIdx].Gallery())
	return nil
}

type PickReq struct {
	CardID int
}

func serveCardPackCount(w http.ResponseWriter, r *http.Request, draft Draft, pIdx int) error {
	serveJSON(w, len(draft.Players[pIdx].CardPacks))
	return nil
}

// Player uses servePick to pick a card from the current Gallery (available
// cards to be picked). If the pick is successful, then the card pack is moved
// to the next player. The card pack is destroyed instead if it has no more cards
// after the pick.
func servePick(w http.ResponseWriter, r *http.Request, draft Draft, pIdx int) error {
	req := new(PickReq)
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Unmarshal PickReq:", err)
		// @TODO http.StatusBadRequest
		return fmt.Errorf("Could not parse request")
	}

	if len(draft.Players[pIdx].CardPacks) == 0 {
		return fmt.Errorf("Invalid card pick")
	}
	for i, card := range draft.Players[pIdx].CardPacks[0].Cards {
		if req.CardID == card.Multiverseid {
			// Add Card to Player Deck
			draft.Players[pIdx].Cards = append(draft.Players[pIdx].Cards, card)

			// Delete Card from Card Pack and move to next player
			cp := draft.Players[pIdx].CardPacks
			cp[0].Cards = append(cp[0].Cards[i:], cp[0].Cards[i+1:]...)
			draft.Players[pIdx].CardPacks = cp[1:]

			if len(cp[0].Cards) > 0 {
				nextPlayer := draft.PlayerAfter(draft.Players[pIdx])
				nextPlayer.CardPacks = append(nextPlayer.CardPacks, cp[0])
			}

			return draft.Save()
		}
	}
	return fmt.Errorf("Card not found")
}
