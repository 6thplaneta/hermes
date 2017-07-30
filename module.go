package hermes

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Moduler interface {
	Init(*App) error
	SetMountPath(string)
	SetApp(*App)
	GetDataSrc() *DataSrc
	SetDataSrc(*DataSrc)
	NewController(Collectionist, string) *Controller
	RegisterController(Controlist)
	Mount(Moduler, string)
	GetName() string
	GetPath() string
	Meta(c *gin.Context)
	SetCrudRoutes(Controlist, []string)
	GET(string, gin.HandlerFunc) gin.IRoutes
	POST(string, gin.HandlerFunc) gin.IRoutes
	PUT(string, gin.HandlerFunc) gin.IRoutes
	DELETE(string, gin.HandlerFunc) gin.IRoutes
	HEAD(string, gin.HandlerFunc) gin.IRoutes
	PATCH(string, gin.HandlerFunc) gin.IRoutes
	OPTIONS(string, gin.HandlerFunc) gin.IRoutes
}

type Module struct {
	Name            string `json:"name"`
	MountPath       string
	HttpControllers []Controlist `json:"controllers"`
	Modules         []Moduler    `json:"modules"`
	App             *App
	DataSrc         *DataSrc
}

func (mg *Module) SetMountPath(path string) {

	mg.MountPath = path
}

func (mg *Module) SetApp(app *App) {
	mg.App = app
}

func (mg *Module) GetDataSrc() *DataSrc {
	return mg.DataSrc
}
func (mg *Module) SetDataSrc(src *DataSrc) {
	mg.DataSrc = src
}

func (mg *Module) Init(app *App) error {
	return nil
}

func NewModule(name string) *Module {
	m := &Module{}
	m.Name = name
	return m
}

func (mg *Module) GetName() string {
	return mg.Name
}

func (mg *Module) GetPath() string {
	return mg.MountPath
}

func (mg *Module) NewController(col Collectionist, path string) *Controller {
	cnt := NewController(col, mg.MountPath+path)
	mg.HttpControllers = append(mg.HttpControllers, cnt)
	tp := col.GetInstanceType()
	AddControllerMap(tp, cnt)
	return cnt
}

func (mg *Module) RegisterController(cnt Controlist) {
	cnt.SetBase(mg.MountPath + cnt.GetBase())
	mg.HttpControllers = append(mg.HttpControllers, cnt)
	col := cnt.GetCollection()
	tp := col.GetInstanceType()
	AddControllerMap(tp, cnt)
}

func (mg *Module) Mount(nmg Moduler, mountbase string) {
	mg.Modules = append(mg.Modules, nmg)
	nmg.SetMountPath(mg.MountPath + mountbase)
	nmg.SetApp(mg.App)
	nmg.SetDataSrc(mg.GetDataSrc())
	err := nmg.Init(mg.App)
	if err != nil {
		panic("mount error at: " + mountbase + " error message is: " + err.Error())
	}

	mg.App.Router.GET(mg.MountPath+mountbase+"/meta", nmg.Meta)
}

func (mg *Module) SetCrudRoutes(cont Controlist, excludes []string) {

	var base string = cont.GetBase()
	mg.App.Router.POST(base, cont.Create)
	// because gin does not have regexp!
	if !Contains(excludes, "Get") {
		mg.App.Router.GET(base+"/items/:id", cont.Get)
	}
	if !Contains(excludes, "Report") {
		mg.App.Router.GET(base+"/report", cont.Report)
	}
	if !Contains(excludes, "Meta") {
		mg.App.Router.GET(base+"/meta", cont.Meta)
	}
	// mg.App.Router.GET(base+"/report", cont.Report)
	if !Contains(excludes, "List") {
		mg.App.Router.GET(base, cont.List)
	}
	if !Contains(excludes, "Delete") {
		mg.App.Router.DELETE(base+"/items/:id", cont.Delete)
	}
	if !Contains(excludes, "Update") {
		mg.App.Router.PUT(base+"/items/:id", cont.Update)
	}
	if !Contains(excludes, "Rel") {
		mg.App.Router.POST(base+"/items/:id/relation/:field", cont.Rel)
	}
	if !Contains(excludes, "UpdateRel") {
		mg.App.Router.PUT(base+"/items/:id/relation/:field", cont.UpdateRel)
	}
	if !Contains(excludes, "UnRel") {
		mg.App.Router.DELETE(base+"/items/:id/relation/:field", cont.UnRel)
	}
	if !Contains(excludes, "GetRel") {
		mg.App.Router.GET(base+"/items/:id/relation/:field", cont.GetRel)

	}
}

func (mg *Module) SetRelRoutes(cont Controlist) {
	var base string = cont.GetBase()
	mg.App.Router.POST(base+"/items/:id/relation/:field", cont.Rel)
	mg.App.Router.PUT(base+"/items/:id/relation/:field", cont.UpdateRel)
	mg.App.Router.DELETE(base+"/items/:id/relation/:field", cont.UnRel)
	mg.App.Router.GET(base+"/items/:id/relation/:field", cont.GetRel)
}

type ModuleInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

func (mg *Module) Meta(c *gin.Context) {
	mgs := make([]ModuleInfo, len(mg.Modules)+len(mg.HttpControllers))
	var i int = 0
	for _, mg := range mg.Modules {
		mi := ModuleInfo{mg.GetName(), mg.GetPath(), "Module"}
		mgs[i] = mi
		i += 1
	}
	for _, cont := range mg.HttpControllers {
		ci := cont.GetInfo()
		if ci.Show {
			mi := ModuleInfo{ci.Name, ci.Path, "Collection"}
			mgs[i] = mi
			i += 1
		}
	}
	c.JSON(http.StatusOK, mgs)
}

// make module to accept routes instead of app.router
func (mg *Module) GET(path string, handler gin.HandlerFunc) gin.IRoutes {
	return mg.App.Router.GET(mg.MountPath+path, handler)
}
func (mg *Module) POST(path string, handler gin.HandlerFunc) gin.IRoutes {
	return mg.App.Router.POST(mg.MountPath+path, handler)
}
func (mg *Module) DELETE(path string, handler gin.HandlerFunc) gin.IRoutes {
	return mg.App.Router.DELETE(mg.MountPath+path, handler)
}
func (mg *Module) PUT(path string, handler gin.HandlerFunc) gin.IRoutes {
	return mg.App.Router.PUT(mg.MountPath+path, handler)
}
func (mg *Module) HEAD(path string, handler gin.HandlerFunc) gin.IRoutes {
	return mg.App.Router.HEAD(mg.MountPath+path, handler)
}
func (mg *Module) PATCH(path string, handler gin.HandlerFunc) gin.IRoutes {
	return mg.App.Router.PATCH(mg.MountPath+path, handler)
}
func (mg *Module) OPTIONS(path string, handler gin.HandlerFunc) gin.IRoutes {
	return mg.App.Router.OPTIONS(mg.MountPath+path, handler)
}
