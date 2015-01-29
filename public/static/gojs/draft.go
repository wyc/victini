package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gopherjs/jquery"
)

//go:generate gopherjs build draft.go

//convenience:
var jQuery = jquery.NewJQuery

type CardPickRequest struct {
	CardID string
}

func pickCardDraft(e jquery.Event, draftIdHex string) {

	spoiledCards := jQuery("div.spoiledCards").Children(".spoiledcard")
	spoiledCards.Off(jquery.CLICK)
	spoiledCard := jQuery(e.CurrentTarget)
	preHidden := spoiledCard.Find(".pre-hidden")
	img := spoiledCard.Children("img")
	countdown := jQuery(spoiledCard).Find(".countdown-secs")
	selectedCardId := spoiledCard.Attr("id")

	countdownExpired := make(chan struct{})

	if img.Is(":visible") {
		img.FadeOut(func() {
			preHidden.Show()
			countdown.Show()
			go startCountdown(countdown, draftIdHex, spoiledCard.Attr("id"), countdownExpired)
		})
	}

	undone := make(chan struct{})

	undoButtons := jQuery(".pick-btn")
	undoButtons.On(jquery.CLICK, func(e jquery.Event) {
		// Extra goroutine needed for gopherjs due to javascript runtime
		go func() {
			undone <- struct{}{}
		}()
	})

	select {
	case <-countdownExpired:
		countdown.SetText("Card picked!")
		bts, err := json.Marshal(CardPickRequest{CardID: selectedCardId})
		if err != nil {
			print(fmt.Sprintf("Error marshalling data: %s", err))
		}
		queryURL := "/draft/" + draftIdHex + "/pick"
		jquery.Post(queryURL, string(bts), func(data interface{}) {
			print("got!")
			print(fmt.Sprintf("%+v", data))
		})

	case <-undone:
		preHidden.FadeOut(func() {
			countdown.FadeOut(func() {
				img.FadeIn()
			})
		})
		undoButtons.Off(jquery.CLICK)
		setCardPickEventHandlers(draftIdHex)

	}

}

func startCountdown(countdown jquery.JQuery, draftIdHex, selectedCardId string, countdownExpired chan struct{}) {
	counts := make(chan int)
	go func() {
		for i := 10; i > 0; i-- {
			counts <- i
			<-time.After(1 * time.Second)
		}
		close(counts)
	}()

	for {
		num, ok := <-counts
		if !ok {
			break
		}
		countdown.SetText(fmt.Sprintf("%d", num))
	}

	countdownExpired <- struct{}{}

}

func setCardPickEventHandlers(draftIdHex string) {
	spoiledCards := jQuery("div.spoiledCards").Children(".spoiledcard")
	spoiledCards.On(jquery.CLICK, func(e jquery.Event) {

		go pickCardDraft(e, draftIdHex)
	})
}

func main() {

	//show jQuery Version on console:
	print("Your current jQuery version is: " + jQuery().Jquery)

	draftIdHex := jQuery("span#draftIdHex").Text()

	setCardPickEventHandlers(draftIdHex)

}
