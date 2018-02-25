package main

import (
	"github.com/gin-gonic/gin"
	"github.com/6thplaneta/hermes"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	app := hermes.NewApp("./test/conf.yml")
	app.InitLogs(app.Conf.GetString("public.logs"))
	app.Mount(hermes.AuthorizationModule, "/auth")
	app.Run()
}
