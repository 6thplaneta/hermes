package hermes

import (
	"net/http"
	"os/signal"
	"syscall"
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

	modules   []Moduler
	metaInfo  []ModuleInfo
	listeners []Listener
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

//
func (o *App) AddListener(listener Listener) {
	o.listeners = append(o.listeners, listener)
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

//
func (o *App) IsMaster() bool {
	return o.Conf.GetBool("is_master")
}

// Run ...
func (o *App) Run() {
	binding := o.Conf.GetString("router.bind-address")
	o.Router.Use(CORSMiddleware())
	go o.Router.Run(binding)
	o.listenForTerminate()
}

// utils

func (o *App) initLogs() {
	logger := logs.NewSimpleLogger()
	level := o.Conf.GetString("logs.level")
	switch level {
	case "off":
		logger.SetLevel(logs.Off)
	case "fatal":
		logger.SetLevel(logs.Fatal)
	case "error":
		logger.SetLevel(logs.Error)
	case "warning":
		logger.SetLevel(logs.Warning)
	case "info":
		logger.SetLevel(logs.Info)
	case "debug":
		logger.SetLevel(logs.Debug)
	case "trace":
		logger.SetLevel(logs.Trace)
	default:
		panic("this level of the log is not supported")
	}
	if o.Conf.GetBool("logs.stdout") {
		logger.Add(os.Stdout)
	}
	logger.SetDir(o.Conf.GetString("logs.path"))
	loc, err := time.LoadLocation(o.Conf.GetString("logs.location"))
	if err != nil {
		panic(err.Error())
	}
	logger.SetLocation(loc)
	logs.SetInstance(logger)
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

//
func (o *App) listenForTerminate() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGABRT, syscall.SIGSTOP, syscall.SIGPWR)
	sig := <-signals
	for _, l := range o.listeners {
		l.OnTerminate(sig)
	}
	if sig != nil {
		logs.Instance().Handle(logs.Off.New(sig.String()))
	}
}
