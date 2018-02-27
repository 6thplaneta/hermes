package hermes

import (
	"net/http"
	"time"

	"os"

	"github.com/6thplaneta/go-server/logs"
	"github.com/gin-gonic/gin"

	"github.com/spf13/viper"
)

// NewApp ...
func NewApp() *App {
	app := &App{}
	app.Conf = newViper()
	app.Router = newGinEngine()
	app.DataSrc = newDataSrc(app.Conf)
	app.Router.GET("/meta", app.meta)
	app.initLogs()
	return app
}

// App ...
type App struct {
	DataSrc *DataSrc
	Router  *gin.Engine
	Conf    *viper.Viper

	modules  []Moduler
	metaInfo []ModuleInfo
}

// GetSettings ...
// Deprecated
func (o *App) GetSettings(name string) Settings {
	settings := o.Conf.GetStringMap(name)
	if settings == nil {
		settings = Settings{}
	} else {
		pubs := o.Conf.GetStringMap("public")
		for k, v := range pubs {
			settings[k] = v
		}
	}
	return settings
}

// Mount ...
func (o *App) Mount(mg Moduler, mountbase string) {
	o.modules = append(o.modules, mg)
	mg.SetMountPath(mountbase)
	mg.SetApp(o)
	mg.SetDataSrc(o.DataSrc)
	err := mg.Init(o)
	if err != nil {
		panic("mount error at: " + mountbase + " error message is: " + err.Error())
	}
	o.Router.GET(mountbase+"/meta", mg.Meta)

}

// Run ...
func (o *App) Run() {
	binding := o.Conf.GetString("router.bind-address")
	o.Router.Use(CORSMiddleware())
	o.Router.Run(binding)
}

// utils

func (o *App) initLogs() {
	level := o.Conf.GetString("logs.level")
	switch level {
	case "off":
		logs.SetLevel(logs.Off)
	case "fatal":
		logs.SetLevel(logs.Fatal)
	case "error":
		logs.SetLevel(logs.Error)
	case "warning":
		logs.SetLevel(logs.Warning)
	case "info":
		logs.SetLevel(logs.Info)
	case "debug":
		logs.SetLevel(logs.Debug)
	case "trace":
		logs.SetLevel(logs.Trace)
	default:
		panic("this level of the log is not supported")
	}
	if o.Conf.GetBool("logs.stdout") {
		logs.Add(os.Stdout)
	}
	logs.SetDir(o.Conf.GetString("logs.path"))
	loc, err := time.LoadLocation(o.Conf.GetString("logs.location"))
	if err != nil {
		panic(err.Error())
	}
	logs.SetLocation(loc)
}

// handlers

func (o *App) meta(c *gin.Context) {
	if o.metaInfo == nil {
		for _, mg := range o.modules {
			o.metaInfo = append(o.metaInfo, ModuleInfo{mg.GetName(), mg.GetPath(), "Module"})
		}
	}
	c.JSON(http.StatusOK, o.metaInfo)
}
