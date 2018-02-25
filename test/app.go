package main

import (
	"github.com/6thplaneta/hermes"
	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	app := hermes.NewApp("./test/conf.yml", gin.Default())
	app.InitLogs(app.Conf.GetString("public.logs"))
	app.Mount(hermes.AuthorizationModule, "/auth")
	app.Run()
}
