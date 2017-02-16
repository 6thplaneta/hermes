package hermes

import (
	"database/sql"
	"fmt"
)

type Permission struct {
	Id    int    `json:"id" hermes:"dbspace:permissions"`
	Title string `json:"title" hermes:"searchable"`
}

type Role_Permission struct {
	Id            int `json:"id" hermes:"dbspace:roles_permissions,midtable,ui-html:None"`
	Role_Id       int `json:"role_id"`
	Permission_Id int `json:"permission_id"`
}

type Role struct {
	Id          int          `json:"id" hermes:"dbspace:roles"`
	Title       string       `json:"title" hermes:"searchable,editable"`
	Description string       `json:"description" hermes:"searchable,editable"`
	Permissions []Permission `json:"permissions" db:"-" hermes:"many2many:Role_Permission"`
	Agents      []Agent      `json:"agents" db:"-" hermes:"many2many:Role_Agent"`
}

type Role_Agent struct {
	Id       int `json:"id" hermes:"dbspace:roles_agents,midtable,ui-html:None"`
	Role_Id  int `json:"role_id"`
	Agent_Id int `json:"agent_id"`
}

func AddRole(role string) (int, error) {
	rlid := []int{}
	err := roleColl.GetDataSrc().DB.Select(&rlid, "select id from roles where title='"+role+"'")
	if err != nil {
		return 0, err
	}
	if len(rlid) == 0 {
		rl := &Role{Title: role}
		rlmatr, errcr := roleColl.Create(SystemToken, nil, rl)
		rlmat := rlmatr.(*Role)
		return rlmat.Id, errcr
	} else {
		return rlid[0], nil
	}

}

func AddRolePermission(roleid int, perms ...string) error {
	_, err := roleColl.Get(SystemToken, roleid, "")
	if err != nil {
		return err
	}
	ids := []int{}
	prmlist := ""
	for _, prm := range perms {
		prmlist += "'" + prm + "',"
	}
	prmlist = prmlist[:len(prmlist)-1]
	errselect := roleColl.GetDataSrc().DB.Select(&ids, "select id from permissions where title in ("+prmlist+");")
	if errselect != nil {
		return errselect
	}
	for _, id := range ids {
		dbid := 0
		err := roleColl.GetDataSrc().DB.Get(&dbid, fmt.Sprintf("select id from roles_permissions where role_id=%d and permission_id=%d ", roleid, id))
		if err == sql.ErrNoRows {
			rl_prm := &Role_Permission{Role_Id: roleid, Permission_Id: id}
			_, errCr := rolePermColl.Create(SystemToken, nil, rl_prm)
			if errCr != nil {
				return errCr
			}
		} else if err != nil {
			return err
		}

	}
	return nil

}
