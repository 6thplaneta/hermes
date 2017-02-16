package hermes

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var SystemToken string

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	SystemToken = randStringBytes(25)
	InitMessages()
}

type App struct {
	Router  *gin.Engine
	Conf    *viper.Viper
	DataSrc *DataSrc
	Logger  *Logger
	Modules []Moduler
}

func (app *App) config(path string) {
	app.Conf.SetConfigFile(path)
	// app.Conf.SetConfigName("conf")
	// app.Conf.AddConfigPath(".")
	// app.Conf.SetConfigType("yml")
	err := app.Conf.ReadInConfig()
	if err != nil {
		panic("could not configure app")
	}
	// app.Conf.Debug()

}

func NewApp(configPath string) *App {
	app := &App{}
	app.Router = gin.Default()
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
	app.Logger = &Logger{}
	app.Logger.InitLogs(path)
}

func (app *App) uninitDB() {
}

func (app *App) killTraper() {
	sigchan := make(chan os.Signal, 10)
	signal.Notify(sigchan, os.Interrupt)
	signal.Notify(sigchan, os.Kill)
	<-sigchan
	app.uninitDB()
	if app.Logger != nil {
		app.Logger.UninitLogs()

	}

	// do last actions and wait for all write operations to end

	os.Exit(0)
}

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
	fmt.Println("mg ", mg)
	err := mg.Init(app)
	if err != nil {
		panic("mount error at: " + mountbase + " error message is: " + err.Error())
	}
	app.Router.GET(mountbase+"/meta", mg.Meta)

}

func (app *App) Run() {
	binding := app.Conf.GetString("app.Bind-Address")
	app.Router.Run(binding)
}
