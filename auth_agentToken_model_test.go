package hermes

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var token_test string

func Test_AgentToken_Start(t *testing.T) {
	addTempTables()
}
func TestCreateToken(t *testing.T) {
	AgentColl.ActiveByDefalut = true
	_, err := AgentColl.Create(SystemToken, nil, &Agent{Identity: "m.ghoreishi1@gmail.com", Password: "123456"})
	assert.NoError(t, err)

	token := AgentToken{}
	newToken := NewToken(1, "login")
	res, err := token.Create(newToken)
	assert.NoError(t, err)
	assert.Equal(t, newToken.Token, res.Token)
	assert.Equal(t, newToken.Type, res.Type)

	token_test = res.Token
}

func TestGetToken(t *testing.T) {
	token := AgentToken{}
	token, err := token.GetToken(token_test, "login")
	assert.NoError(t, err)
	assert.Equal(t, 2, token.Id)
	assert.Equal(t, token_test, token.Token)
	assert.Equal(t, false, token.Is_Expired)
}

func TestExists(t *testing.T) {
	token := AgentToken{}
	exists, err := token.Exists(token_test)
	assert.NoError(t, err)
	assert.Equal(t, true, exists)
}
func TestLogout(t *testing.T) {
	token := AgentToken{}
	err := token.Logout(token_test)
	assert.NoError(t, err)

	token, err = token.GetToken(token_test, "login")
	assert.Error(t, err, "NotFound")

}
func Test_AgentToken_End(t *testing.T) {
	rmTempTables()
}
