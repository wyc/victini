package main

import "github.com/gopherjs/jquery"

//go:generate gopherjs build draft.go

//convenience:
var jQuery = jquery.NewJQuery

func main() {

	// TODO don't hardcode these
	var inputs = []string{"a1", "a2", "a3"}

	//show jQuery Version on console:
	print("Your current jQuery version is: " + jQuery().Jquery)
	print("asdf")

	for _, cardId := range inputs {
		selector := "div#" + cardId
		//catch keyup events on input#name element:
		jQuery(selector).On(jquery.CLICK, func(e jquery.Event) {

			print(e.Target)
			img := jQuery(selector).Children("img")
			btn := jQuery(selector).Children("button")
			print(img)
            if img.Is(":visible") {
                img.FadeOut(func(){
                    btn.Show()
                })
            }
		})
	}
}
