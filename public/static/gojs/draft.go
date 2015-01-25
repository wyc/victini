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

			name := jQuery(e.Target).Val()
			name = jquery.Trim(name)
			print(e.Target)
			jQuery(e.Target).Parent().ToggleClass("selected")

			/**
						//show welcome message:
						if len(name) > 0 {
							jQuery(OUTPUT).SetText("Log in as " + name + " !")
						} else {
							jQuery(OUTPUT).Empty()
						}
			            **/
		})
	}
}
