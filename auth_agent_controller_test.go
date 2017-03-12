package hermes

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var pass_token string

func Test_AgentController_Start(t *testing.T) {
	addTempTables()
}

func Test_GetAgentByToken(t *testing.T) {
	_, err := AgentColl.Create(SystemToken, nil, &Agent{Identity: "m.ghoreishi1@gmail.com", Password: "123456"})
	assert.NoError(t, err)
	AgentTokenColl, err := NewAgentTokenCollection(AgentToken{}, DBTest())
	assert.NoError(t, err)

	newToken := NewToken(1, "login")
	res, err := AgentTokenColl.CreateToken(newToken)
	assert.NoError(t, err)
	assert.Equal(t, res.Token, newToken.Token)
	token_test = res.Token
	// cont := NewAgentController(AgentColl, "")

	// gin.SetMode(gin.TestMode)
	// router := gin.New()
	// router.GET("/test", func(c *gin.Context) {
	// 	cont.GetAgentByToken(c)
	// })

	// request, _ := http.NewRequest("GET", "/test", nil)
	// request.Header.Set("Authorization", token_test)
	// response := httptest.NewRecorder()
	// router.ServeHTTP(response, request)

	// agent := Agent{}
	// json.Unmarshal([]byte(response.Body.String()), &agent)
	// assert.Equal(t, http.StatusOK, response.Code)
	// assert.Equal(t, "m.ghoreishi1@gmail.com", agent.Identity)
}

func Test_ChangePassword(t *testing.T) {
	cont := NewAgentController(AgentColl, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.PUT("/changePassowrd/:id", func(c *gin.Context) {
		cont.ChangePassword(c)
	})

	//wrong password error test

	p := Password{}
	p.Old_Password = "wrongpassword"
	p.New_Password = "mahsagh123"

	sJson, _ := json.Marshal(p)
	contentReader := bytes.NewReader(sJson)

	request, _ := http.NewRequest("PUT", "/changePassowrd/1", contentReader)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "\"Password is not correct!\"\n", response.Body.String())

	//corect password test
	p = Password{}
	p.Old_Password = "123456"
	p.New_Password = "mahsagh123"

	sJson, _ = json.Marshal(p)
	contentReader = bytes.NewReader(sJson)

	request, _ = http.NewRequest("PUT", "/changePassowrd/1", contentReader)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	result, err := AgentColl.Get(SystemToken, 1, "")
	agent := result.(*Agent)
	assert.NoError(t, err)
	assert.Equal(t, GenerateHash("mahsagh123", secretKey), agent.Password)
	assert.NotEqual(t, GenerateHash("test", secretKey), agent.Password)

}

func Test_RequestPasswordToken(t *testing.T) {
	cont := NewAgentController(AgentColl, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/forgetPassword/:identity", func(c *gin.Context) {
		cont.RequestPasswordToken(c)
	})

	//not exist user
	request, _ := http.NewRequest("GET", "/forgetPassword/notexist@gmail.com", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "\"Resource not found!\"\n", response.Body.String())

	//exist user
	request, _ = http.NewRequest("GET", "/forgetPassword/m.ghoreishi1@gmail.com", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	assert.Equal(t, http.StatusOK, response.Code)

}

func Test_ChangePasswordByToken(t *testing.T) {
	cont := NewAgentController(AgentColl, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("changePasswordByToken/:token", func(c *gin.Context) {
		cont.ChangePasswordByToken(c)
	})

	//fake token error

	p := Password{}
	p.New_Password = "123456"

	sJson, _ := json.Marshal(p)
	contentReader := bytes.NewReader(sJson)

	request, _ := http.NewRequest("POST", "/changePasswordByToken/"+uuid.NewV4().String(), contentReader)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "\"Resource not found!\"\n", response.Body.String())

	//existing token test

	var _token []string
	DBTest().DB.Select(&_token, " select token from agent_tokens where type='password' ")

	p = Password{}
	p.New_Password = "123456"

	sJson, _ = json.Marshal(p)
	contentReader = bytes.NewReader(sJson)

	request, _ = http.NewRequest("POST", "/changePasswordByToken/"+_token[0], contentReader)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "\"Resource updated successfully!\"\n", response.Body.String())

}

func Test_Login_Controller(t *testing.T) {
	cont := NewAgentController(AgentColl, "")
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("login", func(c *gin.Context) {
		cont.Login(c)
	})

	//wrong email
	agent := &Agent{}
	agent.Identity = "notexists@gmail.com"
	agent.Password = "123456"

	agent.Device.Platform = "ios"
	agent.Device.Ip = "19.168.20.2"
	agent.Device.Ip = "122356789"
	sJson, _ := json.Marshal(agent)
	contentReader := bytes.NewReader(sJson)

	request, _ := http.NewRequest("POST", "/login", contentReader)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "\"Resource not found!\"\n", response.Body.String())

	//wrong password
	agent = &Agent{}
	agent.Identity = "m.ghoreshi1@gmail.com"
	agent.Password = "wrongpass"

	agent.Device.Platform = "ios"
	agent.Device.Ip = "19.168.20.2"
	agent.Device.Ip = "122356789"
	sJson, _ = json.Marshal(agent)
	contentReader = bytes.NewReader(sJson)

	request, _ = http.NewRequest("POST", "/login", contentReader)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "\"Resource not found!\"\n", response.Body.String())

	//correct email password
	DBTest().DB.Exec(" update agents set is_active=true where identity='m.ghoreishi1@gmail.com' ")

	agent = &Agent{}
	agent.Identity = "m.ghoreishi1@gmail.com"
	agent.Password = "123456"

	agent.Device.Platform = "ios"
	agent.Device.Ip = "19.168.20.2"
	agent.Device.Ip = "122356789"
	sJson, _ = json.Marshal(agent)
	contentReader = bytes.NewReader(sJson)

	request, _ = http.NewRequest("POST", "/login", contentReader)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	agentTok := AgentToken{}
	json.Unmarshal([]byte(response.Body.String()), &agentTok)
	assert.Equal(t, agentTok.Agent_Id, 1)
	assert.Equal(t, agentTok.Is_Expired, false)
	assert.Equal(t, agentTok.Type, "login")
	assert.NotEqual(t, agentTok.Token, "")
	token_test = agentTok.Token
}

func Test_Logout(t *testing.T) {
	cont := NewAgentController(AgentColl, "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("logout/:token", func(c *gin.Context) {
		cont.Logout(c)
	})

	request, _ := http.NewRequest("GET", "/logout/"+token_test, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "\"Resource deleted successfully!\"\n", response.Body.String())

	// agntTok := AgentToken{}
	AgentTokenColl, err := NewAgentTokenCollection(AgentToken{}, DBTest())
	assert.NoError(t, err)

	_, err = AgentTokenColl.GetToken(token_test, "login")
	assert.Error(t, err, "NotFound")

}
func Test_AgentController_End(t *testing.T) {
	rmTempTables()
}
