package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ChimeraCoder/godeckbrew"
)

// Gallery returns the current collection (multiset) of cards that the player
// can currently choose from
func (player Player) Gallery() CardPack {
	if len(player.CardPacks) == 0 {
		return CardPack(make([]*godeckbrew.Card, 0))
	}
	return player.CardPacks[0]
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

			if draft.AllPlayerCardsPicked() {
				log.Println("No more CardPacks going around...dealing!")
				err = draft.Deal()
				if err != nil {
					log.Println("Deal:", err)
				}
			} else if draft.AllCardsPicked() {
				draft.Finished = true
				// @TODO END DRAFT
			}
			return draft.Save()
		}
	}
	return fmt.Errorf("Card not found")
}
