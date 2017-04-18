package hermes

import (
	"errors"
)

var AuthorizationModule *authModule = &authModule{}

type authModule struct {
	// module not defined
	Module
}

var permissions = []string{"Create Agent", "Get Agent", "Edit Agent|$own", "Delete Agent", "Update Agent", "Add Permission", "Edit Role", "Add Role", "Delete Role"}

func createCollections(app *App) error {
	var collErr error

	permColl, collErr = NewDBCollection(&Permission{}, app.DataSrc)
	roleColl, collErr = NewDBCollection(&Role{}, app.DataSrc)
	rolePermColl, collErr = NewDBCollection(&Role_Permission{}, app.DataSrc)
	RoleAgentColl, collErr = NewDBCollection(&Role_Agent{}, app.DataSrc)
	AgentColl, collErr = NewAgentCollection(app.DataSrc)

	AgentTokenColl, collErr = NewAgentTokenCollection(&AgentToken{}, app.DataSrc)
	DeviceColl, collErr = NewDBCollection(&Device{}, app.DataSrc)

	return collErr

}

var secretKey string
var application *App

var permColl, roleColl, rolePermColl, RoleAgentColl Collectionist
var AgentTokenColl *AgentTokenCollection

var permCont, roleCont, rolePermCont, roleAgentCont, agentTokenCont, deviceCont *Controller

func (um *authModule) Init(app *App) error {
	um.Name = "Auth"
	application = app
	settings := app.GetSettings("Agents")
	if settings["Secret"] == nil {
		return errors.New("secret config does not exists")
	}
	secretKey = settings["Secret"].(string)

	permCache = make(map[string][]string, PermissionCacheLength)

	collErr := createCollections(app)
	if collErr != nil {
		errmsg := "Error in creating collection:" + collErr.Error()
		return errors.New(errmsg)
	}

	RegisterPermissions(permissions)

	AgentColl.Conf().SetAuth("Create Agent", "Get Agent", "List Agent", "Update Agent", "Delete Agent", "Relate Agent")
	AgentTokenColl.Conf().SetAuth("Create Agent Token", "Get Agent Token", "List Agent Token", "Update Agent Token", "Delete Agent Token", "Relate Agent Token")
	DeviceColl.Conf().SetAuth("", "", "List Device", "", "", "")

	setRoutes(app)
	return nil

}

func setRoutes(app *App) {
	permCont = AuthorizationModule.NewController(permColl, "/permissions")
	roleCont = AuthorizationModule.NewController(roleColl, "/roles")
	rolePermCont = AuthorizationModule.NewController(rolePermColl, "/roles_permissions")
	roleAgentCont = AuthorizationModule.NewController(RoleAgentColl, "/roles_agents")
	agentCont = NewAgentController(AgentColl, "/agents")
	// agentTokenCont = NewController(agentTokenColl, mountBase+"/permissions")
	agentTokenCont = AuthorizationModule.NewController(AgentTokenColl, "/agent_tokens")

	// app.Router.POST(permCont.GetBase(), permCont.Create)
	app.Router.GET(permCont.GetBase(), permCont.List)
	app.Router.GET(permCont.GetBase()+"/meta", permCont.Meta)

	AuthorizationModule.SetCrudRoutes(agentCont)
	AuthorizationModule.POST("/login", agentCont.Login)
	AuthorizationModule.POST("/flogin/:ftoken", agentCont.FBLogin)
	AuthorizationModule.GET("/logout/:token", agentCont.Logout)
	AuthorizationModule.PUT("/changePassword/:id", agentCont.ChangePassword)
	AuthorizationModule.GET("/forgetPassword/:identity", agentCont.RequestPasswordToken)
	AuthorizationModule.POST("/changePasswordByToken/:token", agentCont.ChangePasswordByToken)
	AuthorizationModule.GET("/activeUser/:token", agentCont.ActiveUserByToken)

	AuthorizationModule.SetCrudRoutes(rolePermCont)
	AuthorizationModule.SetCrudRoutes(roleAgentCont)
	AuthorizationModule.SetCrudRoutes(roleCont)
	AuthorizationModule.SetCrudRoutes(agentTokenCont)

	deviceCont = AuthorizationModule.NewController(DeviceColl, "/devices")
	app.Router.GET(deviceCont.GetBase()+"/meta", deviceCont.Meta)
	app.Router.GET(deviceCont.GetBase()+"/report", deviceCont.Report)

	app.Router.GET(deviceCont.GetBase(), deviceCont.List)

}
