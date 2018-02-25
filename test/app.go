package main

import (
	"github.com/6thplaneta/hermes"
)

func main() {
	app := hermes.NewApp()
	app.Mount(hermes.AuthorizationModule, "/auth")
	app.Run()
}
