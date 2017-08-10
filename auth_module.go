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

//define hermes collections
func createCollections(app *App) error {
	var collErr error

	permColl, collErr = NewDBCollection(&Permission{}, app.DataSrc)
	roleColl, collErr = NewDBCollection(&Role{}, app.DataSrc)
	rolePermColl, collErr = NewDBCollection(&Role_Permission{}, app.DataSrc)
	RoleAgentColl.Conf().SetAuth("Create Role Permission", "Get Role Permission", "List Role Permission", "Update Role Permission", "Delete Role Permission", "Update Role Permission")

	RoleAgentColl, collErr = NewDBCollection(&Role_Agent{}, app.DataSrc)
	RoleAgentColl.Conf().SetAuth("Create Role Agent", "Get Role Agent", "List Role Agent", "Update Role Agent", "Delete Role Agent", "Update Role Agent")

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
	settings := app.GetSettings("agents")
	//secret key is necessary for encrypting passwords
	if settings["secret"] == nil {
		return errors.New("secret config does not exists")
	}
	secretKey = settings["secret"].(string)

	permCache = make(map[string][]string, PermissionCacheLength)

	collErr := createCollections(app)
	if collErr != nil {
		errmsg := "Error in creating collection:" + collErr.Error()
		return errors.New(errmsg)
	}

	//insert permissions in db if not exists
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
	agentTokenCont = AuthorizationModule.NewController(AgentTokenColl, "/agent_tokens")

	app.Router.GET(permCont.GetBase(), permCont.List)
	app.Router.GET(permCont.GetBase()+"/meta", permCont.Meta)

	//auth apis
	AuthorizationModule.SetCrudRoutes(agentCont, []string{})
	AuthorizationModule.POST("/login", agentCont.Login)
	AuthorizationModule.POST("/flogin/:ftoken", agentCont.FBLogin)
	AuthorizationModule.GET("/logout/:token", agentCont.Logout)
	AuthorizationModule.PUT("/changePassword/:id", agentCont.ChangePassword)
	AuthorizationModule.GET("/forgetPassword/:identity", agentCont.RequestPasswordToken)
	AuthorizationModule.POST("/changePasswordByToken/:token", agentCont.ChangePasswordByToken)
	AuthorizationModule.GET("/activeUser/:token", agentCont.ActiveUserByToken)

	AuthorizationModule.SetCrudRoutes(rolePermCont, []string{})
	AuthorizationModule.SetCrudRoutes(roleAgentCont, []string{})
	AuthorizationModule.SetCrudRoutes(roleCont, []string{})
	AuthorizationModule.SetCrudRoutes(agentTokenCont, []string{})

	deviceCont = AuthorizationModule.NewController(DeviceColl, "/devices")
	app.Router.GET(deviceCont.GetBase()+"/meta", deviceCont.Meta)
	app.Router.GET(deviceCont.GetBase()+"/report", deviceCont.Report)

	app.Router.GET(deviceCont.GetBase(), deviceCont.List)

}
