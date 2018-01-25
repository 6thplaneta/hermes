package hermes

import (
	//

	"os"
	"net/http"
	"os/signal"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/6thplaneta/u"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

var SystemToken = uuid.NewV4().String()

func init() {
	InitMessages()
}

type App struct {
	Router  *gin.Engine
	Conf    *viper.Viper
	DataSrc *DataSrc
	Logger  *u.Logger
	Modules []Moduler
}

func (app *App) config(path string) {
	app.Conf.SetConfigFile(path)
	err := app.Conf.ReadInConfig()
	if err != nil {
		panic(Messages["NotFoundConfig"])
	}
}

//func NewApp(configPath string) *App {
//	app := &App{}
//	app.Router = gin.Default()
//	app.Conf = viper.New()
//	app.config(configPath)
//	app.Modules = make([]Moduler, 0)
//	app.Init()
//	app.Router.GET("/meta", app.Meta)
//	return app
//}

func NewApp(configPath string, router *gin.Engine) *App {
	app := &App{}
	app.Router = router
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

func (app *App) InitLogs(path string) {
	var err error
	if app.Logger, err = u.NewLogger(path, 10000, u.Tehran, nil); err != nil {
		panic(err)
	}
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
	binding := app.Conf.GetString("app.bind-address")
	app.Router.Use(CORSMiddleware())
	app.Router.Run(binding)

}

func (app *App) RunTLS(certFile, keyFile string) {
	binding := app.Conf.GetString("app.bind-address")
	app.Router.Use(CORSMiddleware())
	app.Router.RunTLS(binding, certFile, keyFile)
}
