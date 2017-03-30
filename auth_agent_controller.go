package hermes

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type AgentController struct {
	*Controller
}

var agentCont *AgentController

func NewAgentController(coll Collectionist, base string) *AgentController {

	cnt := AuthorizationModule.NewController(coll, base)

	cont := &AgentController{cnt}
	return cont
}

type Password struct {
	Old_Password string `json:"old_password"`
	New_Password string `json:"new_password"`
}

func (agentCont *AgentController) ChangePassword(c *gin.Context) {

	id, _ := strconv.Atoi(c.Param("id"))

	p := Password{}
	c.BindJSON(&p)
	token := c.Request.Header.Get("Authorization")
	_, err := AgentColl.UpdatePasswordByOld(token, id, p.Old_Password, p.New_Password)
	if application.Logger.Level >= 6 {
		secretP := p
		secretP.New_Password = "******"
		secretP.Old_Password = "******"
		s, _ := json.Marshal(secretP)
		application.Logger.LogHttpByBody(c, string(s))
	}

	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceUpdated"])

}

func (agentCont *AgentController) RequestPasswordToken(c *gin.Context) {

	_, err := AgentColl.RequestPasswordToken(c.Param("identity"))
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceCreated"])

}

func (agentCont *AgentController) ChangePasswordByToken(c *gin.Context) {
	p := Password{}
	c.BindJSON(&p)

	err := AgentColl.UpdatePasswordByToken(c.Param("token"), p.New_Password, c.Query("email"))

	if application.Logger.Level >= 6 {
		secretP := p
		secretP.New_Password = "******"
		secretP.Old_Password = "******"
		s, _ := json.Marshal(secretP)
		application.Logger.LogHttpByBody(c, string(s))
	}

	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceUpdated"])

}

func (agentCont *AgentController) ActiveUserByToken(c *gin.Context) {

	err := AgentColl.ActiveUserByToken(c.Param("token"))
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceUpdated"])

}

func (agentCont *AgentController) FBLogin(c *gin.Context) {
	ftoken := c.Param("ftoken")
	agent := Agent{}
	agent.Device.Ip = c.ClientIP()
	agentToken, err := AgentColl.FBLogin(agent, ftoken)
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, agentToken)
}

func (agentCont *AgentController) Login(c *gin.Context) {
	agent := Agent{}
	c.BindJSON(&agent)
	agent.Device.Ip = c.ClientIP()
	agentToken, err := AgentColl.Login(agent, "")
	if err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, agentToken)
}

func (agentCont *AgentController) Logout(c *gin.Context) {

	if err := AgentTokenColl.Logout(c.Param("token")); err != nil {
		HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, Messages["ResourceDeleted"])
}
