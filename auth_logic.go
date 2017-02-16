package hermes

import (
	"strconv"
)

func (ag *Agent) caculatePermissions() ([]string, error) {
	var query string = "SELECT title FROM permissions where permissions.id in (select permission_id from roles_permissions where role_id in (select role_id from roles_agents where agent_id = " + strconv.Itoa(ag.Id) + ")) ;"
	var res []string = make([]string, 0)
	err := AgentColl.DataSrc.DB.Select(&res, query)
	if err != nil {
		return nil, err
	}
	return res, nil

}

func (ag *Agent) hasPermission(perm string) (bool, error) {
	perms, err := ag.caculatePermissions()
	if err != nil {
		return false, err
	}
	return strInArr(perm, perms), nil
}

var permCache map[string][]string
var authEnabled bool = true
var permsCacheEnabled bool = true
var PermissionCacheLength int = 1000

func DisablePermissionCache() {
	permsCacheEnabled = false
}

func evictPermissionCache() {
	permCache = make(map[string][]string, PermissionCacheLength)
}

func checkTokenPermDB(token, perm string) (bool, error) {
	ag, err := AgentColl.GetByLoginToken(token)
	if err != nil {
		return false, err
	}
	allPerms, err := ag.caculatePermissions()

	if err != nil {
		return false, err
	}
	if permsCacheEnabled && len(permCache) < PermissionCacheLength {
		permCache[token] = allPerms
	}
	return strInArr(perm, allPerms), nil

}

func checkTokenPermission(token, perm string) (bool, error) {
	if permsCacheEnabled {
		cachedPrms, hasToken := permCache[token]
		if hasToken {
			return strInArr(perm, cachedPrms), nil
		} else {
			return checkTokenPermDB(token, perm)
		}
	} else {
		return checkTokenPermDB(token, perm)

	}
}

func DisableAuth() {
	authEnabled = false
}

func RegisterPermissions(titles []string) {
	var prm *Permission
	for _, title := range titles {
		if title == "" {
			continue
		}
		objs, err := permColl.ListQuery("title="+title, "")
		if err != nil {
			return
		}
		exPrms := objs.(*[]Permission)
		if len(*exPrms) == 0 {
			prm = new(Permission)
			prm.Title = title
			_, err := permColl.Create(SystemToken, nil, prm)
			if err != nil {
			}

		}
	}
}

// change sysytem with init produced variable
func Authorize(token, permis string, id int, action string, chkfunc func(string, int, string) (bool, error)) bool {
	if token == SystemToken || !authEnabled || permis == "" {
		return true
	}

	//todo mahsa think about better way
	if token == "" {
		return false
	}

	hasperm, err := checkTokenPermission(token, permis)
	if err != nil {
		return false
	}
	if hasperm {
		return true
	} else {
		// check for own auth
		if id == 0 {
			return false
		}
		if chkfunc != nil {
			chk, errc := chkfunc(token, id, action)
			if errc != nil {
				return false
			}
			return chk
		}
		// no chk func provided
		return false

	}
}
