package hermes

import (
	"fmt"
	"github.com/satori/go.uuid"
	"math/rand"
	"strconv"
	"time"
)

type AgentToken struct {
	Id            int       `json:"id" hermes:"dbspace:agent_tokens,ui-html:None"`
	Agent_Id      int       `json:"agent_id" validate:"required"`
	Token         string    `json:"token" validate:"required"`
	Creation_Date time.Time `json:"creation_date" validate:"required" hermes:"type:time"`
	Type          string    `json:"type" validate:"required"`
	Is_Expired    bool      `json:"is_expired" hermes:"editable"`
	Device_Id     int       `json:"device_id" validate:"required"`
	Device        Device    `db:"-" json:"device" validate:"structonly" hermes:"many2one"`
}

type AgentTokenCollection struct {
	*Collection
}

func NewAgentTokenCollection(instance interface{}, dsrc *DataSrc) (*AgentTokenCollection, error) {
	tmpc, err := NewDBCollection(instance, dsrc)
	return &AgentTokenCollection{tmpc}, err
}

func (pt *AgentTokenCollection) GetToken(token string, typ string) (AgentToken, error) {
	result := AgentToken{}
	strQ := fmt.Sprintf("select * from agent_tokens where is_expired=false and token= '%s' ", token)
	if typ != "" {
		strQ += " and type='" + typ + "'"
	}
	err := pt.DataSrc.DB.Get(&result, strQ)
	if err == ErrNoRows {
		return AgentToken{}, ErrTokenInvalid
	}
	if err != nil {
		return AgentToken{}, err
	}

	return result, err

}

func (pt *AgentTokenCollection) Exists(token string) (bool, error) {
	result := AgentToken{}
	err := pt.DataSrc.DB.Get(&result, fmt.Sprintf("select * from agent_tokens where is_expired=false and token= '%s' ", token))

	if err != nil && err.Error() != Messages["DbNotFoundError"] {
		return false, err
	}

	if err != nil && err.Error() == Messages["DbNotFoundError"] {
		return false, nil
	}

	return true, err

}

func (pt *AgentTokenCollection) Logout(token string) error {
	_, err := pt.DataSrc.DB.Exec(fmt.Sprintf("update agent_tokens set is_expired = true where token= '%s' ", token))
	if err != nil && err.Error() != Messages["DbNotFoundError"] {
		return err
	}
	return nil
}

func (pt *AgentTokenCollection) CreateToken(agentToken AgentToken) (AgentToken, error) {
	var id int
	err := pt.DataSrc.DB.QueryRow(getInsertQuery(&agentToken)).Scan(&id)
	if err != nil {
		return AgentToken{}, err
	}

	agentToken.Id = id
	return agentToken, nil
}

func NewToken(agentid int, tokenType string) AgentToken {
	agentToken := AgentToken{}

	agentToken.Agent_Id = agentid

	agentToken.Creation_Date = time.Now()
	agentToken.Type = tokenType

	if tokenType == "password" || tokenType == "activation" {
		agentToken.Token = random(10000, 99999)
	} else {
		u1 := uuid.NewV4()
		agentToken.Token = u1.String()
	}
	return agentToken
}

func random(min, max int) string {
	rand.Seed(time.Now().Unix())
	intR := rand.Intn(max-min) + min
	return strconv.Itoa(intR)
}
