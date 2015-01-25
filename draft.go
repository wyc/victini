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
	Emails    []string      `bson:"emails"`
	Started   bool          `bson:"started"`
	Finished  bool          `bson:"finished"`
}

// For now, all drafts are KTK drafts
func (draft *Draft) MakePacks() error {
	numPlayers := len(draft.Players)
	const packsPerPlayer = 3
	set, err := godeckbrew.GetSet("KTK")
	if err != nil {
		return err
	}
	boosters := make([]CardPack, numPlayers*packsPerPlayer)
	for i, _ := range boosters {
		boosters[i] = CardPack(set.NewBoosterPack())
	}
	draft.CardPacks = []CardPack(boosters)
	return draft.Save()
}

func NewDraft(name string, emails []string) *Draft {
	return &Draft{
		Id:        bson.NewObjectId(),
		CreatedOn: time.Now(),
		Name:      name,
		CardPacks: make([]CardPack, 0),
		Players:   make([]Player, 0),
		Emails:    emails,
	}
}

type CreateDraftReq struct {
	Name   string
	Emails []string
}

func serveCreateDraft(w http.ResponseWriter, r *http.Request) error {
	log.Println("Create draft request from", r.RemoteAddr, ":", r.URL)
	user, err := LoggedInUser(r)
	if err != nil {
		log.Println("User not logged in")
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return nil
	}

	draft, err := user.ActiveDraft()
	if draft != nil {
		log.Println(user.Email, "already has an active draft")
		http.Error(w, "You already have an active draft!", http.StatusBadRequest)
		return nil
	} else if err != nil {
		return err
	}

	req := new(CreateDraftReq)
	err = json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return err
	}

	req.Emails = append(req.Emails, user.Email)

	draft = NewDraft(req.Name, req.Emails)
	err = draft.Insert()
	if err != nil {
		return err
	}

	serveJSON(w, nil)
	return nil
}

func serveJoinDraft(w http.ResponseWriter, r *http.Request) error {
	log.Println("Start draft request from", r.RemoteAddr, ":", r.URL)
	user, err := LoggedInUser(r)
	if err != nil {
		log.Println("User not logged in")
		http.Error(w, "Not logged in", http.StatusUnauthorized)
		return nil
	}

	draft, err := user.ActiveDraft()
	if draft != nil {
		log.Println(user.Email, "already has an active draft")
		http.Error(w, "You already have an active draft!", http.StatusBadRequest)
		return nil
	} else if err != nil {
		return err
	}

	draft, err = DraftFromURL(r)
	if err != nil {
		return err
	}

	if draft.Finished {
		return fmt.Errorf("Draft has already finished")
	} else if draft.Started {
		return fmt.Errorf("Draft has already started")
	}

	err = draft.AddPlayer(*user)
	if err != nil {
		return err
	}

	serveJSON(w, nil)
	return nil
}

func (draft Draft) AddPlayer(user User) error {
	// @TODO actually add player
	return draft.Save()
}

// true iff all players have chosen all their cards
func (d Draft) AllCardsPicked() bool {
	if len(d.CardPacks) > 0 {
		return false
	}

	for _, player := range d.Players {
		if len(player.CardPacks) != 0 {
			return false
		}
	}
	return true
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

func (draft Draft) Save() error   { return DB.C("Drafts").UpdateId(draft.Id, draft) }
func (draft Draft) Insert() error { return DB.C("Drafts").Insert(draft) }

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

func DraftFromURL(r *http.Request) (*Draft, error) {
	vars := mux.Vars(r)
	draftIdHex := vars["DraftIdHex"]
	if !bson.IsObjectIdHex(draftIdHex) {
		log.Printf(`Invalid Draft ID: "%s"`, draftIdHex)
		return nil, fmt.Errorf("Invalid Draft ID")
	}

	draftId := bson.ObjectIdHex(draftIdHex)
	var draft Draft
	if err := DB.C("Drafts").FindId(draftId).One(&draft); err != nil {
		log.Println(`Draft not found: "%s"`, draftId.Hex())
		return nil, fmt.Errorf("Draft not found")
	}

	return &draft, nil
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

	draft, err := DraftFromURL(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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

	err = dh(w, r, *draft, playerIdx)
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

			if draft.AllCardsPicked() {
				draft.Finished = true
				// @TODO END DRAFT
			}
			return draft.Save()
		}
	}
	return fmt.Errorf("Card not found")
}
