package hermes

import (
	"time"

	"github.com/6thplaneta/go-server/logs"
	//

	"net/http"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

var SystemToken = uuid.NewV4().String()

func init() {
	InitMessages()
}

type App struct {
	Router  *gin.Engine
	Conf    *viper.Viper
	DataSrc *DataSrc
	Modules []Moduler
}

func (app *App) config(path string) {
	app.Conf.SetConfigFile(path)
	err := app.Conf.ReadInConfig()
	if err != nil {
		panic(Messages["NotFoundConfig"])
	}
}

func NewApp(configPath string) *App {
	app := &App{}
	app.Router = gin.New()
	app.Router.Use(ginLogger(), gin.Recovery())
	app.Conf = viper.New()
	app.config(configPath)
	app.Modules = make([]Moduler, 0)
	app.Init()
	app.Router.GET("/meta", app.Meta)
	return app
}

func (app *App) Init() {
	datasrc := &DataSrc{}
	err := datasrc.Init(app.Conf)
	if err != nil {
		panic(err)
	}
	app.DataSrc = datasrc
	app.initLogs()
	go app.killTraper()
}

func (app *App) Meta(c *gin.Context) {
	mgs := make([]ModuleInfo, len(app.Modules))
	for i, mg := range app.Modules {
		mi := ModuleInfo{mg.GetName(), mg.GetPath(), "Module"}
		mgs[i] = mi
	}
	c.JSON(http.StatusOK, mgs)
}

func (app *App) initLogs() {
	var path interface{}
	path = app.Conf.GetStringMap("logs")["path"]
	if path != nil {
		logs.SetDir(path.(string))
	}
	level := app.Conf.GetStringMap("logs")["level"].(string)
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
	}
	if app.Conf.GetStringMap("logs")["stdout"].(bool) {
		logs.Add(os.Stdout)
	}
	loc, err := time.LoadLocation(app.Conf.GetStringMap("logs")["location"].(string))
	if err != nil {
		panic(err.Error())
	}
	logs.SetLocation(loc)
}

func (app *App) uninitDB() {
}

func (app *App) killTraper() {
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, os.Kill)
	<-sigchan
	app.uninitDB()
	//if app.Logger != nil {
	//	app.Logger.UninitLogs()
	//
	//}

	// do last actions and wait for all write operations to end

	os.Exit(0)
}

//
func (app *App) GetSettings(name string) Settings {
	settings := app.Conf.GetStringMap(name)
	if settings == nil {
		return Settings{}
	} else {
		pubs := app.Conf.GetStringMap("public")
		for k, v := range pubs {
			settings[k] = v
		}
		return settings
	}

}

func (app *App) Mount(mg Moduler, mountbase string) {
	app.Modules = append(app.Modules, mg)
	mg.SetMountPath(mountbase)
	mg.SetApp(app)
	mg.SetDataSrc(app.DataSrc)
	err := mg.Init(app)
	if err != nil {
		panic("mount error at: " + mountbase + " error message is: " + err.Error())
	}
	app.Router.GET(mountbase+"/meta", mg.Meta)

}

func (app *App) Run() {
	binding := app.Conf.GetString("router.bind-address")
	app.Router.Use(CORSMiddleware())
	app.Router.Run(binding)

}
