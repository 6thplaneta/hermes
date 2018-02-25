package main

import (
	"github.com/6thplaneta/hermes"
)

func main() {
	app := hermes.NewApp("./conf.yml")
	app.Mount(hermes.AuthorizationModule, "/auth")
	app.Run()
}
