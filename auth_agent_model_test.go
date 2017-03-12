package hermes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Agent_Start(t *testing.T) {

	addTempTables()
}
func TestCreateAgent(t *testing.T) {

	agent, err := AgentColl.Create(SystemToken, nil, &Agent{Identity: "m.ghoreishi1@gmail.com", Password: ""})
	assert.Error(t, err, "NotValid")

	//lenght of passowrd must be greater than 6
	agent, err = AgentColl.Create(SystemToken, nil, &Agent{Identity: "m.ghoreishi1@gmail.com", Password: "12345"})
	assert.Error(t, err, "NotValid")

	//lenght of passowrd must be greater than 6
	agent, err = AgentColl.Create(SystemToken, nil, &Agent{Identity: "m.ghoreishi1@gmail.com", Password: "123456"})
	assert.NoError(t, err)
	ragent := agent.(*Agent)
	assert.Equal(t, 1, ragent.Id)

	//duplicate error
	agent, err = AgentColl.Create(SystemToken, nil, &Agent{Identity: "m.ghoreishi1@gmail.com", Password: "123456"})
	assert.Error(t, err, "DuplicateData")
}

func Test_UpdatePasswordByToken(t *testing.T) {

	AgentTokenColl := AgentTokenCollection{}

	token, err := AgentTokenColl.CreateToken(NewToken(1, "password"))
	assert.NoError(t, err)
	token_test = token.Token

	//old password 123456 -- change to 654321
	err = AgentColl.UpdatePasswordByToken(token_test, "654321", "m.ghoreishi1@gmail.com")
	assert.NoError(t, err)

	//old password error
	_, err = AgentColl.GetByIdentityPass("m.ghoreishi1@gmail.com", "123456")
	assert.Error(t, err, "DbNotFoundError")

	//new password
	agent, err := AgentColl.GetByIdentityPass("m.ghoreishi1@gmail.com", "654321")
	assert.NoError(t, err)
	assert.Equal(t, 1, agent.Id)
}

func Test_UpdatePasswordByOld(t *testing.T) {
	//current password 654321 -- change to mahsa1
	updated, err := AgentColl.UpdatePasswordByOld(SystemToken, 1, "654321", "mahsa1")
	assert.NoError(t, err)
	assert.Equal(t, true, updated)

	//old password error
	_, err = AgentColl.GetByIdentityPass("m.ghoreishi1@gmail.com", "654321")
	assert.Error(t, err, "DbNotFoundError")

	//new password
	agent, err := AgentColl.GetByIdentityPass("m.ghoreishi1@gmail.com", "mahsa1")
	assert.NoError(t, err)
	assert.Equal(t, 1, agent.Id)
}

func Test_GetByIdentityPass(t *testing.T) {
	//current password : mahsa1
	//test wrong email address
	_, err := AgentColl.GetByIdentityPass("notexist@gmail.com", "123456")
	assert.Error(t, err, "DbNotFoundError")

	//test wrong password
	_, err = AgentColl.GetByIdentityPass("m.ghoreishi1@gmail.com", "123456")
	assert.Error(t, err, "DbNotFoundError")

	//test currect password
	result, err := AgentColl.GetByIdentityPass("m.ghoreishi1@gmail.com", "mahsa1")
	assert.NoError(t, err)
	assert.Equal(t, 1, result.Id)
}

func Test_ExistsByIdentity(t *testing.T) {
	//test wrong email address
	exists, err := AgentColl.ExistsByIdentity("notexist@gmail.com")
	assert.NoError(t, err)
	assert.Equal(t, false, exists)

	//test currect email
	exists, err = AgentColl.ExistsByIdentity("m.ghoreishi1@gmail.com")
	assert.NoError(t, err)
	assert.Equal(t, true, exists)
}

func Test_Login(t *testing.T) {
	//current password mahsa1
	//not existing email
	agent := Agent{}
	agent.Identity = "notexists@gmail.com"
	agent.Password = "123456"

	agent.Device.Platform = "ios"
	agent.Device.Ip = "19.168.20.2"
	agent.Device.Ip = "122356789"

	tok, err := AgentColl.Login(agent, "")
	assert.Error(t, err, "NotFound")
	assert.Equal(t, "", tok.Token)

	//existing email
	agent = Agent{}
	agent.Identity = "m.ghoreishi1@gmail.com"
	agent.Password = "wrongpass"

	agent.Device.Platform = "ios"
	agent.Device.Ip = "19.168.20.2"
	agent.Device.Ip = "122356789"

	tok, err = AgentColl.Login(agent, "")
	assert.Error(t, err, "NotFound")
	assert.Equal(t, "", tok.Token)

	//existing email
	agent = Agent{}
	agent.Identity = "m.ghoreishi1@gmail.com"
	agent.Password = "mahsa1"

	agent.Device.Platform = "ios"
	agent.Device.Ip = "19.168.20.2"
	agent.Device.Ip = "122356789"

	tok, err = AgentColl.Login(agent, "")
	assert.NoError(t, err)
	assert.NotEqual(t, "", tok.Token)

}
func Test_Agent_End(t *testing.T) {
	rmTempTables()
}
