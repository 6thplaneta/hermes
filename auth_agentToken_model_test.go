package hermes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var token_test string

func Test_AgentToken_Start(t *testing.T) {
	addTempTables()
}
func TestCreateToken(t *testing.T) {
	AgentColl.ActiveByDefalut = true
	_, err := AgentColl.Create(SystemToken, nil, &Agent{Identity: "m.ghoreishi1@gmail.com", Password: "123456"})
	// assert.NoError(t, err)

	AgentTokenColl, err := NewAgentTokenCollection(AgentToken{}, DBTest())
	assert.NoError(t, err)

	newToken := NewToken(1, "login")
	res, err := AgentTokenColl.CreateToken(newToken)
	assert.NoError(t, err)
	assert.Equal(t, newToken.Token, res.Token)
	assert.Equal(t, newToken.Type, res.Type)

	token_test = res.Token
}

func TestGetToken(t *testing.T) {
	AgentTokenColl, err := NewAgentTokenCollection(AgentToken{}, DBTest())
	assert.NoError(t, err)

	token, err := AgentTokenColl.GetToken(token_test, "login")
	assert.NoError(t, err)
	// assert.Equal(t, 2, token.Id)
	assert.Equal(t, token_test, token.Token)
	assert.Equal(t, false, token.Is_Expired)
}

func TestExists(t *testing.T) {
	// token := AgentToken{}
	AgentTokenColl, err := NewAgentTokenCollection(AgentToken{}, DBTest())
	assert.NoError(t, err)

	exists, err := AgentTokenColl.Exists(token_test)
	assert.NoError(t, err)
	assert.Equal(t, true, exists)
}
func TestLogout(t *testing.T) {
	AgentTokenColl, err := NewAgentTokenCollection(AgentToken{}, DBTest())
	assert.NoError(t, err)

	err = AgentTokenColl.Logout(token_test)
	assert.NoError(t, err)

	_, err = AgentTokenColl.GetToken(token_test, "login")
	assert.Error(t, err, "NotFound")

}
func Test_AgentToken_End(t *testing.T) {
	rmTempTables()
	DBTest().DB.Exec("deallocate all;")

}
