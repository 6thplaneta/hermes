package hermes

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Agent struct {
	Id          int    `json:"id" hermes:"dbspace:agents"`
	Identity    string `json:"identity" validate:"required"`
	Password    string `json:"password,omitempty"`
	Is_Active   bool   `json:"is_active" hermes:"editable"`
	Is_Super    bool   `json:"is_super" hermes:"editable"`
	FId         string `json:"fid"`
	GId         string `json:"gid"`
	No_Password bool   `json:"-"`
	Roles       []Role `json:"roles,omitempty" db:"-" hermes:"many2many:Role_Agent"`
	Device      Device `db:"-" json:"device"`
}

type AgentCollection struct {
	*Collection
	ActiveByDefalut bool
}

var AgentColl *AgentCollection

func NewAgentCollection(datasrc *DataSrc) (*AgentCollection, error) {
	coll, err := NewDBCollection(&Agent{}, datasrc)

	typ := reflect.TypeOf(Agent{})
	AgentColl := &AgentCollection{coll, false}
	CollectionsMap[typ] = AgentColl

	return AgentColl, err
}

//override
func (col *AgentCollection) Create(token string, trans *sql.Tx, inpuser interface{}) (interface{}, error) {

	var newAgent Agent
	agent := inpuser.(*Agent)
	if agent.FId != "" || agent.GId != "" {
		//
		agent.No_Password = true
		agent.Is_Active = true
	}

	newAgent = *agent
	if agent.No_Password == false {

		if agent.Password == "" {
			return &Agent{}, ErrPassRequired
		}

		match, _ := regexp.MatchString(Messages["PasswordFormat"], agent.Password)

		if match == false {
			return &Agent{}, ErrPassFormat
		}
		newAgent.Password = GenerateHash(agent.Password, secretKey)
	}
	isExist, err := AgentColl.ExistsByIdentity(agent.Identity)

	if err != nil {
		//general error
		return &Agent{}, err
	}
	if isExist {
		return &Agent{}, ErrDuplicate
	}
	// if agent.No_Password == false {
	// 	newAgent.Is_Active = col.ActiveByDefalut
	// } else {
	// 	newAgent.Is_Active = true
	// }
	result, err := AgentColl.Collection.Create(token, nil, &newAgent)

	if err != nil {
		return &Agent{}, err
	}
	obj := result.(*Agent)

	if agent.No_Password == false {
		_, err := AgentTokenColl.CreateToken(NewToken(obj.Id, "activation"))
		if err != nil {
			return &Agent{}, err
		}
	}
	return obj, err
}

//override
func (col *AgentCollection) Update(token string, id int, obj interface{}) error {
	var err error
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Update, id, "UPDATE", cnf.CheckAccess) {

		agent, err := AgentColl.GetByLoginToken(token)
		if err != nil {
			return err
		}

		if id == agent.Id {
			token = SystemToken
		}
	}
	err = AgentColl.Collection.Update(token, id, obj)
	if err != nil {
		return err
	}

	return nil
}

func (col *AgentCollection) GetByToken(token string) (interface{}, error) {
	result := Agent{}
	agentToken, err := AgentTokenColl.GetToken(token, "login")
	if err != nil {
		return nil, err
	}
	err = col.DataSrc.DB.Get(&result, fmt.Sprintf("select * from agents where id= %d", agentToken.Agent_Id))
	return result, err
}

// duplicate should be deleted
func (col *AgentCollection) GetByLoginToken(token string) (Agent, error) {
	agentToken, err := AgentTokenColl.GetToken(token, "login")
	if err != nil {
		return Agent{}, err
	}

	result := Agent{}
	err = col.DataSrc.DB.Get(&result, fmt.Sprintf("select * from agents where id= %d", agentToken.Agent_Id))

	if err != nil && err == ErrNoRows {
		return Agent{}, ErrNotFound
	}
	return result, err
}

