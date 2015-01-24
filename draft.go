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
func (player Player) Gallery() CardPack {
	if len(player.CardPacks) == 0 {
		return CardPack(make([]*godeckbrew.Card, 0))
	}
	return player.CardPacks[0]
}

type CardPack []*godeckbrew.Card

func (p CardPack) Value() (godeckbrew.Cents, error) {
	var value godeckbrew.Cents = 0
	for _, card := range []*godeckbrew.Card(p) {
		p, err := card.Price()
		if err != nil {
			return -1, fmt.Errorf("Error retriving price for card %s: %s", card.Name, err)
		}
		value += p
	}
	return value, nil
}

type Draft struct {
	Id        bson.ObjectId `bson:"_id"`
	CreatedOn time.Time     `bson:"created_on"`
	Name      string        `bson:"name"`
	CardPacks []CardPack    `bson:"card_packs"`
	Players   []Player      `bson:"players"`
}

// For now, all drafts are KTK drafts

func NewDraft(numPlayers int) (*Draft, error) {
	const packsPerPlayer = 3
	set, err := godeckbrew.GetSet("KTK")
	if err != nil {
		return &Draft{}, err
	}
	boosters := make([]CardPack, numPlayers*packsPerPlayer)
	for i, _ := range boosters {
		boosters[i] = CardPack(set.NewBoosterPack())
	}
	draft := Draft{}
	draft.CardPacks = []CardPack(boosters)
	return &draft, nil
}

// Finished is true iff all players have chosen all their cards
func (d Draft) Finished() bool {
	return len(d.CardPacks) == 0
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
	log.Println("Request from", r.RemoteAddr, ":", r.URL)
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
	for i, card := range draft.Players[pIdx].CardPacks[0] {
		if req.CardID == card.Multiverseid {
			// Add Card to Player Deck
			draft.Players[pIdx].Cards = append(draft.Players[pIdx].Cards, *card)

			// Delete Card from Card Pack and move to next player
			cp := draft.Players[pIdx].CardPacks
			cp[0] = append(cp[0][i:], cp[0][i+1:]...)
			draft.Players[pIdx].CardPacks = cp[1:]

			if len(cp[0]) > 0 {
				nextPlayer := draft.PlayerAfter(draft.Players[pIdx])
				nextPlayer.CardPacks = append(nextPlayer.CardPacks, cp[0])
			}

			return draft.Save()
		}
	}
	return fmt.Errorf("Card not found")
}
