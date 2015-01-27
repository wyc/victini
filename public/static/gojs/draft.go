package main

import "github.com/gopherjs/jquery"

//go:generate gopherjs build draft.go

//convenience:
var jQuery = jquery.NewJQuery

func main() {

	//show jQuery Version on console:
	print("Your current jQuery version is: " + jQuery().Jquery)
	print("asdf")

	spoiledCards := jQuery("div.spoiledCards").Children(".spoiledcard")
	undoButtons := jQuery(".pick-btn")
	spoiledCards.On(jquery.CLICK, func(e jquery.Event) {
		spoiledCard := jQuery(e.CurrentTarget)
		img := spoiledCard.Children("img")
		btn := spoiledCard.Children("button")
		if img.Is(":visible") {
			img.FadeOut(func() {
				btn.Show()
			})
		}
	})

	undoButtons.On(jquery.CLICK, func(e jquery.Event) {
        // TODO actually undo the pick
	})
}