func (col *AgentCollection) UpdatePasswordByToken(token string, newPassword string, identity string) error {
	passToken, err := AgentTokenColl.GetToken(token, "password")
	if err != nil {
		return err
	}

	if newPassword == "" {
		return ErrPassRequired
	}

	match, _ := regexp.MatchString(Messages["PasswordFormat"], newPassword)

	if match == false {
		return ErrPassFormat
	}

	result, err := AgentColl.Get(SystemToken, passToken.Agent_Id, "")
	var agent = result.(*Agent)

	if agent.Identity != identity {
		return errors.New("No Match")

	}

	if err != nil && err == ErrNoRows {
		return ErrPassRequired
	} else if err != nil && err != ErrNoRows {
		return err
	}

	r, err := col.DataSrc.DB.Exec(fmt.Sprintf("update %s set password = '%s' where id= %d", col.Dbspace, GenerateHash(newPassword, secretKey), agent.Id))
	if err != nil {
		return err
	}
	rowCount, err := r.RowsAffected()
	if rowCount == 0 {
		return ErrNotFound
	}
	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='password' and agent_id=%d ", agent.Id))
	if err != nil {
		return err
	}
	return nil
}

func (col *AgentCollection) UpdatePasswordByOld(tokenn string, id int, oldPassword string, newPassword string) (bool, error) {
	if newPassword == "" || oldPassword == "" {
		return false, ErrPassRequired
	}

	match, _ := regexp.MatchString(Messages["PasswordFormat"], newPassword)

	if match == false {
		return false, ErrPassFormat
	}

	result, err := AgentColl.Get(SystemToken, id, "")
	var agent = result.(*Agent)

	if err != nil && err == ErrNoRows {
		return false, ErrPassRequired
	} else if err != nil && err != ErrNoRows {
		return false, err
	}

	if agent.Password != GenerateHash(oldPassword, secretKey) {
		return false, ErrPassword
	}

	r, err := col.DataSrc.DB.Exec(fmt.Sprintf("update %s set password = '%s' where id= %d", col.Dbspace, GenerateHash(newPassword, secretKey), id))
	if err != nil {
		return false, err
	}
	rowCount, err := r.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowCount == 0 {
		return false, ErrNotFound
	}
	return true, nil
}

func (col *AgentCollection) GetByFId(identity string, fid string) (Agent, error) {

	agent := Agent{}
	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where identity= '%s' and fid= '%s' ", identity, fid))
	if err != nil {
		if err == ErrNoRows {
			return Agent{}, ErrNotFound
		}
		return Agent{}, err
	}

	return agent, nil
}

func (col *AgentCollection) GetByGId(identity string, gid string) (Agent, error) {

	agent := Agent{}
	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where identity= '%s' and gid= '%s' ", identity, gid))
	if err != nil {
		if err == ErrNoRows {
			return Agent{}, ErrNotFound
		}
		return Agent{}, err
	}

	return agent, nil
}

// func (col *AgentCollection) GetByVKId(identity string, vkid string) (Agent, error) {

// 	agent := Agent{}
// 	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where identity= '%s' and vkid= '%s' ", identity, vkid))
// 	if err != nil {
// 		if err == ErrNoRows {
// 			return Agent{}, ErrNotFound
// 		}
// 		return Agent{}, err
// 	}

// 	return agent, nil
// }

func (col *AgentCollection) GetByIdentityPass(identity string, password string) (Agent, error) {

	agent := Agent{}
	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where identity= '%s'", identity))
	if err != nil {
		if err == ErrNoRows {
			return Agent{}, ErrNotFound
		}
	}

	if agent.Password != GenerateHash(password, secretKey) {

		return Agent{}, ErrNotFound
	}

	if agent.Identity != "" {
		return agent, nil
	}

	return Agent{}, err
}

func (col *AgentCollection) ExistsByIdentity(identity string) (bool, error) {
	var agent Agent
	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where identity = '%s'", identity))

	if err != nil {
		if err == ErrNoRows {

			return false, nil
		} else {
			return false, err

		}
	}
	if agent.Identity != "" {
		return true, nil
	}
	return false, err

}

func (col *AgentCollection) FBLogin(agent Agent, accessToken string) (AgentToken, error) {
	user, err := GetFbUser(accessToken)
	if err != nil {
		return AgentToken{}, err
	}

	email, _ := user.GetString("email")
	fid, _ := user.GetString("id")

	exists, err := col.ExistsByIdentity(email)
	if err != nil {
		return AgentToken{}, err
	}
	var agentId int
	if !exists {
		agent := Agent{}
		agent.Identity = email
		agent.Is_Active = true
		agent.FId = fid

		result, err := col.Collection.Create(SystemToken, trans, &agent)
		if err != nil {
			return AgentToken{}, err

		}
		agentId = result.(*Agent).Id
	} else {
		result, err := col.ListQuery("identity="+email, "")
		if err != nil {
			return AgentToken{}, err
		}
		arr := *result.(*[]Agent)

		lagent := arr[0]
		//update agent if has no fid
		if lagent.FId == "" {
			lagent.FId = fid
			err = col.Update(SystemToken, lagent.Id, lagent)
			if err != nil {
				return AgentToken{}, err
			}
		}
		agentId = lagent.Id
	}

	agent.Id = agentId
	agent.Identity = email
	agent.FId = fid
	return col.Login(agent, "Facebook")

}

