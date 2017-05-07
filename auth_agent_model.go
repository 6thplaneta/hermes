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
	Id        int    `json:"id" hermes:"dbspace:agents"`
	Identity  string `json:"identity" validate:"required"`
	Password  string `json:"password,omitempty"`
	Is_Active bool   `json:"is_active" hermes:"editable"`
	// Is_Super   bool   `json:"is_super" hermes:"editable"`
	FId        string `json:"fid"`
	GId        string `json:"gid"`
	Roles      []Role `json:"roles,omitempty" db:"-" hermes:"many2many:Role_Agent"`
	Device     Device `db:"-" json:"device"`
	Is_Deleted bool   `json:"is_deleted" hermes:"index,editable"`
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
	required_password := true
	//sign up by facebook or gmail does not require password field for authentication
	//and is active (doesn't require to use activation code to activate the agent)
	if agent.FId != "" || agent.GId != "" {
		//
		required_password = false
		agent.Is_Active = true
	}

	newAgent = *agent
	if required_password == true {
		//check password required and format
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

	result, err := AgentColl.Collection.Create(token, nil, &newAgent)

	if err != nil {
		return &Agent{}, err
	}
	obj := result.(*Agent)

	if required_password == true {
		//create activation token for activating the user
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
		//check if the access token owner is the same as the object owner
		agent, err := AgentColl.GetByLoginToken(token)
		if err != nil {
			return err
		}

		// give access to the action if the token is the owner of resource
		//SystemToken has access to do all actions in the system
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

// get agent of the access token
func (col *AgentCollection) GetByToken(token string) (interface{}, error) {
	result := Agent{}
	agentToken, err := AgentTokenColl.GetToken(token, "login")
	if err != nil {
		return nil, err
	}
	err = col.DataSrc.DB.Get(&result, fmt.Sprintf("select * from agents where id= %d", agentToken.Agent_Id))
	return result, err
}

// get agent of the login token
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

//change password by the token code sent to email
func (col *AgentCollection) UpdatePasswordByToken(token string, newPassword string, identity string) error {
	identity = strings.ToLower(identity)
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

//change password by old password
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

//get agent by facebook id
//fid=facebook id
func (col *AgentCollection) GetByFId(identity string, fid string) (Agent, error) {

	agent := Agent{}
	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where lower(identity)= lower('%s') and fid= '%s' and is_deleted=false ", identity, fid))
	if err != nil {
		if err == ErrNoRows {
			return Agent{}, ErrNotFound
		}
		return Agent{}, err
	}

	return agent, nil
}

//get agent by gmail id
//gid=gmail id
func (col *AgentCollection) GetByGId(identity string, gid string) (Agent, error) {

	agent := Agent{}
	q := fmt.Sprintf("select * from agents where lower(identity)= lower('%s') and gid= '%s' and is_deleted=false ", identity, gid)
	err := col.DataSrc.DB.Get(&agent, q)

	if err != nil {
		if err == ErrNoRows {
			return Agent{}, ErrNotFound
		}
		return Agent{}, err
	}

	return agent, nil
}

//search in db for an agent with passed identity and password
func (col *AgentCollection) GetByIdentityPass(identity string, password string) (Agent, error) {

	agent := Agent{}
	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where lower(identity)= lower('%s')", identity))
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

//search in db for an agent with passed identity
func (col *AgentCollection) ExistsByIdentity(identity string) (bool, error) {
	var agent Agent
	err := col.DataSrc.DB.Get(&agent, fmt.Sprintf("select * from agents where lower(identity) = lower('%s') and is_deleted=false ", identity))

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

//login by facebook account
func (col *AgentCollection) FBLogin(agent Agent, accessToken string) (AgentToken, error) {

	//get the user info from facebook
	user, err := GetFbUser(accessToken)
	if err != nil {
		return AgentToken{}, err
	}

	email, _ := user.GetString("email")
	fid, _ := user.GetString("id")

	//check if the user exists in db with email address of facebook account
	exists, err := col.ExistsByIdentity(email)
	if err != nil {
		return AgentToken{}, err
	}
	var agentId int
	if !exists {

		//creat the agent if not exists
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
		//fill fid if the agent exists before and hasn't facebook id
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
	//get token info of activation token
	activationToken, err := AgentTokenColl.GetToken(act_token, "activation")
	if err != nil {
		return err
	}
	//get agent info by agent_id
	result, err := AgentColl.Get(SystemToken, activationToken.Agent_Id, "")
	if err != nil {
		return err
	}

	agent := result.(*Agent)
	//activate the agent and update the agent
	agent.Is_Active = true
	err = AgentColl.Update(SystemToken, agent.Id, agent)
	if err != nil {
		return err
	}

	//make expired all activation tokens of the agent (after expiration activation token works no longer)
	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='activation' and agent_id=%d ", agent.Id))
	if err != nil {
		return err
	}
	return nil

}

//forget password
//the function takes identity as input and creats no reset password token
func (col *AgentCollection) RequestPasswordToken(identity string) (AgentToken, error) {
	identity = strings.ToLower(identity)
	var agents []Agent
	//check whether the agent with this identity exists or not
	err := col.DataSrc.DB.Select(&agents, "select * from agents where identity='"+identity+"' and is_deleted=false and password<>''")
	if err != nil {
		return AgentToken{}, err
	}

	//if agent not exists return error
	if len(agents) == 0 {
		return AgentToken{}, ErrNotFound
	}
	agent := agents[0]

	//expire all previous reset password tokens of the agent
	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='password' and agent_id=%d ", agent.Id))
	if err != nil {
		return AgentToken{}, err
	}
	//create and return new reset password token
	agentToken, err := AgentTokenColl.CreateToken(NewToken(agent.Id, "password"))

	if err != nil {
		return AgentToken{}, err
	}
	return agentToken, nil

}

//url is the auth type(common type of auth ="", auth by facebook="Facebook" auth by gmail="Google")
func (col *AgentCollection) Login(agent Agent, url string) (AgentToken, error) {
	//check validation rules
	validationError := ValidateStruct(agent)
	if validationError != nil {
		return AgentToken{}, validationError
	}
	var ragent Agent
	var err error
	if url == "" {
		//get agent by identity and password
		ragent, err = col.GetByIdentityPass(agent.Identity, agent.Password)

		if err != nil {
			return AgentToken{}, err
		}
	} else if url == "Facebook" {
		//get agent by facebook id
		ragent, err = col.GetByFId(agent.Identity, agent.FId)
		if err != nil {
			return AgentToken{}, err
		}
	} else if url == "Google" {
		//get agent by gmail id
		ragent, err = col.GetByGId(agent.Identity, agent.GId)

		if err != nil {
			return AgentToken{}, err
		}
	}

	//if the agent is deactive and the user is trying to login with a deactive account ,
	//create new activation token and return the token
	if ragent.Is_Active == false {

		_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='activation' and agent_id=%d ", ragent.Id))
		if err != nil {
			return AgentToken{}, err
		}

		act_tok, err := AgentTokenColl.CreateToken(NewToken(ragent.Id, "activation"))

		if err != nil {
			return AgentToken{}, err

		}

		return act_tok, ErrAgentNotActive

	}

	//create new login token
	newToken := NewToken(ragent.Id, "login")

	//make expired previous tokens that the user logged in with the same device(logout from the the same device)
	strQ := "update agent_tokens set is_expired=true where device_id in(select id from devices where uuid='" + agent.Device.Uuid + "') and type='login'"
	_, err1 := col.DataSrc.DB.Exec(strQ)

	if err1 != nil {
		application.Logger.Error(err1.Error())
	}
	var rdevice *Device

	//	check if the user has logged in with the uuid and ip previosuly
	results, err := DeviceColl.ListQuery("uuid="+agent.Device.Uuid+"&ip="+agent.Device.Ip, "")

	result1 := *results.(*[]Device)
	//if the user logged in before, update the device information. Otherwise, create new device
	if len(result1) > 0 {

		rdevice = &result1[len(result1)-1]
		rdevice.CM_Id = agent.Device.CM_Id
		rdevice.Version = agent.Device.Version

		newToken.Device_Id = rdevice.Id

		err := DeviceColl.Update(SystemToken, rdevice.Id, rdevice)

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
	//create the successful login token and return it
	agentToken, err := AgentTokenColl.CreateToken(newToken)
	if err != nil {
		return AgentToken{}, err
	}

	return agentToken, nil
}
