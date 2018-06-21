package main

import (
	"github.com/6thplaneta/hermes"
	_ "github.com/lib/pq"
)

func main() {
	app := hermes.NewApp()
	app.Mount(hermes.AuthorizationModule, "/auth")
	println(app.IsMaster())
	app.Run()
}
