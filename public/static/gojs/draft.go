package main

import (
	"fmt"
	"time"

	"github.com/gopherjs/jquery"
)

//go:generate gopherjs build draft.go

//convenience:
var jQuery = jquery.NewJQuery

func startCountdown(countdown jquery.JQuery, selectedCardId string) {
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
	countdown.SetText("Card picked!")
	jquery.Post("/draft/1", selectedCardId, func(data interface{}) {
		print("got!")
		print(fmt.Sprintf("%+v", data))
	})
}

func main() {

	//show jQuery Version on console:
	print("Your current jQuery version is: " + jQuery().Jquery)

	spoiledCards := jQuery("div.spoiledCards").Children(".spoiledcard")
	undoButtons := jQuery(".pick-btn")
	spoiledCards.On(jquery.CLICK, func(e jquery.Event) {
		spoiledCard := jQuery(e.CurrentTarget)
		img := spoiledCard.Children("img")
		btn := spoiledCard.Find(".pick-btn")
		countdown := jQuery(spoiledCard).Find(".countdown-secs")
		if img.Is(":visible") {
			img.FadeOut(func() {
				btn.Show()
				countdown.Show()
				go startCountdown(countdown, spoiledCard.Attr("id"))
			})
		}
	})

	undoButtons.On(jquery.CLICK, func(e jquery.Event) {
		jQuery(".countdown-secs")
		// TODO actually undo the pick
	})
}