func (col *AgentCollection) ActiveUserByToken(act_token string) error {
	activationToken, err := AgentTokenColl.GetToken(act_token, "activation")
	if err != nil {
		return err
	}

	result, err := AgentColl.Get(SystemToken, activationToken.Agent_Id, "")
	if err != nil {
		return err
	}

	agent := result.(*Agent)
	agent.Is_Active = true
	err = AgentColl.Update(SystemToken, agent.Id, agent)
	if err != nil {
		return err
	}

	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='activation' and agent_id=%d ", agent.Id))
	if err != nil {
		return err
	}
	return nil

}

func (col *AgentCollection) RequestPasswordToken(identity string) (AgentToken, error) {
	result, err := AgentColl.ListQuery("identity="+identity, "")
	agentL := *result.(*[]Agent)
	if err != nil {
		return AgentToken{}, err
	}
	if len(agentL) == 0 {
		return AgentToken{}, ErrNotFound
	}
	agent := agentL[0]

	// agentToken := AgentToken{}
	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='password' and agent_id=%d ", agent.Id))
	if err != nil {
		return AgentToken{}, err
	}
	agentToken, err := AgentTokenColl.CreateToken(NewToken(agent.Id, "password"))
	if err != nil {
		return AgentToken{}, err
	}
	return agentToken, nil

}

func (col *AgentCollection) Login(agent Agent, url string) (AgentToken, error) {
	validationError := ValidateStruct(agent)

	if validationError != nil {
		return AgentToken{}, validationError
	}
	var ragent Agent
	var err error
	if url == "" {
		ragent, err = col.GetByIdentityPass(agent.Identity, agent.Password)

		if err != nil {
			return AgentToken{}, err
		}
	} else if url == "Facebook" {
		ragent, err = col.GetByFId(agent.Identity, agent.FId)

		if err != nil {
			return AgentToken{}, err
		}
	} else if url == "Google" {
		ragent, err = col.GetByGId(agent.Identity, agent.GId)

		if err != nil {
			return AgentToken{}, err
		}
	}

	if ragent.Is_Active == false {
		_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='activation' and agent_id=%d ", ragent.Id))
		if err != nil {
			return AgentToken{}, err
		}

		_, err := AgentTokenColl.CreateToken(NewToken(ragent.Id, "activation"))

		if err != nil {
			return AgentToken{}, err

		}

		return AgentToken{}, ErrAgentNotActive

	}

	//create token
	newToken := NewToken(ragent.Id, "login")

	//logout before tokens
	strQ := "update agent_tokens set is_expired=true where device_id in(select id from devices where uuid='" + agent.Device.Uuid + "') and type='login'"
	_, err1 := col.DataSrc.DB.Exec(strQ)

	if err1 != nil {
		application.Logger.Error(err1.Error())
	}
	// var result []Device
	var rdevice *Device

	results, err := DeviceColl.ListQuery("uuid="+agent.Device.Uuid+"&ip="+agent.Device.Ip, "")

	result1 := *results.(*[]Device)
	if len(result1) > 0 {

		rdevice = &result1[len(result1)-1]
		rdevice.CM_Id = agent.Device.CM_Id
		rdevice.Version = agent.Device.Version

		newToken.Device_Id = rdevice.Id

		err := DeviceColl.Update(SystemToken, rdevice.Id, rdevice)

		// TODO fix this for diffrent dbs right now it has two problem of being messy (string compare) and just worrks with postgresql
		if err != nil && !strings.Contains(err.Error(), Messages["DuplicateIndex"]) {

			return AgentToken{}, err
		}
	} else {

		result, err := DeviceColl.Create(SystemToken, nil, &agent.Device)
		if err != nil && !strings.Contains(err.Error(), Messages["DuplicateIndex"]) {

			return AgentToken{}, err
		}

		rdevice = result.(*Device)

		newToken.Device_Id = rdevice.Id
	}
	agentToken, err := AgentTokenColl.CreateToken(newToken)
	if err != nil {
		return AgentToken{}, err
	}

	return agentToken, nil
}
