package main

import "github.com/gopherjs/jquery"

//go:generate gopherjs build test.go

//convenience:
var jQuery = jquery.NewJQuery

const (
	INPUT  = "input#name"
	OUTPUT = "span#output"
)

func main() {

	//show jQuery Version on console:
	print("Your current jQuery version is: " + jQuery().Jquery)
	print("asdf")

	//catch keyup events on input#name element:
	jQuery(INPUT).On(jquery.KEYUP, func(e jquery.Event) {

		name := jQuery(e.Target).Val()
		name = jquery.Trim(name)
		print("name is" + name)

		//show welcome message:
		if len(name) > 0 {
			jQuery(OUTPUT).SetText("Log in as " + name + " !")
		} else {
			jQuery(OUTPUT).Empty()
		}
	})
}
